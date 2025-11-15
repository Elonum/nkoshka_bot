package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// MainMenu — основное меню с функциями ТЗ
func MainMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Генерация текста"),
			tgbotapi.NewKeyboardButton("Генерация картинки"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Редактор текста"),
			tgbotapi.NewKeyboardButton("Контент-план"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Настройки НКО"),
			tgbotapi.NewKeyboardButton("Помощь"),
		),
	)
}

// YesNoInline — inline-кнопки для да/нет (например, для опроса НКО)
func YesNoInline() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, рассказать о НКО", "nko_yes"),
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "nko_skip"),
		),
	)
}

// TextModesInline — режимы для генерации текста (свободный, структурированный, по примеру)
func TextModesInline() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Свободный текст", "text_free"),
			tgbotapi.NewInlineKeyboardButtonData("Структурированная форма", "text_struct"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("По примеру (как Ночлежка)", "text_example"),
		),
	)
}

// StylesInline — выбор стиля поста (рекомендация ТЗ для креатива)
func StylesInline() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Разговорный", "style_conversational"),
			tgbotapi.NewInlineKeyboardButtonData("Официальный", "style_formal"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Художественный", "style_artistic"),
		),
	)
}

// ... Можно добавить больше, например, для контент-плана (период: неделя, месяц)
