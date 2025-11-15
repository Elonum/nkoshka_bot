// states.go
package main

import (
	"log"
	"sync"
	"time"
)

var (
	userStates = make(map[int64]*UserState)
	mu         sync.Mutex
)

// GetUserState — получить или создать состояние
func GetUserState(chatID int64) *UserState {
	mu.Lock()
	defer mu.Unlock()

	if state, exists := userStates[chatID]; exists {
		return state
	}

	// Создаём новое
	state := &UserState{
		ChatID:    chatID,
		State:     "idle",
		TempData:  make(map[string]string),
		UpdatedAt: time.Now(),
		NKO: NKOData{
			Name:        "",
			Description: "",
			Activities:  "",
			Style:       "",
		},
	}
	userStates[chatID] = state
	return state
}

// SaveUserState — просто обновляем время (не нужно, но для единообразия)
func SaveUserState(state *UserState) {
	mu.Lock()
	defer mu.Unlock()
	state.UpdatedAt = time.Now()
	userStates[state.ChatID] = state
}

// ResetUserState — сброс в начальное
func ResetUserState(chatID int64) {
	mu.Lock()
	defer mu.Unlock()
	delete(userStates, chatID)
}

// InitDB — теперь просто заглушка
func InitDB() {
	log.Println("In-memory state storage initialized (no SQLite)")
}
