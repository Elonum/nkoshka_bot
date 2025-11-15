package main

import (
	"log"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
	mu sync.Mutex
)

// InitDB — инициализация БД для состояний
func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("states.db"), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}
	DB.AutoMigrate(&UserState{})
}

// GetUserState — получить состояние по chatID (или создать новое)
func GetUserState(chatID int64) *UserState {
	mu.Lock()
	defer mu.Unlock()

	state := &UserState{ChatID: chatID}
	if err := DB.Where("chat_id = ?", chatID).First(state).Error; err != nil {
		// Создаём новое, если не найдено
		state.State = "idle"
		state.TempData = make(map[string]string)
		DB.Create(state)
	}
	return state
}

// SaveUserState — сохранить состояние
func SaveUserState(state *UserState) {
	mu.Lock()
	defer mu.Unlock()
	DB.Save(state)
}

// ResetUserState — сброс в idle
func ResetUserState(chatID int64) {
	state := GetUserState(chatID)
	state.State = "idle"
	state.TempData = make(map[string]string)
	SaveUserState(state)
}
