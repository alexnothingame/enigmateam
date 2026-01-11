
package main

import (
  "log"
  "net/http"
  "os"
  "time"

  "github.com/go-chi/chi/v5"
  "github.com/go-chi/chi/v5/middleware"
  "github.com/go-redis/redis/v8"

  "telegram-module/internal/api"
  "telegram-module/internal/config"
  "telegram-module/internal/botlogic"
  "telegram-module/internal/store"
)

func main() {
  // env
  redisAddr := getenv("REDIS_ADDR", "redis:6379")
  redisPass := os.Getenv("REDIS_PASSWORD")
  redisDB := 0

  authURL := getenv("AUTH_BASE_URL", "http://auth:8000")
  mainURL := getenv("MAIN_BASE_URL", "http://main:8080")
  commandsPath := getenv("COMMANDS_PATH", "/app/commands.json")
  sessionTTL := getenvDur("SESSION_TTL", 72*time.Hour)

  notifPath := getenv("MAIN_NOTIFICATION_PATH", "/notification")
  notifAckPath := getenv("MAIN_NOTIFICATION_ACK_PATH", "/notification/ack")

  // redis
  rdb := redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisPass, DB: redisDB})
  if err := rdb.Ping(rdb.Context()).Err(); err != nil {
    log.Fatal("Redis недоступен:", err)
  }

  reg, err := config.LoadRegistry(commandsPath)
  if err != nil { log.Fatal(err) }

  st := store.New(rdb, "tg")
  auth := api.NewAuthClient(authURL)
  main := api.NewMainClient(mainURL)

  h := botlogic.NewHandler(botlogic.Deps{
    Store: st,
    Registry: reg,
    Auth: auth,
    Main: main,
    SessionTTL: sessionTTL,
    NotificationPath: notifPath,
    NotificationAckPath: notifAckPath,
  })

  r := chi.NewRouter()
  r.Use(middleware.RequestID)
  r.Use(middleware.RealIP)
  r.Use(middleware.Logger)
  r.Use(middleware.Recoverer)

  r.Get("/health", func(w http.ResponseWriter, r *http.Request){
    w.WriteHeader(200); w.Write([]byte("ok"))
  })

  r.Post("/update", h.HandleUpdate)
  r.Get("/cron/check-login", h.CronCheckLogin)
  r.Get("/cron/check-notifications", h.CronCheckNotifications)

  addr := getenv("LISTEN_ADDR", ":8080")
  log.Println("Bot Logic слушает", addr)
  log.Fatal(http.ListenAndServe(addr, r))
}

func getenv(k, d string) string {
  if v := os.Getenv(k); v != "" { return v }
  return d
}

func getenvDur(k string, d time.Duration) time.Duration {
  if v := os.Getenv(k); v != "" {
    if dur, err := time.ParseDuration(v); err == nil { return dur }
  }
  return d
}
