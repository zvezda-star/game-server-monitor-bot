package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken string
}

// загружаю переменные из .env
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Предупреждение: файл .env не найден, использую системные переменные")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Ошибка: переменная BOT_TOKEN не найдена в .env файле")
	}

	return &Config{
		BotToken: token,
	}
}