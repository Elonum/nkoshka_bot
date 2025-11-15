package main // для теста, но в проекте — без main

import (
	"strconv"

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

	// Обработка загруженных файлов (изображений для генерации картинок)
	if message.Photo != nil && len(message.Photo) > 0 {
		if state.State == "image_desc" {
			// Получаем самое большое изображение
			photo := message.Photo[len(message.Photo)-1]
			fileID := photo.FileID

			// Получаем информацию о файле
			file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка загрузки файла: "+err.Error()+"\n\nПопробуй отправить изображение ещё раз."))
				return
			}

			// Формируем URL файла
			fileURL := file.Link(bot.Token)

			// Если есть подпись к фото, используем её как описание
			desc := message.Caption
			if desc == "" {
				desc = "Обработать изображение"
			}

			data := map[string]interface{}{
				"desc":      desc,
				"image_url": fileURL,
				"file_id":   fileID,
				"nko":       state.NKO,
			}

			post, err := CallBackend("/generate_image", data)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка генерации изображения: "+err.Error()+"\n\nПопробуй ещё раз или измени описание."))
			} else {
				SendPostToUser(chatID, post, bot)
				msg := tgbotapi.NewMessage(chatID, "✨ Готово! Выбери действие с постом:")
				msg.ReplyMarkup = PostActionInline(post.PostID)
				bot.Send(msg)
			}
			ResetUserState(chatID)
			return
		}
	}

	// Если в состоянии опроса/ввода — обрабатываем как ответ
	if state.State != "idle" {
		// Если нет текста, но есть состояние - просим ввести текст
		if text == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Пожалуйста, введи текстовое описание или прикрепи изображение."))
			return
		}
		processStateInput(state, text, bot)
		return
	}

	// Обычные команды/кнопки
	switch text {
	case "/start":
		welcomeText := `👋 Добро пожаловать в NKOshka Bot!

Я твой персональный SMM-менеджер для создания контента для НКО.

✨ Что я умею:
• 📝 Генерировать тексты постов (свободная форма или структурированная)
• 🎨 Создавать изображения по описанию
• ✏️ Исправлять ошибки в текстах
• 📅 Составлять контент-планы на любой период

🚀 Чтобы начать работу:
1. Расскажи о своей НКО — это поможет создавать более точный и релевантный контент
2. Выбери нужную функцию из меню

💡 Совет: Чем подробнее ты опишешь свою НКО, тем лучше будут результаты!

Готов начать? Расскажи о своей НКО для лучших результатов`
		msg := tgbotapi.NewMessage(chatID, welcomeText)
		msg.ReplyMarkup = YesNoInline()
		bot.Send(msg)
	case "/help", "Помощь":
		sendHelpMessage(bot, chatID)
	case "Генерация текста":
		msg := tgbotapi.NewMessage(chatID, "📝 Выбери режим генерации текста:\n\n• Свободный текст — опиши идею поста\n• Структурированная форма — пошаговый ввод данных о событии")
		msg.ReplyMarkup = TextModesInline()
		bot.Send(msg)
	case "Генерация картинки":
		state.State = "image_desc"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "🎨 Опиши картинку, которую нужно создать, или прикрепи изображение для обработки:\n\n💡 Чем подробнее описание, тем лучше результат!"))
	case "Редактор текста":
		state.State = "edit_text"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "✏️ Введи текст, который нужно исправить и улучшить:\n\nЯ найду ошибки, улучшу стиль и сделаю текст более читаемым."))
	case "Контент-план":
		msg := tgbotapi.NewMessage(chatID, "📅 На сколько дней создать контент-план?\n\nВыбери готовый вариант или укажи свой период.")
		msg.ReplyMarkup = ContentPlanPeriodInline()
		bot.Send(msg)
	case "Ввести данные НКО":
		// Показываем текущие данные НКО
		nkoInfo := "📋 Текущие данные НКО:\n\n"
		if state.NKO.Name != "" {
			nkoInfo += "🏷️ Название: " + state.NKO.Name + "\n"
		}
		if state.NKO.Description != "" {
			nkoInfo += "📝 Описание: " + state.NKO.Description + "\n"
		}
		if state.NKO.Activities != "" {
			nkoInfo += "🎯 Деятельность: " + state.NKO.Activities + "\n"
		}
		if state.NKO.Style != "" {
			nkoInfo += "✨ Стиль постов: " + state.NKO.Style + "\n"
		}
		if state.NKO.Name == "" && state.NKO.Description == "" {
			nkoInfo += "⚠️ Данные НКО не заполнены.\n\n"
			nkoInfo += "Для создания качественного контента рекомендуется заполнить информацию о НКО."
		}
		nkoInfo += "\n\n🔄 Обновить данные НКО?"
		msg := tgbotapi.NewMessage(chatID, nkoInfo)
		msg.ReplyMarkup = YesNoInline()
		bot.Send(msg)
	default:
		msg := tgbotapi.NewMessage(chatID, "❓ Не понял команду. Выбери действие из меню ниже:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
	}
}

// buildPrompt — формирует промпт для AI на основе данных пользователя
func buildPrompt(mode string, ideaOrExample string, desc string, nko NKOData, tempData map[string]string) string {
	var prompt string

	// Базовые данные о НКО
	nkoInfo := ""
	if nko.Name != "" {
		nkoInfo += "НКО: " + nko.Name + ". "
	}
	if nko.Description != "" {
		nkoInfo += "Описание: " + nko.Description + ". "
	}
	if nko.Activities != "" {
		nkoInfo += "Деятельность: " + nko.Activities + ". "
	}

	// Инструкция по стилю
	styleInstruction := ""
	if nko.Style != "" {
		styleInstructions := map[string]string{
			"разговорный":       "Используй разговорный, живой стиль, как в личном общении",
			"официальный":       "Используй официальный, деловой стиль, как в официальных документах",
			"художественный":    "Используй художественный, образный стиль с метафорами и яркими образами",
			"эмоциональный":     "Используй эмоциональный, вдохновляющий стиль, вызывающий чувства",
			"информационный":    "Используй информационный, новостной стиль, как в новостных заметках",
			"призыв к действию": "Используй стиль призыва к действию, мотивирующий и побуждающий",
			"благодарственный":  "Используй благодарственный, тёплый стиль, выражающий признательность",
			"дружелюбный":       "Используй дружелюбный, неформальный стиль, как в общении с друзьями",
		}
		if instruction, ok := styleInstructions[nko.Style]; ok {
			styleInstruction = instruction + ". "
		}
	}

	switch mode {
	case "free":
		prompt = styleInstruction + nkoInfo + "Создай пост на тему: " + ideaOrExample

	case "structured":
		event := tempData["event"]
		date := tempData["date"]
		location := tempData["location"]
		invited := tempData["invited"]
		details := tempData["details"]

		prompt = styleInstruction + nkoInfo + "Создай пост о событии. "
		if event != "" {
			prompt += "Событие: " + event + ". "
		}
		if date != "" {
			prompt += "Дата: " + date + ". "
		}
		if location != "" {
			prompt += "Место: " + location + ". "
		}
		if invited != "" {
			prompt += "Приглашённые: " + invited + ". "
		}
		if details != "" {
			prompt += "Дополнительные детали: " + details + ". "
		}

	default:
		prompt = styleInstruction + nkoInfo + "Создай пост"
	}

	return prompt
}

// processStateInput — обработка ввода в состояниях
func processStateInput(state *UserState, input string, bot *tgbotapi.BotAPI) {
	chatID := state.ChatID

	switch state.State {
	case "nko_name", "nko_update_name":
		state.NKO.Name = input
		if state.State == "nko_update_name" {
			state.State = "nko_update_desc"
		} else {
			state.State = "nko_desc"
		}
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "📝 Опиши свою НКО:\n\nРасскажи, чем занимается твоя организация, какие цели и задачи она решает."))
	case "nko_desc", "nko_update_desc":
		state.NKO.Description = input
		if state.State == "nko_update_desc" {
			state.State = "nko_update_activities"
		} else {
			state.State = "nko_activities"
		}
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "🎯 Укажи формы деятельности НКО:\n\nНапример: помощь бездомным, экологические проекты, образовательные программы и т.д."))
	case "nko_activities", "nko_update_activities":
		state.NKO.Activities = input
		state.State = "nko_style"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✨ Выбери стиль постов для твоей НКО:\n\nСтиль влияет на тон и подачу контента. Выбери наиболее подходящий вариант.")
		msg.ReplyMarkup = StylesInline()
		bot.Send(msg)
	case "image_desc":
		data := map[string]interface{}{
			"desc": input,
			"nko":  state.NKO,
		}
		post, err := CallBackend("/generate_image", data)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка генерации изображения: "+err.Error()+"\n\nПопробуй ещё раз или измени описание."))
		} else {
			SendPostToUser(chatID, post, bot)
			msg := tgbotapi.NewMessage(chatID, "✨ Готово! Выбери действие с постом:")
			msg.ReplyMarkup = PostActionInline(post.PostID)
			bot.Send(msg)
		}
		ResetUserState(chatID)
	case "edit_text":
		data := map[string]interface{}{
			"text": input,
		}
		post, err := CallBackend("/edit_text", data)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка редактирования текста: "+err.Error()+"\n\nПопробуй ещё раз."))
		} else {
			// Для редактора: post.MainText — исправленный, ошибки можно в content
			bot.Send(tgbotapi.NewMessage(chatID, "✅ Текст исправлен и улучшен:\n\n"+post.MainText))
		}
		ResetUserState(chatID)
	case "text_free_input":
		// Формируем промпт на основе данных НКО и идеи
		prompt := buildPrompt("free", input, "", state.NKO, nil)
		data := map[string]interface{}{
			"prompt": prompt,
			"nko":    state.NKO,
		}
		post, err := CallBackend("/generate_text", data)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка генерации изображения: "+err.Error()+"\n\nПопробуй ещё раз или измени описание."))
		} else {
			SendPostToUser(chatID, post, bot)
			msg := tgbotapi.NewMessage(chatID, "✨ Готово! Выбери действие с постом:")
			msg.ReplyMarkup = PostActionInline(post.PostID)
			bot.Send(msg)
		}
		ResetUserState(chatID)
	case "plan_period":
		// Валидация ввода
		daysNum, err := strconv.Atoi(input)
		if err != nil || daysNum < 1 || daysNum > 365 {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Введи число от 1 до 365"))
			return
		}
		state.TempData["plan_days"] = input
		state.State = "plan_frequency"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "📊 Как часто публиковать посты?\n\nВыбери подходящую частоту публикаций для контент-плана:")
		msg.ReplyMarkup = ContentPlanFrequencyInline()
		bot.Send(msg)
		return

	// Структурированная форма генерации текста
	case "text_struct_event":
		state.TempData["event"] = input
		state.State = "text_struct_date"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "📅 Укажи дату события:\n\nНапример: 25 декабря 2024 или 25.12.2024"))
		return
	case "text_struct_date":
		state.TempData["date"] = input
		state.State = "text_struct_location"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "📍 Укажи место проведения события:\n\nНапример: Москва, концертный зал или онлайн"))
		return
	case "text_struct_location":
		state.TempData["location"] = input
		state.State = "text_struct_invited"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "👥 Кто приглашён на событие?\n\nОпиши аудиторию или спикеров. Например: известные музыканты, волонтёры, эксперты и т.д."))
		return
	case "text_struct_invited":
		state.TempData["invited"] = input
		state.State = "text_struct_details"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "📝 Дополнительные детали:\n\nУкажи программу мероприятия, условия участия, контакты и другую важную информацию."))
		return
	case "text_struct_details":
		state.TempData["details"] = input
		// Формируем промпт на основе всех собранных данных
		prompt := buildPrompt("structured", "", "", state.NKO, state.TempData)
		data := map[string]interface{}{
			"prompt": prompt,
			"nko":    state.NKO,
		}
		post, err := CallBackend("/generate_text", data)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка генерации изображения: "+err.Error()+"\n\nПопробуй ещё раз или измени описание."))
		} else {
			SendPostToUser(chatID, post, bot)
			msg := tgbotapi.NewMessage(chatID, "✨ Готово! Выбери действие с постом:")
			msg.ReplyMarkup = PostActionInline(post.PostID)
			bot.Send(msg)
		}
		ResetUserState(chatID)
		return

	// Отправка поста в чат
	case "post_send_chat":
		postID := state.TempData["post_id"]
		chatTarget := input
		data := map[string]interface{}{
			"post_id": postID,
			"chat_id": chatTarget,
		}
		_, err := CallBackend("/send_post", data)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка отправки поста: "+err.Error()+"\n\nПроверь правильность chat_id или username."))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "✅ Пост успешно отправлен в чат: "+chatTarget))
		}
		ResetUserState(chatID)
		return
	}
}

// handleCallback — обработка inline-кнопок
func handleCallback(callback *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	bot.Request(tgbotapi.NewCallback(callback.ID, "")) // подтверждение

	state := GetUserState(chatID)

	// Обработка callback'ов для НКО
	switch data {
	case "nko_yes":
		// Проверяем, это обновление или первичная настройка
		if state.NKO.Name != "" {
			// Обновление существующих данных
			state.State = "nko_update_name"
		} else {
			// Первичная настройка
			state.State = "nko_name"
		}
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "🏷️ Введи название твоей НКО:"))
		return
	case "nko_skip":
		msg := tgbotapi.NewMessage(chatID, "✅ Хорошо, будем создавать обезличенные посты.\n\nВыбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		state.State = "idle"
		SaveUserState(state)
		return
	}

	// Обработка callback'ов для режимов генерации текста
	switch data {
	case "text_free":
		state.State = "text_free_input"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "💡 Опиши идею для поста:\n\nРасскажи, о чём должен быть пост, какую информацию нужно донести до аудитории."))
		return
	case "text_struct":
		state.State = "text_struct_event"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "📌 Опиши событие (что за мероприятие, повод и т.д.):"))
		return
	}

	// Обработка callback'ов для стилей
	switch data {
	case "style_conversational":
		state.NKO.Style = "разговорный"
		state.State = "idle"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✅ Стиль постов сохранён: Разговорный\n\nТеперь все посты будут создаваться в разговорном, живом стиле. Выбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		return
	case "style_formal":
		state.NKO.Style = "официальный"
		state.State = "idle"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✅ Стиль постов сохранён: Официальный\n\nТеперь все посты будут создаваться в официальном, деловом стиле. Выбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		return
	case "style_artistic":
		state.NKO.Style = "художественный"
		state.State = "idle"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✅ Стиль постов сохранён: Художественный\n\nТеперь все посты будут создаваться в художественном, образном стиле. Выбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		return
	case "style_emotional":
		state.NKO.Style = "эмоциональный"
		state.State = "idle"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✅ Стиль постов сохранён: Эмоциональный\n\nТеперь все посты будут создаваться в эмоциональном, вдохновляющем стиле. Выбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		return
	case "style_informational":
		state.NKO.Style = "информационный"
		state.State = "idle"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✅ Стиль постов сохранён: Информационный\n\nТеперь все посты будут создаваться в информационном, новостном стиле. Выбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		return
	case "style_call_to_action":
		state.NKO.Style = "призыв к действию"
		state.State = "idle"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✅ Стиль постов сохранён: Призыв к действию\n\nТеперь все посты будут создаваться в стиле призыва к действию, мотивирующем и побуждающем. Выбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		return
	case "style_gratitude":
		state.NKO.Style = "благодарственный"
		state.State = "idle"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✅ Стиль постов сохранён: Благодарственный\n\nТеперь все посты будут создаваться в благодарственном, тёплом стиле. Выбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		return
	case "style_friendly":
		state.NKO.Style = "дружелюбный"
		state.State = "idle"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "✅ Стиль постов сохранён: Дружелюбный\n\nТеперь все посты будут создаваться в дружелюбном, неформальном стиле. Выбери функцию из меню:")
		msg.ReplyMarkup = MainMenu()
		bot.Send(msg)
		return
	}

	// Обработка callback'ов для контент-плана (выбор периода)
	switch data {
	case "plan_7":
		state.TempData["plan_days"] = "7"
		state.State = "plan_frequency"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "📊 Как часто публиковать посты?\n\nВыбери подходящую частоту публикаций для контент-плана:")
		msg.ReplyMarkup = ContentPlanFrequencyInline()
		bot.Send(msg)
		return
	case "plan_14":
		state.TempData["plan_days"] = "14"
		state.State = "plan_frequency"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "📊 Как часто публиковать посты?\n\nВыбери подходящую частоту публикаций для контент-плана:")
		msg.ReplyMarkup = ContentPlanFrequencyInline()
		bot.Send(msg)
		return
	case "plan_30":
		state.TempData["plan_days"] = "30"
		state.State = "plan_frequency"
		SaveUserState(state)
		msg := tgbotapi.NewMessage(chatID, "📊 Как часто публиковать посты?\n\nВыбери подходящую частоту публикаций для контент-плана:")
		msg.ReplyMarkup = ContentPlanFrequencyInline()
		bot.Send(msg)
		return
	case "plan_custom":
		state.State = "plan_period"
		SaveUserState(state)
		bot.Send(tgbotapi.NewMessage(chatID, "📅 На сколько дней создать контент-план?\n\nВведи число от 1 до 365:"))
		return
	}

	// Обработка callback'ов для выбора частоты публикаций
	switch data {
	case "freq_daily":
		days := state.TempData["plan_days"]
		processContentPlan(chatID, days, "ежедневно", state, bot)
		return
	case "freq_every_other":
		days := state.TempData["plan_days"]
		processContentPlan(chatID, days, "через день", state, bot)
		return
	case "freq_twice_week":
		days := state.TempData["plan_days"]
		processContentPlan(chatID, days, "2 раза в неделю", state, bot)
		return
	case "freq_thrice_week":
		days := state.TempData["plan_days"]
		processContentPlan(chatID, days, "3 раза в неделю", state, bot)
		return
	}

	// Обработка действий с постами (отправить, перегенерировать)
	if len(data) > 10 {
		prefix := data[:10]
		if prefix == "post_send_" {
			postID := data[10:]
			state.State = "post_send_chat"
			state.TempData["post_id"] = postID
			SaveUserState(state)
			bot.Send(tgbotapi.NewMessage(chatID, "📤 В какой чат отправить пост?\n\nВведи chat_id (например: -1001234567890) или username канала/группы (например: @channel_name):"))
			return
		}
		if len(data) > 15 && data[:15] == "post_regenerate_" {
			postID := data[15:]
			reqData := map[string]interface{}{
				"post_id":    postID,
				"regenerate": true,
			}
			post, err := CallBackend("/regenerate_post", reqData)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка перегенерации поста: "+err.Error()+"\n\nПопробуй ещё раз."))
			} else {
				SendPostToUser(chatID, post, bot)
				msg := tgbotapi.NewMessage(chatID, "✨ Готово! Выбери действие с постом:")
				msg.ReplyMarkup = PostActionInline(post.PostID)
				bot.Send(msg)
			}
			return
		}
	}

	// Если callback не распознан
	bot.Send(tgbotapi.NewMessage(chatID, "❓ Неизвестная команда. Выбери действие из меню:"))
}

// processContentPlan — обработка создания контент-плана
func processContentPlan(chatID int64, days string, frequency string, state *UserState, bot *tgbotapi.BotAPI) {
	data := map[string]interface{}{
		"days": days,
		"freq": frequency,
		"nko":  state.NKO,
	}
	post, err := CallBackend("/content_plan", data)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка создания контент-плана: "+err.Error()+"\n\nПопробуй ещё раз."))
	} else {
		bot.Send(tgbotapi.NewMessage(chatID, "📅 Контент-план на "+days+" дней (частота публикаций: "+frequency+"):\n\n"+post.MainText))
	}
	ResetUserState(chatID)
}

func sendHelpMessage(bot *tgbotapi.BotAPI, chatID int64) {
	helpText := `NKOshka Bot — твой SMM-менеджер для добрых дел

Что я умею?
Генерирую посты, картинки и планы — быстро и красиво

Как начать?
1️⃣ Напиши /start → расскажи о НКО
2️⃣ Выбери функцию из меню

Функции:
• Генерация текста — 3 режима
• Картинка — по описанию
• Редактор — исправляю ошибки
• Контент-план — на неделю/месяц

Совет:
Чем больше расскажешь о НКО — тем точнее посты!

Готов? Нажми кнопку ниже или напиши /start`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ReplyMarkup = MainMenu()

	bot.Send(msg)
}
