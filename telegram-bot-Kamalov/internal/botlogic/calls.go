
package botlogic

import (
  "fmt"
  "strconv"
  "strings"

  "telegram-module/internal/api"
  "telegram-module/internal/config"
)

func (h *Handler) buildCall(spec config.CommandSpec, args []string) (api.Call, string) {
  // validate args count: allow last arg to contain spaces if user used quotes; ParseArgs does.
  if len(args) < len(spec.Args) {
    return api.Call{}, "Недостаточно параметров."
  }
  // map arg name -> value
  m := map[string]string{}
  for i, name := range spec.Args {
    m[name] = args[i]
  }

  // substitute path placeholders {x}
  path := spec.Path
  for k, v := range m {
    path = strings.ReplaceAll(path, "{"+k+"}", v)
  }

  var body any = nil
  // bodies by known paths (best-effort шаблоны)
  switch spec.Name {
  case "user_fullname_set":
    body = map[string]any{"fullName": m["fullName"]}
  case "user_roles_set":
    roles := splitCSV(m["rolesCsv"])
    body = map[string]any{"roles": roles}
  case "user_block_set":
    b, err := parseBool(m["blocked"])
    if err != nil { return api.Call{}, "Параметр blocked должен быть on/off/true/false/1/0" }
    body = map[string]any{"blocked": b}

  case "course_update":
    body = map[string]any{"name": m["name"], "description": m["description"]}
  case "course_test_active":
    b, err := parseBool(m["active"])
    if err != nil { return api.Call{}, "Параметр active должен быть on/off/true/false/1/0" }
    body = map[string]any{"active": b}
  case "course_test_add":
    body = map[string]any{"title": m["title"]}
  case "course_student_add":
    body = map[string]any{"userId": m["user_id"]}
  case "course_create":
    body = map[string]any{"name": m["name"], "description": m["description"], "teacherId": m["teacher_id"]}

  case "question_create":
    opts := splitCSV(m["optionsCsv"])
    idx, err := strconv.Atoi(m["correctIndex"])
    if err != nil { return api.Call{}, "correctIndex должен быть числом" }
    body = map[string]any{"title": m["title"], "text": m["text"], "options": opts, "correctIndex": idx}
  case "question_update":
    opts := splitCSV(m["optionsCsv"])
    idx, err := strconv.Atoi(m["correctIndex"])
    if err != nil { return api.Call{}, "correctIndex должен быть числом" }
    body = map[string]any{"title": m["title"], "text": m["text"], "options": opts, "correctIndex": idx}

  case "test_question_add":
    body = map[string]any{"questionId": m["question_id"]}
  case "test_question_order":
    ids := splitCSV(m["questionIdsCsv"])
    body = map[string]any{"questionIds": ids}
  case "answer_update":
    idx, err := strconv.Atoi(m["index"])
    if err != nil { return api.Call{}, "index должен быть числом" }
    body = map[string]any{"index": idx}
  }

  // Some commands may have extra args; ignore for now.

  // Validate all placeholders substituted
  if strings.Contains(path, "{") {
    return api.Call{}, fmt.Sprintf("Не хватает параметров для URL (%s).", path)
  }

  return api.Call{Method: spec.Method, Path: path, Body: body}, ""
}

func splitCSV(s string) []string {
  s = strings.TrimSpace(s)
  if s == "" { return nil }
  parts := strings.Split(s, ",")
  out := make([]string, 0, len(parts))
  for _, p := range parts {
    p = strings.TrimSpace(p)
    if p != "" { out = append(out, p) }
  }
  return out
}

func parseBool(s string) (bool, error) {
  s = strings.ToLower(strings.TrimSpace(s))
  switch s {
  case "1", "true", "yes", "on", "да":
    return true, nil
  case "0", "false", "no", "off", "нет":
    return false, nil
  default:
    return false, fmt.Errorf("bad bool")
  }
}
