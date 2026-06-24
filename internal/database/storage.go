package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type SavedServer struct {
	IP   string `json:"ip"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type Storage struct {
	mu       sync.RWMutex
	filePath string
	Data     map[int64][]SavedServer
}

// создаю хранилище и загружаю данные
func NewStorage(filePath string) (*Storage, error) {
	s := &Storage{
		filePath: filePath,
		Data:     make(map[int64][]SavedServer),
	}

	// проверяю существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll("data", 0755); err != nil {
			return nil, err
		}
		err := s.save()
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	// читаю данные из файла
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// превращаю json в map
	err = json.Unmarshal(fileData, &s.Data)
	if err != nil {
		s.Data = make(map[int64][]SavedServer)
		err = s.save()
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// добавляю сервер пользователю
func (s *Storage) AddServer(userID int64, server SavedServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data[userID] = append(s.Data[userID], server)
	return s.save()
}

// удаляю сервер по номеру
func (s *Storage) RemoveServer(userID int64, index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	servers, exists := s.Data[userID]
	if !exists {
		return fmt.Errorf("у вас нет сохраненных серверов")
	}

	if index < 0 || index >= len(servers) {
		return fmt.Errorf("сервер с номером %d не найден", index+1)
	}

	s.Data[userID] = append(servers[:index], servers[index+1:]...)

	if len(s.Data[userID]) == 0 {
		delete(s.Data, userID)
	}

	return s.save()
}

// получаю список серверов пользователя
func (s *Storage) GetServers(userID int64) []SavedServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Data[userID]
}

// сохраняю данные в json файл
func (s *Storage) save() error {
	fileData, err := json.MarshalIndent(s.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, fileData, 0644)
}