
package main

import (
  "bytes"
  "encoding/json"
  "log"
  "net/http"
  "os"
  "time"
  "strconv"

  tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

  "telegram-module/internal/model"
)

func main() {
  token := os.Getenv("TELEGRAM_TOKEN")
  if token == "" {
    log.Fatal("TELEGRAM_TOKEN не задан")
  }
  botLogicURL := getenv("BOT_LOGIC_URL", "http://nginx:8080")
  pollTimeout := getenvInt("POLL_TIMEOUT", 60)

  cronLoginEvery := getenvDur("CRON_CHECK_LOGIN_EVERY", 10*time.Second)
  cronNotifEvery := getenvDur("CRON_CHECK_NOTIF_EVERY", 20*time.Second)

  bot, err := tgbotapi.NewBotAPI(token)
  if err != nil { log.Fatal(err) }
  bot.Debug = getenv("BOT_DEBUG","") == "1"
  log.Printf("Telegram Client запущен как @%s", bot.Self.UserName)

  u := tgbotapi.NewUpdate(0)
  u.Timeout = pollTimeout
  updates := bot.GetUpdatesChan(u)

  httpClient := &http.Client{Timeout: 15*time.Second}

  // cron: check login
  go func(){
    t := time.NewTicker(cronLoginEvery)
    for range t.C {
      callCron(httpClient, botLogicURL + "/cron/check-login", bot)
    }
  }()

  // cron: notifications
  go func(){
    t := time.NewTicker(cronNotifEvery)
    for range t.C {
      callCron(httpClient, botLogicURL + "/cron/check-notifications", bot)
    }
  }()

  for upd := range updates {
    // each update in goroutine
    go func(u tgbotapi.Update) {
      resp, err := forwardUpdate(httpClient, botLogicURL + "/update", &u)
      if err != nil {
        log.Println("forwardUpdate error:", err)
        return
      }
      sendResponse(bot, resp)
    }(upd)
  }
}

func forwardUpdate(c *http.Client, url string, upd *tgbotapi.Update) (*model.Response, error) {
  b, _ := json.Marshal(upd)
  req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
  req.Header.Set("Content-Type", "application/json")
  resp, err := c.Do(req)
  if err != nil { return nil, err }
  defer resp.Body.Close()
  var out model.Response
  if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
    return nil, err
  }
  return &out, nil
}

func callCron(c *http.Client, url string, bot *tgbotapi.BotAPI) {
  req, _ := http.NewRequest("GET", url, nil)
  resp, err := c.Do(req)
  if err != nil { return }
  defer resp.Body.Close()
  var out model.Response
  if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return }
  sendResponse(bot, &out)
}

func sendResponse(bot *tgbotapi.BotAPI, resp *model.Response) {
  if resp == nil { return }

  if resp.CallbackAnswer != nil {
    cb := tgbotapi.NewCallback(resp.CallbackAnswer.CallbackQueryID, resp.CallbackAnswer.Text)
    cb.ShowAlert = resp.CallbackAnswer.ShowAlert
    if _, err := bot.Request(cb); err != nil {
      // ignore
    }
  }

  for _, m := range resp.Messages {
    msg := tgbotapi.NewMessage(m.ChatID, m.Text)
    if m.DisablePreview {
      msg.DisableWebPagePreview = true
    }
    if m.ParseMode != "" {
      msg.ParseMode = m.ParseMode
    }
    if m.Keyboard != nil {
      msg.ReplyMarkup = toTGKeyboard(m.Keyboard)
    }
    if _, err := bot.Send(msg); err != nil {
      // ignore, but log
      log.Println("send message error:", err)
    }
  }
}

func toTGKeyboard(k *model.InlineKeyboard) tgbotapi.InlineKeyboardMarkup {
  rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(k.Rows))
  for _, r := range k.Rows {
    row := make([]tgbotapi.InlineKeyboardButton, 0, len(r))
    for _, b := range r {
      row = append(row, tgbotapi.NewInlineKeyboardButtonData(b.Text, b.Data))
    }
    rows = append(rows, row)
  }
  return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func getenv(k, d string) string {
  if v := os.Getenv(k); v != "" { return v }
  return d
}

func getenvInt(k string, d int) int {
  if v := os.Getenv(k); v != "" {
    if n, err := strconv.Atoi(v); err == nil { return n }
  }
  return d
}

func getenvDur(k string, d time.Duration) time.Duration {
  if v := os.Getenv(k); v != "" {
    if dur, err := time.ParseDuration(v); err == nil { return dur }
  }
  return d
}
