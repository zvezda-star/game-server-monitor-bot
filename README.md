# Game Server Monitor Bot

Telegram бот для мониторинга игровых серверов. Поддерживает Minecraft и все игры Steam (CS:GO, CS2, TF2, DayZ, Left 4 Dead 2 и другие).

## Команды

`/add minecraft IP:port` - Добавить Minecraft сервер
`/add steam IP:port` - Добавить Steam сервер
`/status` - Проверить статус всех серверов |
`/remove номер` - Удалить сервер по номеру
`/privacy` - Политика конфиденциальности
`/about` - Информация о боте

## Установка

1. Создай файл `.env` с токеном бота
2. Запусти `go mod tidy`
3. Запусти `go run cmd/bot/main.go`
