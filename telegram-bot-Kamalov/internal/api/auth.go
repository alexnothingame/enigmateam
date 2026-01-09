
package api

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "time"
)

type AuthClient struct {
  BaseURL string
  HTTP *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
  return &AuthClient{
    BaseURL: baseURL,
    HTTP: &http.Client{Timeout: 8 * time.Second},
  }
}

// StartLogin asks Auth service to start login flow (oauth/code).
// Expected contract (can be adapted by your Auth module):
// POST {base}/telegram/login/start  body: {"chat_id":123,"type":"github"|"yandex"|"code"}
// response: {"login_token":"...","url":"https://..."}  (for oauth) OR {"login_token":"...","hint":"..."} (for code)
type StartLoginResponse struct {
  LoginToken string `json:"login_token"`
  URL string `json:"url,omitempty"`
  Hint string `json:"hint,omitempty"`
}

func (c *AuthClient) StartLogin(chatID int64, typ string) (*StartLoginResponse, error) {
  payload := map[string]any{"chat_id": chatID, "type": typ}
  b, _ := json.Marshal(payload)
  req, _ := http.NewRequest("POST", c.BaseURL + "/telegram/login/start", bytes.NewReader(b))
  req.Header.Set("Content-Type", "application/json")
  resp, err := c.HTTP.Do(req)
  if err != nil { return nil, err }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("auth start login: статус %d", resp.StatusCode)
  }
  var out StartLoginResponse
  if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
  return &out, nil
}

// CheckLogin checks whether oauth login was completed.
// GET {base}/telegram/login/check?login_token=...
// response: {"status":"pending"|"denied"|"success", "access_token":"", "refresh_token":"", "access_exp_unix":123}
type CheckLoginResponse struct {
  Status string `json:"status"`
  AccessToken string `json:"access_token,omitempty"`
  RefreshToken string `json:"refresh_token,omitempty"`
  AccessExpUnix int64 `json:"access_exp_unix,omitempty"`
  Reason string `json:"reason,omitempty"`
}

func (c *AuthClient) CheckLogin(loginToken string) (*CheckLoginResponse, error) {
  req, _ := http.NewRequest("GET", c.BaseURL + "/telegram/login/check?login_token=" + loginToken, nil)
  resp, err := c.HTTP.Do(req)
  if err != nil { return nil, err }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("auth check login: статус %d", resp.StatusCode)
  }
  var out CheckLoginResponse
  if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
  return &out, nil
}

// VerifyCode verifies one-time code for code-based auth.
// POST {base}/telegram/login/verify-code body {"login_token":"...","code":"1234"}
type VerifyCodeResponse struct {
  AccessToken string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  AccessExpUnix int64 `json:"access_exp_unix"`
}

func (c *AuthClient) VerifyCode(loginToken, code string) (*VerifyCodeResponse, error) {
  payload := map[string]any{"login_token": loginToken, "code": code}
  b, _ := json.Marshal(payload)
  req, _ := http.NewRequest("POST", c.BaseURL + "/telegram/login/verify-code", bytes.NewReader(b))
  req.Header.Set("Content-Type", "application/json")
  resp, err := c.HTTP.Do(req)
  if err != nil { return nil, err }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("auth verify code: статус %d", resp.StatusCode)
  }
  var out VerifyCodeResponse
  if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
  return &out, nil
}

// Refresh tokens.
// POST {base}/telegram/token/refresh body {"refresh_token":"..."} -> {"access_token":"...","refresh_token":"...","access_exp_unix":123}
type RefreshResponse struct {
  AccessToken string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  AccessExpUnix int64 `json:"access_exp_unix"`
}

func (c *AuthClient) Refresh(refreshToken string) (*RefreshResponse, error) {
  payload := map[string]any{"refresh_token": refreshToken}
  b, _ := json.Marshal(payload)
  req, _ := http.NewRequest("POST", c.BaseURL + "/telegram/token/refresh", bytes.NewReader(b))
  req.Header.Set("Content-Type", "application/json")
  resp, err := c.HTTP.Do(req)
  if err != nil { return nil, err }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("auth refresh: статус %d", resp.StatusCode)
  }
  var out RefreshResponse
  if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
  return &out, nil
}
