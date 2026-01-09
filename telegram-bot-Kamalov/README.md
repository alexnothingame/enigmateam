
# Telegram-модуль (Telegram Client + Bot Logic + Nginx + Redis)

Этот репозиторий — готовый модуль Telegram для вашего проекта **строго по архитектуре**:
- **telegram-client**: подключается к Telegram Servers, получает апдейты, отправляет их в Bot Logic через Nginx, отправляет ответы пользователям
- **bot-logic**: HTTP-сервис, хранит состояния/токены в Redis, ходит в Auth и Главный модуль по HTTP
- **nginx**: балансирует запросы на 2 копии bot-logic
- **redis**: хранение сессий пользователя

## Быстрый старт

1) Скопируйте `.env.example` в `.env` и вставьте `TELEGRAM_TOKEN`.

2) Запуск:
```bash
docker compose up --build
```

## Основные команды

- `/start` — старт
- `/login github|yandex|code` — вход
- `/logout` — выход
- `/menu` — меню разделов и команд
- `/help` — список всех команд
- `/cancel` — отмена авторизации/сброс локального состояния

## Команды действий ("~40 команд" из таблиц)

Все команды описаны в `commands.json`.
Если у вашего Главного модуля другие пути/методы — правьте `commands.json` (и при необходимости шаблоны body в `internal/botlogic/calls.go`).

Примеры:
- `/users_list`
- `/user_get 12`
- `/user_fullname_set 12 "Иванов Иван"`
- `/courses_list`
- `/course_tests 3`
- `/question_create "Название" "Текст" "a,b,c" 1`
- `/test_question_add 10 55`
- `/attempt_create 10`

## Важно про интеграцию с Auth модулем

Bot Logic ожидает от Auth модуля следующие эндпоинты (их можно легко согласовать/переименовать в одном месте `internal/api/auth.go`):

- `POST /telegram/login/start`
- `GET  /telegram/login/check?login_token=...`
- `POST /telegram/login/verify-code`
- `POST /telegram/token/refresh`

Если у вас API отличается — просто поправьте клиент `internal/api/auth.go`.

## Проверка сценария "циклические запросы"

Telegram Client по таймеру вызывает:
- `/cron/check-login` (каждые 10 секунд)
- `/cron/check-notifications` (каждые 20 секунд)

Это соответствует вашим описанным циклическим проверкам.
