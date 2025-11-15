package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		HandleUpdate(update, bot) // Теперь вся логика в handlers.go
	}
}

// HandleUpdate — заглушка, перенесём в handlers.go позже
func HandleUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я NKOshka Bot. Выбери действие:")
		msg.ReplyMarkup = MainMenu() // Используем клавиатуру из keyboards.go
		bot.Send(msg)
	}
}
