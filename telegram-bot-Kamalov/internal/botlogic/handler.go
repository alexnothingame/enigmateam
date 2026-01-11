
package botlogic

import (
  "encoding/json"
  "fmt"
  "net/http"
  "strings"
  "time"

  tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

  "telegram-module/internal/api"
  "telegram-module/internal/config"
  "telegram-module/internal/model"
  "telegram-module/internal/store"
)

type Deps struct {
  Store *store.Store
  Registry *config.Registry
  Auth *api.AuthClient
  Main *api.MainClient
  SessionTTL time.Duration
  NotificationPath string
  NotificationAckPath string
}

type Handler struct {
  st *store.Store
  reg *config.Registry
  auth *api.AuthClient
  main *api.MainClient
  ttl time.Duration
  notifPath string
  notifAckPath string
}

func NewHandler(d Deps) *Handler {
  return &Handler{
    st: d.Store,
    reg: d.Registry,
    auth: d.Auth,
    main: d.Main,
    ttl: d.SessionTTL,
    notifPath: d.NotificationPath,
    notifAckPath: d.NotificationAckPath,
  }
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
  var upd tgbotapi.Update
  if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
    http.Error(w, "bad json", 400)
    return
  }

  resp := h.processUpdate(&upd)

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CronCheckLogin(w http.ResponseWriter, r *http.Request) {
  sessions, err := h.st.ScanSessions()
  if err != nil {
    http.Error(w, err.Error(), 500)
    return
  }
  out := model.Response{}
  for _, s := range sessions {
    if s.State != store.StateAnonymous { continue }
    if s.LoginToken == "" { continue }
    if s.WaitingCode { continue } // code flow checked by user input
    // oauth check
    res, err := h.auth.CheckLogin(s.LoginToken)
    if err != nil {
      continue
    }
    switch res.Status {
    case "pending":
      continue
    case "denied":
      // Сбрасываем попытку логина
      s.LoginToken = ""
      s.LoginType = ""
      s.WaitingCode = false
      _ = h.st.Set(s.ChatID, s, h.ttl)
      out.Messages = append(out.Messages, model.OutMessage{
        ChatID: s.ChatID,
        Text: "❌ Авторизация отклонена или истекла. Нажмите /login, чтобы попробовать снова.",
      })
    case "success":
      s.State = store.StateAuthorized
      s.AccessToken = res.AccessToken
      s.RefreshToken = res.RefreshToken
      s.AccessExpUnix = res.AccessExpUnix
      s.LoginToken = ""
      s.LoginType = ""
      s.WaitingCode = false
      _ = h.st.Set(s.ChatID, s, h.ttl)
      out.Messages = append(out.Messages, model.OutMessage{
        ChatID: s.ChatID,
        Text: "✅ Авторизация завершена! Откройте /menu и пользуйтесь командами.",
      })
    }
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(out)
}

func (h *Handler) CronCheckNotifications(w http.ResponseWriter, r *http.Request) {
  sessions, err := h.st.ScanSessions()
  if err != nil {
    http.Error(w, err.Error(), 500)
    return
  }
  out := model.Response{}
  for _, s := range sessions {
    if s.State != store.StateAuthorized { continue }
    if ok := h.ensureFreshToken(s); !ok {
      out.Messages = append(out.Messages, model.OutMessage{
        ChatID: s.ChatID,
        Text: "ℹ️ Сессия истекла. Нажмите /login чтобы войти снова.",
      })
      continue
    }
    // Получаем уведомления из главного модуля
    res, err := h.main.Do(s.AccessToken, api.Call{Method:"GET", Path: h.notifPath})
    if err != nil { continue }
    if res.StatusCode != 200 { continue }
    bodyText := api.PrettyBody(res.ContentType, res.BodyBytes)
    if bodyText == "(пустой ответ)" || strings.TrimSpace(bodyText) == "" {
      continue
    }
    // ACK (опционально)
    _, _ = h.main.Do(s.AccessToken, api.Call{Method:"POST", Path: h.notifAckPath, Body: map[string]any{"ack": true}})
    out.Messages = append(out.Messages, model.OutMessage{
      ChatID: s.ChatID,
      Text: "🔔 Уведомления:\n" + bodyText,
    })
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(out)
}

func (h *Handler) processUpdate(upd *tgbotapi.Update) model.Response {
  // Determine chat ID
  var chatID int64
  if upd.Message != nil {
    chatID = upd.Message.Chat.ID
  } else if upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil {
    chatID = upd.CallbackQuery.Message.Chat.ID
  } else {
    return model.Response{}
  }

  sess, _ := h.st.Get(chatID)
  if sess == nil {
    sess = &store.Session{ChatID: chatID, State: store.StateAnonymous}
    _ = h.st.Set(chatID, sess, h.ttl)
  }

  // Callback
  if upd.CallbackQuery != nil {
    return h.handleCallback(sess, upd.CallbackQuery)
  }

  // Message
  msg := upd.Message
  text := strings.TrimSpace(msg.Text)

  if msg.IsCommand() {
    cmd := "/" + msg.Command()
    argsLine := strings.TrimSpace(msg.CommandArguments())
    args := ParseArgs(argsLine)
    return h.handleCommand(sess, cmd, args)
  }

  // non-command input
  if sess.State == store.StateAnonymous && sess.WaitingCode && sess.LoginType == "code" {
    // treat as code
    return h.handleCodeInput(sess, text)
  }

  // If user is authorized and types something not command: hint
  if sess.State == store.StateAuthorized {
    return model.Response{
      Messages: []model.OutMessage{{
        ChatID: chatID,
        Text: "ℹ️ Используйте /menu или /help. Команды начинаются со слеша (/).",
      }},
    }
  }

  // default: show login prompt
  return h.replyLoginPrompt(chatID, "🔐 Для работы нужна авторизация.")
}

func (h *Handler) handleCommand(sess *store.Session, cmd string, args []string) model.Response {
  chatID := sess.ChatID
  switch cmd {
  case "/start":
    if sess.State == store.StateAuthorized {
      return model.Response{Messages: []model.OutMessage{{
        ChatID: chatID,
        Text: "👋 Привет! Ты уже авторизован. Открой /menu.",
      }}}
    }
    // ensure session exists
    sess.State = store.StateAnonymous
    _ = h.st.Set(chatID, sess, h.ttl)
    return h.replyLoginPrompt(chatID, "👋 Привет! Выбери способ входа.")
  case "/help":
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: h.helpText(),
    }}}
  case "/menu":
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: "📋 Меню. Выбери раздел:",
      Keyboard: h.menuKeyboard(),
    }}}
  case "/login":
    if len(args) == 0 {
      return h.replyLoginPrompt(chatID, "Выбери способ авторизации:")
    }
    typ := strings.ToLower(args[0])
    if typ != "github" && typ != "yandex" && typ != "code" {
      return model.Response{Messages: []model.OutMessage{{
        ChatID: chatID,
        Text: "⚠️ Неверный тип. Используй: /login github | /login yandex | /login code",
      }}}
    }
    // start login via auth service
    start, err := h.auth.StartLogin(chatID, typ)
    if err != nil {
      return model.Response{Messages: []model.OutMessage{{
        ChatID: chatID,
        Text: "⚠️ Не удалось начать авторизацию. Проверь, доступен ли сервис авторизации.",
      }}}
    }
    sess.State = store.StateAnonymous
    sess.LoginType = typ
    sess.LoginToken = start.LoginToken
    sess.WaitingCode = (typ == "code")
    _ = h.st.Set(chatID, sess, h.ttl)

    if typ == "code" {
      hint := start.Hint
      if hint == "" {
        hint = "Открой Web-клиент и возьми код входа, затем отправь его сюда."
      }
      return model.Response{Messages: []model.OutMessage{
        {ChatID: chatID, Text: "📟 Авторизация по коду."},
        {ChatID: chatID, Text: hint},
        {ChatID: chatID, Text: "Отправь сюда код одним сообщением. Для отмены: /cancel"},
      }}
    }

    url := start.URL
    if url == "" {
      url = fmt.Sprintf("%s (URL не пришёл от auth)", h.auth.BaseURL)
    }
    return model.Response{Messages: []model.OutMessage{
      {ChatID: chatID, Text: "🌐 Авторизация через " + strings.Title(typ)},
      {ChatID: chatID, Text: "Перейди по ссылке и заверши вход:\n" + url, DisablePreview: true},
      {ChatID: chatID, Text: "После входа бот сам подтвердит авторизацию (периодическая проверка)."},
    }}
  case "/logout":
    // optional: all=true
    sess.State = store.StateAnonymous
    sess.AccessToken = ""
    sess.RefreshToken = ""
    sess.AccessExpUnix = 0
    sess.LoginType = ""
    sess.LoginToken = ""
    sess.WaitingCode = false
    sess.ActiveAttemptID = ""
    sess.ActiveTestID = ""
    _ = h.st.Set(chatID, sess, h.ttl)
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: "🚪 Вы вышли из системы. Для входа: /login",
    }}}
  case "/cancel":
    if sess.State == store.StateAnonymous && (sess.LoginToken != "" || sess.WaitingCode) {
      sess.LoginType = ""
      sess.LoginToken = ""
      sess.WaitingCode = false
      _ = h.st.Set(chatID, sess, h.ttl)
      return model.Response{Messages: []model.OutMessage{{
        ChatID: chatID,
        Text: "⛔ Отменено. Можешь снова зайти через /login.",
      }}}
    }
    if sess.ActiveAttemptID != "" {
      sess.ActiveAttemptID = ""
      sess.ActiveTestID = ""
      _ = h.st.Set(chatID, sess, h.ttl)
      return model.Response{Messages: []model.OutMessage{{
        ChatID: chatID,
        Text: "⛔ Текущая попытка/тест сброшены (локально).",
      }}}
    }
    return model.Response{Messages: []model.OutMessage{{ChatID:chatID, Text:"Нечего отменять."}}}
  }

  // If it's a registry command
  spec, ok := h.reg.FindBySlash(cmd)
  if !ok {
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: "⚠️ Неизвестная команда. Открой /help",
    }}}
  }

  // auth required for all registry commands
  if sess.State != store.StateAuthorized {
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: "🔒 Для этой команды нужна авторизация. Нажми /login",
    }}}
  }

  // ensure token fresh
  if ok := h.ensureFreshToken(sess); !ok {
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: "ℹ️ Сессия истекла. Нажми /login, чтобы войти снова.",
    }}}
  }

  call, errText := h.buildCall(spec, args)
  if errText != "" {
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: errText + "\nПодсказка: " + h.usage(spec),
    }}}
  }

  result, err := h.main.Do(sess.AccessToken, call)
  if err != nil {
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: "⚠️ Ошибка вызова Главного модуля: " + err.Error(),
    }}}
  }

  pretty := api.PrettyBody(result.ContentType, result.BodyBytes)
  header := api.StatusToText(result.StatusCode)
  text := fmt.Sprintf("%s\nКоманда: %s\n\n%s", header, spec.Slash, pretty)

  // Telegram message length limit: trim
  if len(text) > 3800 {
    text = text[:3800] + "\n... (обрезано)"
  }

  return model.Response{Messages: []model.OutMessage{{ChatID:chatID, Text:text}}}
}

func (h *Handler) handleCodeInput(sess *store.Session, code string) model.Response {
  chatID := sess.ChatID
  res, err := h.auth.VerifyCode(sess.LoginToken, code)
  if err != nil {
    return model.Response{Messages: []model.OutMessage{{
      ChatID: chatID,
      Text: "❌ Код не принят. Проверь код и попробуй ещё раз. Для отмены: /cancel",
    }}}
  }
  sess.State = store.StateAuthorized
  sess.AccessToken = res.AccessToken
  sess.RefreshToken = res.RefreshToken
  sess.AccessExpUnix = res.AccessExpUnix
  sess.LoginType = ""
  sess.LoginToken = ""
  sess.WaitingCode = false
  _ = h.st.Set(chatID, sess, h.ttl)
  return model.Response{Messages: []model.OutMessage{
    {ChatID: chatID, Text:"✅ Авторизация успешна!"},
    {ChatID: chatID, Text:"Открой /menu и выбирай команды."},
  }}
}

func (h *Handler) handleCallback(sess *store.Session, q *tgbotapi.CallbackQuery) model.Response {
  chatID := sess.ChatID
  data := q.Data

  resp := model.Response{
    CallbackAnswer: &model.CallbackAnswer{CallbackQueryID: q.ID},
  }

  if strings.HasPrefix(data, "MENU:") {
    group := strings.TrimPrefix(data, "MENU:")
    // find group
    for _, g := range h.reg.Groups {
      if g.Prefix == group {
        resp.Messages = append(resp.Messages, model.OutMessage{
          ChatID: chatID,
          Text: "Раздел: " + g.Title + "\nВыбери команду (для команд с параметрами — используй слеш-команду вручную):",
          Keyboard: h.groupKeyboard(g),
        })
        return resp
      }
    }
    resp.Messages = append(resp.Messages, model.OutMessage{ChatID:chatID, Text:"Раздел не найден."})
    return resp
  }

  
  if strings.HasPrefix(data, "LOGIN:") {
    typ := strings.TrimPrefix(data, "LOGIN:")
    // эмулируем /login <тип>
    r := h.handleCommand(sess, "/login", []string{typ})
    r.CallbackAnswer = resp.CallbackAnswer
    return r
  }

if strings.HasPrefix(data, "RUN:") {
    slash := strings.TrimPrefix(data, "RUN:")
    spec, ok := h.reg.FindBySlash(slash)
    if !ok {
      resp.Messages = append(resp.Messages, model.OutMessage{ChatID:chatID, Text:"Команда не найдена."})
      return resp
    }
    if len(spec.Args) > 0 {
      resp.Messages = append(resp.Messages, model.OutMessage{ChatID:chatID, Text:"Эта команда требует параметры. Используй: " + h.usage(spec)})
      return resp
    }
    // execute with no args
    r := h.handleCommand(sess, slash, nil)
    // preserve callback answer
    r.CallbackAnswer = resp.CallbackAnswer
    return r
  }

  resp.Messages = append(resp.Messages, model.OutMessage{ChatID:chatID, Text:"Неизвестное действие кнопки."})
  return resp
}

func (h *Handler) replyLoginPrompt(chatID int64, title string) model.Response {
  return model.Response{Messages: []model.OutMessage{{
    ChatID: chatID,
    Text: title,
    Keyboard: &model.InlineKeyboard{Rows: [][]model.InlineButton{
      {{Text:"GitHub", Data:"LOGIN:github"}, {Text:"Yandex", Data:"LOGIN:yandex"}},
      {{Text:"Код", Data:"LOGIN:code"}},
    }},
  }}}
}

func (h *Handler) menuKeyboard() *model.InlineKeyboard {
  rows := [][]model.InlineButton{}
  // build rows 2 per row
  row := []model.InlineButton{}
  for _, g := range h.reg.Groups {
    row = append(row, model.InlineButton{Text: g.Title, Data: "MENU:" + g.Prefix})
    if len(row) == 2 {
      rows = append(rows, row)
      row = []model.InlineButton{}
    }
  }
  if len(row) > 0 { rows = append(rows, row) }
  return &model.InlineKeyboard{Rows: rows}
}

func (h *Handler) groupKeyboard(g config.GroupSpec) *model.InlineKeyboard {
  rows := [][]model.InlineButton{}
  for _, name := range g.Commands {
    spec, ok := h.reg.FindByName(name)
    if !ok { continue }
    // only no-arg commands as clickable
    if len(spec.Args) == 0 {
      rows = append(rows, []model.InlineButton{{Text: spec.Slash, Data: "RUN:" + spec.Slash}})
    } else {
      // show as info button that tells usage
      rows = append(rows, []model.InlineButton{{Text: spec.Slash + " …", Data: "RUN:" + spec.Slash}})
    }
  }
  return &model.InlineKeyboard{Rows: rows}
}

func (h *Handler) helpText() string {
  // Main commands + all from registry
  b := strings.Builder{}
  b.WriteString("🆘 Команды:\n")
  b.WriteString("/start — старт\n")
  b.WriteString("/login github|yandex|code — авторизация\n")
  b.WriteString("/logout — выход\n")
  b.WriteString("/menu — меню\n")
  b.WriteString("/cancel — отмена авторизации/сброс локального теста\n\n")
  b.WriteString("📌 Команды действий (из таблиц):\n")
  for _, c := range h.reg.Commands {
    if strings.HasPrefix(c.Slash, "/") && c.Slash != "/starttest" {
      b.WriteString(c.Slash)
      if len(c.Args) > 0 {
        b.WriteString(" ")
        b.WriteString(strings.Join(c.Args, " "))
      }
      b.WriteString(" — ")
      b.WriteString(c.Desc)
      b.WriteString("\n")
    }
  }
  return b.String()
}

func (h *Handler) usage(spec config.CommandSpec) string {
  if len(spec.Args) == 0 { return spec.Slash }
  return spec.Slash + " " + strings.Join(spec.Args, " ")
}

// ensureFreshToken checks expiration and refreshes access token if needed.
// returns true if still authorized.
func (h *Handler) ensureFreshToken(sess *store.Session) bool {
  if sess.State != store.StateAuthorized { return false }
  if sess.AccessToken == "" { return false }
  if sess.AccessExpUnix == 0 {
    return true // cannot validate; assume ok
  }
  if time.Now().Unix() < sess.AccessExpUnix-5 {
    return true
  }
  if sess.RefreshToken == "" { return false }
  ref, err := h.auth.Refresh(sess.RefreshToken)
  if err != nil {
    // expire session
    sess.State = store.StateAnonymous
    sess.AccessToken = ""
    sess.RefreshToken = ""
    sess.AccessExpUnix = 0
    sess.LoginToken = ""
    sess.LoginType = ""
    sess.WaitingCode = false
    _ = h.st.Set(sess.ChatID, sess, h.ttl)
    return false
  }
  sess.AccessToken = ref.AccessToken
  sess.RefreshToken = ref.RefreshToken
  sess.AccessExpUnix = ref.AccessExpUnix
  _ = h.st.Set(sess.ChatID, sess, h.ttl)
  return true
}
