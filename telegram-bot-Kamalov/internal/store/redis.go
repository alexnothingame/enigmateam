
package store

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/go-redis/redis/v8"
)

type SessionState string

const (
  StateUnknown   SessionState = "unknown"
  StateAnonymous SessionState = "anonymous"
  StateAuthorized SessionState = "authorized"
)

type Session struct {
  ChatID int64 `json:"chat_id"`
  State SessionState `json:"state"`

  // Login flow
  LoginType string `json:"login_type,omitempty"` // github|yandex|code
  LoginToken string `json:"login_token,omitempty"`
  // For code flow: we wait for user to send code
  WaitingCode bool `json:"waiting_code,omitempty"`

  // Auth tokens
  AccessToken  string `json:"access_token,omitempty"`
  RefreshToken string `json:"refresh_token,omitempty"`
  AccessExpUnix int64 `json:"access_exp_unix,omitempty"`

  // Simple test run state
  ActiveAttemptID string `json:"active_attempt_id,omitempty"`
  ActiveTestID string `json:"active_test_id,omitempty"`
}

type Store struct {
  rdb *redis.Client
  ctx context.Context
  prefix string
}

func New(rdb *redis.Client, prefix string) *Store {
  return &Store{rdb:rdb, ctx: context.Background(), prefix: prefix}
}

func (s *Store) key(chatID int64) string {
  return fmt.Sprintf("%s:session:%d", s.prefix, chatID)
}

func (s *Store) Get(chatID int64) (*Session, error) {
  val, err := s.rdb.Get(s.ctx, s.key(chatID)).Result()
  if err == redis.Nil {
    return nil, nil
  }
  if err != nil { return nil, err }
  var sess Session
  if err := json.Unmarshal([]byte(val), &sess); err != nil {
    return nil, err
  }
  return &sess, nil
}

func (s *Store) Set(chatID int64, sess *Session, ttl time.Duration) error {
  sess.ChatID = chatID
  b, err := json.Marshal(sess)
  if err != nil { return err }
  return s.rdb.Set(s.ctx, s.key(chatID), b, ttl).Err()
}

func (s *Store) Delete(chatID int64) error {
  return s.rdb.Del(s.ctx, s.key(chatID)).Err()
}

// ScanSessions returns sessions matching pattern. Used by cron jobs.
func (s *Store) ScanSessions() ([]*Session, error) {
  var cursor uint64
  var out []*Session
  pattern := fmt.Sprintf("%s:session:*", s.prefix)
  for {
    keys, next, err := s.rdb.Scan(s.ctx, cursor, pattern, 200).Result()
    if err != nil { return nil, err }
    cursor = next
    if len(keys) > 0 {
      vals, err := s.rdb.MGet(s.ctx, keys...).Result()
      if err != nil { return nil, err }
      for _, v := range vals {
        if v == nil { continue }
        str, ok := v.(string); if !ok { continue }
        var sess Session
        if err := json.Unmarshal([]byte(str), &sess); err == nil {
          out = append(out, &sess)
        }
      }
    }
    if cursor == 0 { break }
  }
  return out, nil
}
