# scraper-bot

Telegram-бот на Go для скачивания видео по ссылкам (TikTok / YouTube / Instagram) и отправки результата обратно в чат.  
Показывает уверенный backend-подход: разделение слоев, обработка сообщений в очереди, работа с goroutines, Docker и конфиг через env.

## Технологии

- Go 1.24
- Telegram Bot API (`github.com/kartmos/kartmos-telegram-bot-api/v5`)
- `viper` для конфигурации (env + optional yaml)
- `yt-dlp` для скачивания видео (YouTube/TikTok)
- `ffmpeg-go` + `ffmpeg` для сжатия файлов > 50 MB
- RapidAPI (резолвинг прямой ссылки для Instagram Reel)
- Docker / Docker Compose
- goroutines и каналы (long polling + обработка апдейтов)

## Архитектура (простыми словами)

Бот запускает long polling к Telegram и получает входящие апдейты в канал (`updates`). Это и есть очередь сообщений на входе.

Дальше поток такой:

1. `cmd/app/main.go` собирает зависимости: конфиг, Telegram-клиент, сервис, репозитории.
2. `internal/transport/telegram/client.go` читает апдейты и передает сообщения в сервис.
3. `internal/service/bot_service.go` решает, что делать:
   - команды (`/start`, `/help`, `/status`, `/log`),
   - или обработка ссылки на видео.
4. `internal/repository/video/repository.go` скачивает видео:
   - через `yt-dlp` (YouTube/TikTok),
   - или через RapidAPI + прямую загрузку (Instagram).
5. Если файл слишком большой, сервис сжимает его через `ffmpeg` и отправляет в Telegram.

> Внутри приложения активно используются goroutines и каналы Go: Telegram updates-канал, сетевые запросы, long polling, обработка через `context` и graceful shutdown по `SIGINT/SIGTERM`.

## Конфиг через env

Минимум для старта:

- `BOT_TOKEN` — токен Telegram-бота (обязательно)

Полезные дополнительные переменные:

- `RAPID_TOKENS` — токены RapidAPI для Instagram через запятую
- `ADMIN_ID` — Telegram user id администратора
- `DOWNLOAD_DIR` — путь к папке загрузок (default: `./downloads`)
- `YT_DLP_PATH` — путь к бинарнику `yt-dlp` (default локально: `./bin/yt-dlp`, в Docker: `/app/bin/yt-dlp`)
- `HELP_FILE`, `WELCOME_FILE`, `COMMAND_FILE` — пути к текстам команд/приветствия
- `LOG_FILE` — путь к лог-файлу (используется командой `/log`)
- `APP_CONFIG_PATH` — путь к yaml-конфигу (опционально, если нужен не дефолтный путь)

Пример:

```bash
export BOT_TOKEN="123456:your-token"
export ADMIN_ID="291182090"
export RAPID_TOKENS="token1,token2"
export YT_DLP_PATH="./bin/yt-dlp"
export LOG_FILE="./logs/bot.log"
```

Или через файл `.env`:

```bash
cp .env.example .env
```

## Запуск

### 1) Локально

```bash
go mod download
make yt-dlp
mkdir -p downloads logs
go run ./cmd/app
```

Или сборка бинаря:

```bash
make build
./bin/bot
```

### 2) Через Docker

Сборка:

```bash
docker build -f build/deploy/Dockerfile -t scraper-bot .
```

Запуск:

```bash
docker run --rm \
  --name scraper-bot \
  -e BOT_TOKEN="123456:your-token" \
  -e ADMIN_ID="291182090" \
  -e RAPID_TOKENS="token1,token2" \
  -e YT_DLP_PATH="/app/bin/yt-dlp" \
  scraper-bot
```

Проверка логов контейнера:

```bash
docker logs -f scraper-bot
```

### 3) Через Docker Compose

```bash
cp .env.example .env
docker compose -f build/deploy/docker-compose.yaml up --build -d
docker compose -f build/deploy/docker-compose.yaml logs -f
```

## Пример использования бота

1. Отправьте `/start` — бот пришлет приветственное сообщение.
2. Отправьте `/help` — бот покажет подсказку по командам.
3. Отправьте сообщение с ссылкой, например:

```text
https://www.instagram.com/reel/XXXXXXXXX/
Очень важный мем
```

4. Бот скачает видео, при необходимости сожмет и отправит обратно с подписью.
5. Если вы admin (`ADMIN_ID`), доступны `/status` и `/log`.

## Структура проекта

```text
.
├── .env.example
├── cmd/
│   └── app/
│       └── main.go               # точка входа
├── config/
│   └── config.go                 # загрузка конфига (env/yaml)
├── internal/
│   ├── transport/
│   │   └── telegram/
│   │       └── client.go         # Telegram long polling + отправка сообщений
│   ├── service/
│   │   ├── bot_service.go        # бизнес-логика бота
│   │   └── contracts.go          # интерфейсы и модели
│   └── repository/
│       ├── video/
│       │   └── repository.go     # скачивание видео
│       └── rapid/
│           └── reel_resolver.go  # резолв Instagram Reel через RapidAPI
├── build/
│   └── deploy/
│       ├── Dockerfile
│       └── docker-compose.yaml
├── asserts/
│   ├── help.txt
│   ├── welcome.txt
│   └── command.txt
├── Makefile
├── go.mod
└── go.sum
```

## Логирование

- Критичные ошибки старта пишутся через стандартный `log`.
- Для Docker удобнее смотреть вывод через `docker logs`.
- Команда `/log` отправляет администратору файл по пути `LOG_FILE`.
