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
			tgbotapi.NewKeyboardButton("Ввести данные НКО"),
			tgbotapi.NewKeyboardButton("Помощь"),
		),
	)
}

// YesNoInline — inline-кнопки для да/нет (например, для опроса НКО)
func YesNoInline() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Ввести данные НКО", "nko_yes"),
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "nko_skip"),
		),
	)
}

// TextModesInline — режимы для генерации текста (свободный, структурированный)
func TextModesInline() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Свободный текст", "text_free"),
			tgbotapi.NewInlineKeyboardButtonData("Структурированная форма", "text_struct"),
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
			tgbotapi.NewInlineKeyboardButtonData("Эмоциональный", "style_emotional"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Информационный", "style_informational"),
			tgbotapi.NewInlineKeyboardButtonData("Призыв к действию", "style_call_to_action"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Благодарственный", "style_gratitude"),
			tgbotapi.NewInlineKeyboardButtonData("Дружелюбный", "style_friendly"),
		),
	)
}

// ContentPlanPeriodInline — выбор периода для контент-плана
func ContentPlanPeriodInline() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("7 дней", "plan_7"),
			tgbotapi.NewInlineKeyboardButtonData("14 дней", "plan_14"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("30 дней", "plan_30"),
			tgbotapi.NewInlineKeyboardButtonData("Свой вариант", "plan_custom"),
		),
	)
}

// ContentPlanFrequencyInline — выбор частоты публикаций для контент-плана
func ContentPlanFrequencyInline() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Ежедневно", "freq_daily"),
			tgbotapi.NewInlineKeyboardButtonData("Через день", "freq_every_other"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("2 раза в неделю", "freq_twice_week"),
			tgbotapi.NewInlineKeyboardButtonData("3 раза в неделю", "freq_thrice_week"),
		),
	)
}

// PostActionInline — действия с готовым постом
func PostActionInline(postID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Перегенерировать", "post_regenerate_"+postID),
			tgbotapi.NewInlineKeyboardButtonData("📤 Отправить", "post_send_"+postID),
		),
	)
}
