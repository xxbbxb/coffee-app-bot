package main

import (
	"sync"
)

type InMemoryKeyValue struct {
	storage sync.Map
	rw      sync.RWMutex
}

func (s *InMemoryKeyValue) GetKey(chatId int64, key string) interface{} {
	value, ok := s.storage.Load(chatId)
	if !ok {
		return nil
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}

	return m[key]
}

func (s *InMemoryKeyValue) SetKey(chatId int64, key string, val interface{}) {
	s.rw.Lock()
	defer s.rw.Unlock()

	m, _ := s.storage.LoadOrStore(chatId, make(map[string]interface{}))
	data := m.(map[string]interface{})

	if val != nil {
		data[key] = val
	} else {
		delete(data, key)
	}
}

func (s *InMemoryKeyValue) Clear(chatId int64) {
	s.storage.Delete(chatId)
}
