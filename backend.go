package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"golang.org/x/image/font/basicfont"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CallAIAgent — отправка данных в AI агента (который работает на другом устройстве)
// Архитектура: Бот → AI агент (удалённый) → Бэкенд
// AI агент получает JSON от бота, формирует промпт и вызывает бэкенд
func CallAIAgent(endpoint string, data map[string]interface{}, tgID int64) (PostJSON, error) {
	aiAgentURL := os.Getenv("AI_AGENT_URL")
	if aiAgentURL == "" {
		return PostJSON{}, fmt.Errorf("AI_AGENT_URL not set in .env")
	}

	// Добавляем метаданные для AI агента
	requestData := map[string]interface{}{
		"endpoint":  endpoint, // Куда AI агент должен обратиться в бэкенде
		"data":      data,     // Данные для обработки
		"tg_id":     tgID,     // ID пользователя Telegram (обязательно для бэкенда)
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
func CallBackend(endpoint string, data map[string]interface{}, tgID int64) (PostJSON, error) {
	return CallAIAgent(endpoint, data, tgID)
}

// InitUser — инициализация пользователя в бэкенде через AI агента
func InitUser(tgID int64, username string) error {
	aiAgentURL := os.Getenv("AI_AGENT_URL")
	if aiAgentURL == "" {
		return fmt.Errorf("AI_AGENT_URL not set in .env")
	}

	data := map[string]interface{}{
		"tg_id":    tgID,
		"username": username,
	}

	requestData := map[string]interface{}{
		"endpoint":  "/api/auth/init",
		"data":      data,
		"tg_id":     tgID,
		"timestamp": time.Now().Unix(),
	}

	body, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Initializing user: tg_id=%d, username=%s", tgID, username)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(aiAgentURL+"/api/auth/init", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("AI agent connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes := make([]byte, 1024)
		n, _ := resp.Body.Read(bodyBytes)
		errorBody := string(bodyBytes[:n])
		return fmt.Errorf("AI agent error: status %d, details: %s", resp.StatusCode, errorBody)
	}

	log.Printf("[DEBUG] User initialized successfully: tg_id=%d", tgID)
	return nil
}

// SendPostToUser — отправка сгенерированного поста в чат (текст + объединённое изображение из слоёв)
func SendPostToUser(chatID int64, post PostJSON, bot *tgbotapi.BotAPI) error {
	// Отправляем основной текст
	if post.MainText != "" {
		msg := tgbotapi.NewMessage(chatID, post.MainText)
		bot.Send(msg)
	}

	// Если есть слои, объединяем их в одно изображение
	if len(post.Content) > 0 {
		// Сортируем слои по order_index (используем поле из структуры Layer)
		layers := make([]Layer, len(post.Content))
		copy(layers, post.Content)
		for i := 0; i < len(layers)-1; i++ {
			for j := i + 1; j < len(layers); j++ {
				if layers[i].OrderIndex > layers[j].OrderIndex {
					layers[i], layers[j] = layers[j], layers[i]
				}
			}
		}

		// Объединяем слои в одно изображение
		finalImageBytes, err := composeLayers(layers)
		if err != nil {
			log.Printf("[ERROR] Failed to compose layers: %v", err)
			// Fallback: отправляем изображения отдельно, если не удалось объединить
			return sendLayersSeparately(chatID, post.Content, bot)
		}

		// Отправляем итоговое изображение
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
			Name:  "post.jpg",
			Bytes: finalImageBytes,
		})
		bot.Send(photo)
	}

	return nil
}

// composeLayers — объединяет все слои в одно изображение
func composeLayers(layers []Layer) ([]byte, error) {
	// Определяем размеры canvas (по умолчанию 1080x1080 для квадратного поста)
	canvasWidth := 1080
	canvasHeight := 1080

	// Ищем максимальные размеры из всех слоёв
	for _, layer := range layers {
		if x, ok := layer.Data["x"].(float64); ok {
			if w, ok := layer.Data["w"].(float64); ok {
				if int(x+w) > canvasWidth {
					canvasWidth = int(x + w)
				}
			}
		}
		if y, ok := layer.Data["y"].(float64); ok {
			if h, ok := layer.Data["h"].(float64); ok {
				if int(y+h) > canvasHeight {
					canvasHeight = int(y + h)
				}
			}
		}
	}

	// Создаём canvas
	dc := gg.NewContext(canvasWidth, canvasHeight)
	dc.SetRGB(1, 1, 1) // Белый фон
	dc.Clear()

	// Рисуем слои по порядку
	for _, layer := range layers {
		switch layer.Type {
		case "rectangle":
			drawRectangle(dc, layer.Data)
		case "image":
			if err := drawImage(dc, layer.Data); err != nil {
				log.Printf("[WARN] Failed to draw image layer: %v", err)
			}
		case "text":
			drawText(dc, layer.Data)
		}
	}

	// Конвертируем в JPEG байты
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dc.Image(), &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// drawRectangle — рисует прямоугольник
func drawRectangle(dc *gg.Context, data map[string]interface{}) {
	x := getFloat(data, "x", 0)
	y := getFloat(data, "y", 0)
	w := getFloat(data, "width", 100)
	if w == 0 {
		w = getFloat(data, "w", 100)
	}
	h := getFloat(data, "height", 100)
	if h == 0 {
		h = getFloat(data, "h", 100)
	}

	// Парсим цвет
	colorStr := getString(data, "color", "#ffffff")
	c := parseColor(colorStr)

	dc.SetColor(c)
	dc.DrawRectangle(x, y, w, h)
	dc.Fill()
}

// drawImage — накладывает изображение
func drawImage(dc *gg.Context, data map[string]interface{}) error {
	// Пробуем получить base64 изображение
	imageBase64, ok := data["image_base64"].(string)
	if !ok || imageBase64 == "" {
		return fmt.Errorf("no image_base64 in layer data")
	}

	// Декодируем base64
	imageBytes, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Декодируем изображение
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	x := getFloat(data, "x", 0)
	y := getFloat(data, "y", 0)
	scale := getFloat(data, "scale", 1.0)
	opacity := getFloat(data, "opacity", 1.0)

	// Применяем opacity
	dc.Push()
	dc.SetAlpha(opacity)
	
	// Применяем масштаб и позицию
	if scale != 1.0 {
		// Сохраняем текущее состояние
		dc.Push()
		// Перемещаемся в позицию
		dc.Translate(x, y)
		// Масштабируем
		dc.Scale(scale, scale)
		// Рисуем изображение в начале координат (после трансформации)
		dc.DrawImage(img, 0, 0)
		dc.Pop()
	} else {
		dc.DrawImage(img, int(x), int(y))
	}
	
	dc.Pop()

	return nil
}

// drawText — рисует текст
func drawText(dc *gg.Context, data map[string]interface{}) {
	text := getString(data, "text", "")
	if text == "" {
		return
	}

	x := getFloat(data, "x", 0)
	y := getFloat(data, "y", 0)
	fontSize := getFloat(data, "font_size", 48)
	colorStr := getString(data, "color", "#000000")
	align := getString(data, "align", "left")

	// Парсим цвет
	c := parseColor(colorStr)
	dc.SetColor(c)

	// Устанавливаем шрифт
	face := basicfont.Face7x13
	dc.SetFontFace(face)
	dc.SetFontSize(fontSize)

	// Выравнивание
	switch align {
	case "center":
		dc.DrawStringAnchored(text, x, y, 0.5, 0.5)
	case "right":
		dc.DrawStringAnchored(text, x, y, 1, 0.5)
	default:
		dc.DrawString(text, x, y)
	}
}

// parseColor — парсит цвет из строки (#rrggbb или #rrggbbaa)
func parseColor(colorStr string) color.Color {
	colorStr = strings.TrimPrefix(colorStr, "#")
	if len(colorStr) == 6 {
		r, _ := strconv.ParseUint(colorStr[0:2], 16, 8)
		g, _ := strconv.ParseUint(colorStr[2:4], 16, 8)
		b, _ := strconv.ParseUint(colorStr[4:6], 16, 8)
		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	} else if len(colorStr) == 8 {
		r, _ := strconv.ParseUint(colorStr[0:2], 16, 8)
		g, _ := strconv.ParseUint(colorStr[2:4], 16, 8)
		b, _ := strconv.ParseUint(colorStr[4:6], 16, 8)
		a, _ := strconv.ParseUint(colorStr[6:8], 16, 8)
		return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	}
	return color.Black
}

// getFloat — безопасно получает float64 из map
func getFloat(data map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

// getString — безопасно получает string из map
func getString(data map[string]interface{}, key string, defaultValue string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// sendLayersSeparately — fallback: отправляет слои отдельно (старая логика)
func sendLayersSeparately(chatID int64, layers []Layer, bot *tgbotapi.BotAPI) error {
	for _, layer := range layers {
		switch layer.Type {
		case "image":
			imageBase64, ok := layer.Data["image_base64"].(string)
			if ok && imageBase64 != "" {
				imageBytes, err := base64.StdEncoding.DecodeString(imageBase64)
				if err != nil {
					log.Printf("[ERROR] Failed to decode base64 image: %v", err)
					continue
				}
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
					Name:  "image.jpg",
					Bytes: imageBytes,
				})
				bot.Send(photo)
			}
		case "text":
			text, ok := layer.Data["text"].(string)
			if ok {
				textMsg := tgbotapi.NewMessage(chatID, text)
				bot.Send(textMsg)
			}
		}
	}
	return nil
}
