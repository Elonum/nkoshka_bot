package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleUpdate — основная обработка (вызывается из main.go)
func HandleUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message != nil {
		handleMessage(update.Message, bot)
	} else if update.CallbackQuery != nil {
		handleCallback(update.CallbackQuery, bot)
	}
}

// handleMessage — обработка текстовых сообщений и команд
func handleMessage(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	chatID := message.Chat.ID
	text := message.Text

	state := GetUserState(chatID) // из states.go

	// Если в состоянии опроса/ввода — обрабатываем как ответ
	if state.State != "idle" {
		processStateInput(state, text, bot)
		return
	}

	// Обычные команды/кнопки
	switch text {
	case "/start":
		msg := tgbotapi.NewMessage(chatID, "Привет! Расскажи о своей НКО для лучших результатов?")
		msg.ReplyMarkup = YesNoInline() // из keyboards.go
		bot.Send(msg)
	case "Генерация текста":
		msg := tgbotapi.NewMessage(chatID, "Выбери режим генерации текста:")
		msg.ReplyMarkup = TextModesInline() // из keyboards.go
		bot.Send(msg)
	case "Генерация картинки":
		state.State = "image_desc"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "Опиши картинку (или прикрепи файл):"))
	case "Редактор текста":
		state.State = "edit_text"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "Введи текст для редактирования:"))
	case "Контент-план":
		state.State = "plan_period"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "На сколько дней план? (например, 7)"))
	case "Настройки НКО":
		msg := tgbotapi.NewMessage(chatID, "Обновить данные НКО?")
		msg.ReplyMarkup = YesNoInline()
		bot.Send(msg)
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Не понял. Выбери из меню:"))
	}
}

// processStateInput — обработка ввода в состояниях
func processStateInput(state *UserState, input string, bot *tgbotapi.BotAPI) {
	chatID := state.ChatID

	switch state.State {
	case "nko_name":
		state.NKO.Name = input
		state.State = "nko_desc"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "Опиши НКО:"))
	case "nko_desc":
		state.NKO.Description = input
		state.State = "nko_activities"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "Формы деятельности:"))
	case "nko_activities":
		state.NKO.Activities = input
		state.State = "nko_style"
		msg := tgbotapi.NewMessage(chatID, "Выбери стиль постов:")
		msg.ReplyMarkup = StylesInline() // из keyboards.go
		bot.Send(msg)
	case "image_desc":
		// Здесь вызов бэкенда: CallBackend("/generate_image", map[string]string{"desc": input})
		// Пока заглушка
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Генерирую картинку по: %s", input)))
		ResetUserState(chatID)
		// ... аналогично для других состояний (edit_text, plan_period и т.д.)
	}
}

// handleCallback — обработка inline-кнопок
func handleCallback(callback *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	bot.Request(tgbotapi.NewCallback(callback.ID, "")) // подтверждение

	state := GetUserState(chatID)

	switch data {
	case "nko_yes":
		state.State = "nko_name"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "Название НКО:"))
	case "nko_skip":
		msg := tgbotapi.NewMessage(chatID, "Ок, используем обезличенные посты. Выбери функцию:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
	case "text_free":
		state.State = "text_free_input"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "Введи идею для поста:"))
		// ... другие callback (style_*, text_struct и т.д.)
	}
}
