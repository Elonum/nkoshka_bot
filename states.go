// states.go
package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

var (
	userStates = make(map[int64]*UserState)
	mu         sync.Mutex
	nkoDataFile = "nko_data.json" // Файл для хранения данных НКО
	nkoData    = make(map[int64]NKOData) // Кэш данных НКО
	nkoDataMu  sync.Mutex
)

// GetUserState — получить или создать состояние
func GetUserState(chatID int64) *UserState {
	mu.Lock()
	defer mu.Unlock()

	if state, exists := userStates[chatID]; exists {
		return state
	}

	// Загружаем сохранённые данные НКО из файла
	nko := LoadNKOData(chatID)

	// Создаём новое состояние с загруженными данными НКО
	state := &UserState{
		ChatID:    chatID,
		State:     "idle",
		TempData:  make(map[string]string),
		UpdatedAt: time.Now(),
		NKO:       nko,
	}
	userStates[chatID] = state
	return state
}

// SaveUserState — сохраняем состояние и данные НКО
func SaveUserState(state *UserState) {
	mu.Lock()
	defer mu.Unlock()
	state.UpdatedAt = time.Now()
	userStates[state.ChatID] = state
	
	// Сохраняем данные НКО в файл, если они заполнены
	if state.NKO.Name != "" || state.NKO.Description != "" {
		SaveNKOData(state.ChatID, state.NKO)
	}
}

// ResetUserState — сброс состояния, но сохраняем данные НКО
func ResetUserState(chatID int64) {
	mu.Lock()
	defer mu.Unlock()
	
	// Сохраняем данные НКО перед сбросом
	if state, exists := userStates[chatID]; exists {
		if state.NKO.Name != "" || state.NKO.Description != "" {
			SaveNKOData(chatID, state.NKO)
		}
	}
	
	// Сбрасываем только состояние и временные данные, НКО данные сохраняем
	if state, exists := userStates[chatID]; exists {
		state.State = "idle"
		state.TempData = make(map[string]string)
		state.UpdatedAt = time.Now()
	} else {
		// Если состояния нет, создаём новое с сохранёнными данными НКО
		nko := LoadNKOData(chatID)
		userStates[chatID] = &UserState{
			ChatID:    chatID,
			State:     "idle",
			TempData:  make(map[string]string),
			UpdatedAt: time.Now(),
			NKO:       nko,
		}
	}
}

// SaveNKOData — сохранить данные НКО в файл
func SaveNKOData(chatID int64, nko NKOData) {
	nkoDataMu.Lock()
	defer nkoDataMu.Unlock()
	
	// Обновляем кэш
	nkoData[chatID] = nko
	
	// Загружаем существующие данные
	allData := make(map[int64]NKOData)
	if data, err := os.ReadFile(nkoDataFile); err == nil {
		json.Unmarshal(data, &allData)
	}
	
	// Обновляем данные для этого пользователя
	allData[chatID] = nko
	
	// Сохраняем в файл
	if data, err := json.MarshalIndent(allData, "", "  "); err == nil {
		os.WriteFile(nkoDataFile, data, 0644)
	} else {
		log.Printf("Ошибка сохранения данных НКО: %v", err)
	}
}

// LoadNKOData — загрузить данные НКО из файла
func LoadNKOData(chatID int64) NKOData {
	nkoDataMu.Lock()
	defer nkoDataMu.Unlock()
	
	// Проверяем кэш
	if nko, exists := nkoData[chatID]; exists {
		return nko
	}
	
	// Загружаем из файла
	if data, err := os.ReadFile(nkoDataFile); err == nil {
		allData := make(map[int64]NKOData)
		if err := json.Unmarshal(data, &allData); err == nil {
			// Обновляем кэш
			for k, v := range allData {
				nkoData[k] = v
			}
			// Возвращаем данные для этого пользователя
			if nko, exists := allData[chatID]; exists {
				return nko
			}
		}
	}
	
	// Возвращаем пустые данные, если ничего не найдено
	return NKOData{
		Name:        "",
		Description: "",
		Activities:  "",
		Style:       "",
	}
}

// InitDB — инициализация хранилища
func InitDB() {
	log.Println("In-memory state storage initialized (no SQLite)")
	
	// Создаём файл для данных НКО, если его нет
	if _, err := os.Stat(nkoDataFile); os.IsNotExist(err) {
		// Создаём пустой файл
		os.WriteFile(nkoDataFile, []byte("{}"), 0644)
		log.Printf("Создан файл для хранения данных НКО: %s", nkoDataFile)
	} else {
		// Загружаем существующие данные в кэш
		if data, err := os.ReadFile(nkoDataFile); err == nil {
			allData := make(map[int64]NKOData)
			if err := json.Unmarshal(data, &allData); err == nil {
				nkoDataMu.Lock()
				for k, v := range allData {
					nkoData[k] = v
				}
				nkoDataMu.Unlock()
				log.Printf("Загружены данные НКО для %d пользователей", len(allData))
			}
		}
	}
}
