package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CallAIAgent — отправка данных в AI агента (который работает на другом устройстве)
// Архитектура: Бот → AI агент (удалённый) → Бэкенд
// AI агент получает JSON от бота, формирует промпт и вызывает бэкенд
func CallAIAgent(endpoint string, data map[string]interface{}) (PostJSON, error) {
	aiAgentURL := os.Getenv("AI_AGENT_URL")
	if aiAgentURL == "" {
		return PostJSON{}, fmt.Errorf("AI_AGENT_URL not set in .env")
	}

	// Добавляем метаданные для AI агента
	requestData := map[string]interface{}{
		"endpoint":  endpoint, // Куда AI агент должен обратиться в бэкенде
		"data":      data,     // Данные для обработки
		"timestamp": time.Now().Unix(),
	}

	body, err := json.Marshal(requestData)
	if err != nil {
		return PostJSON{}, err
	}

	// Логируем отправляемый JSON для отладки
	log.Printf("[DEBUG] Sending to AI agent (%s): %s", endpoint, string(body))

	// Создаём HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: 60 * time.Second, // Таймаут 60 секунд для генерации
	}

	resp, err := client.Post(aiAgentURL+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return PostJSON{}, fmt.Errorf("AI agent connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Читаем тело ответа для деталей ошибки
		bodyBytes := make([]byte, 1024)
		n, _ := resp.Body.Read(bodyBytes)
		errorBody := string(bodyBytes[:n])

		return PostJSON{}, fmt.Errorf("AI agent error: status %d, details: %s", resp.StatusCode, errorBody)
	}

	var post PostJSON // из models.go
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return PostJSON{}, fmt.Errorf("AI agent response parse error: %w", err)
	}

	return post, nil
}

// CallBackend — отправка данных в AI агента (обёртка для удобства)
func CallBackend(endpoint string, data map[string]interface{}) (PostJSON, error) {
	return CallAIAgent(endpoint, data)
}

// SendPostToUser — отправка сгенерированного поста в чат (текст + картинки)
func SendPostToUser(chatID int64, post PostJSON, bot *tgbotapi.BotAPI) error {
	// Отправляем основной текст
	msg := tgbotapi.NewMessage(chatID, post.MainText)
	bot.Send(msg)

	// Отправляем изображения из content
	for _, layer := range post.Content {
		switch layer.Type {
		case "image":
			url, ok := layer.Data["url"].(string)
			if ok {
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(url))
				bot.Send(photo)
			}
		case "text":
			// Если есть текстовые слои, можно отправить их отдельно
			text, ok := layer.Data["text"].(string)
			if ok {
				textMsg := tgbotapi.NewMessage(chatID, text)
				bot.Send(textMsg)
			}
			// rectangle и другие типы игнорируем для отправки
		}
	}
	return nil
}
