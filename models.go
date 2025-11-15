// models.go
package main

import "time"

// NKOData — данные об организации
type NKOData struct {
	Name        string
	Description string
	Activities  string
	Style       string
}

// UserState — состояние пользователя
type UserState struct {
	ChatID    int64
	State     string
	NKO       NKOData
	TempData  map[string]string
	UpdatedAt time.Time
}

// PostJSON — формат поста от бэкенда
type PostJSON struct {
	PostID         string  `json:"post_id"`
	PostAuthor     int64   `json:"post_author"`
	AssignedChatID []int64 `json:"assigned_chat_id"`
	MainText       string  `json:"main_text"`
	Content        []Layer `json:"content"`
}

type Layer struct {
	LayerID    string                 `json:"layer_id"`
	Type       string                 `json:"type"`
	OrderIndex int                    `json:"order_index"`
	Data       map[string]interface{} `json:"data"`
}
