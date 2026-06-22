package main

import (
	"fmt"
	"game-server-monitor/internal/bot"
	"game-server-monitor/internal/config"
	"game-server-monitor/internal/database"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	fmt.Println("Запускаю бота для мониторинга игровых серверов...")

	cfg := config.Load()

	storage, err := database.NewStorage("data/users.json")
	if err != nil {
		log.Fatalf("Ошибка при создании базы данных: %v", err)
	}

	botAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Не удалось создать бота: %v", err)
	}

	botAPI.Debug = false
	log.Printf("Бот запущен под аккаунтом %s", botAPI.Self.UserName)

	handler := bot.NewBotHandler(storage, botAPI)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nПолучен сигнал остановки, завершаю работу...")
		os.Exit(0)
	}()

	fmt.Println("Бот успешно запущен и слушает сообщения.")

	for update := range updates {
		if update.Message != nil {
			go handler.HandleMessage(update.Message)
		}
	}
}