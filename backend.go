package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CallBackend — универсальный вызов API бэкенда
// endpoint: "/generate_text", "/generate_image" и т.д.
// data: map с параметрами (nko, idea, desc...)
func CallBackend(endpoint string, data map[string]interface{}) (PostJSON, error) {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		return PostJSON{}, fmt.Errorf("BACKEND_URL not set in .env")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return PostJSON{}, err
	}

	resp, err := http.Post(backendURL+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return PostJSON{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PostJSON{}, fmt.Errorf("backend error: %d", resp.StatusCode)
	}

	var post PostJSON // из models.go
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return PostJSON{}, err
	}

	return post, nil
}

// SendPostToUser — отправка сгенерированного поста в чат (текст + картинки)
func SendPostToUser(chatID int64, post PostJSON, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatID, post.MainText)
	bot.Send(msg)

	for _, layer := range post.Content {
		switch layer.Type {
		case "image":
			url, ok := layer.Data["url"].(string)
			if ok {
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(url))
				bot.Send(photo)
			}
			// ... можно добавить для text (если нужно), rectangle игнорим
		}
	}
	return nil
}
