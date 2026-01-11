
package api

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "strings"
  "time"
)

type MainClient struct {
  BaseURL string
  HTTP *http.Client
}

func NewMainClient(baseURL string) *MainClient {
  return &MainClient{
    BaseURL: strings.TrimRight(baseURL, "/"),
    HTTP: &http.Client{Timeout: 12 * time.Second},
  }
}

type Call struct {
  Method string
  Path string
  Body any // marshalled to json when not nil
}

type Result struct {
  StatusCode int
  BodyBytes []byte
  ContentType string
}

func (c *MainClient) Do(accessToken string, call Call) (*Result, error) {
  var body io.Reader
  if call.Body != nil {
    b, err := json.Marshal(call.Body)
    if err != nil { return nil, err }
    body = bytes.NewReader(b)
  }
  req, _ := http.NewRequest(call.Method, c.BaseURL + call.Path, body)
  if call.Body != nil {
    req.Header.Set("Content-Type", "application/json")
  }
  if accessToken != "" {
    req.Header.Set("Authorization", "Bearer " + accessToken)
  }
  resp, err := c.HTTP.Do(req)
  if err != nil { return nil, err }
  defer resp.Body.Close()
  bb, _ := io.ReadAll(resp.Body)
  return &Result{
    StatusCode: resp.StatusCode,
    BodyBytes: bb,
    ContentType: resp.Header.Get("Content-Type"),
  }, nil
}

func PrettyBody(contentType string, b []byte) string {
  // if json -> indent
  if strings.Contains(contentType, "application/json") {
    var anyVal any
    if err := json.Unmarshal(b, &anyVal); err == nil {
      pretty, _ := json.MarshalIndent(anyVal, "", "  ")
      return string(pretty)
    }
  }
  // fallback: as text (limit)
  s := string(b)
  if len(s) > 3500 {
    s = s[:3500] + "\n... (обрезано)"
  }
  if s == "" {
    s = "(пустой ответ)"
  }
  return s
}

func StatusToText(code int) string {
  switch code {
  case 200, 201: return "✅ Успешно"
  case 400: return "⚠️ Ошибка запроса (400)"
  case 401: return "🔒 Не авторизован (401)"
  case 403: return "⛔ Доступ запрещён (403)"
  case 404: return "❓ Не найдено (404)"
  case 409: return "⚠️ Конфликт (409)"
  case 418: return "🚫 Заблокирован (418)"
  default:
    return fmt.Sprintf("Ответ: %d", code)
  }
}
