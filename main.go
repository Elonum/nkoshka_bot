package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env (сначала текущая директория, потом абсолютный путь)
	err := godotenv.Load()
	if err != nil {
		// Пробуем абсолютный путь для GoLand
		err = godotenv.Load("C:\\MainProjects\\nko_bot_frontend\\.env")
		if err != nil {
			log.Println("No .env file found")
		}
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Panic("BOT_TOKEN not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Инициализация хранилища данных
	InitDB()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		HandleUpdate(update, bot) // Теперь вся логика в handlers.go
	}
}
