package filestorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vshevchenk0/kv-storage/internal/record"
)

type JsonRecord struct {
	Key    string         `json:"key"`
	Record *record.Record `json:"record"`
}

type JsonStorage struct {
	file    *os.File
	encoder *json.Encoder
	changes map[string]*record.Record
	mu      *sync.Mutex
}

// Creates new JSON storage with AOF (append-only file) strategy.
func NewJsonStorage() (*JsonStorage, error) {
	file, err := os.OpenFile("./cache/cache.json", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("error opening cache file: %v", err)
	}
	return &JsonStorage{
		file:    file,
		encoder: json.NewEncoder(file),
		changes: make(map[string]*record.Record),
		mu:      &sync.Mutex{},
	}, nil
}

func (s *JsonStorage) Load(target map[string]*record.Record) error {
	decoder := json.NewDecoder(s.file)
	for {
		var jsonRecord JsonRecord
		if err := decoder.Decode(&jsonRecord); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if jsonRecord.Record.Value == "" {
			delete(target, jsonRecord.Key)
			continue
		}
		if time.Now().After(jsonRecord.Record.Exp) && !jsonRecord.Record.Exp.IsZero() {
			continue
		}
		target[jsonRecord.Key] = jsonRecord.Record
	}
	return nil
}

func (s *JsonStorage) Save(key string, record *record.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	jsonRecord := JsonRecord{Key: key, Record: record}
	if err := s.encoder.Encode(jsonRecord); err != nil {
		return err
	}
	delete(s.changes, key)
	return nil
}

func (s *JsonStorage) SaveAsync(key string, value *record.Record) {
	s.mu.Lock()
	s.changes[key] = value
	s.mu.Unlock()
}

func (s *JsonStorage) Delete(key string) error {
	return s.Save(key, record.NewRecord())
}

func (s *JsonStorage) DeleteAsync(key string) {
	s.mu.Lock()
	s.changes[key] = record.NewRecord()
	s.mu.Unlock()
}

func (s *JsonStorage) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, record := range s.changes {
		jsonRecord := JsonRecord{Key: key, Record: record}
		if err := s.encoder.Encode(jsonRecord); err != nil {
			return err
		}
	}
	s.changes = make(map[string]*record.Record)
	return nil
}

func (s *JsonStorage) Stop() error {
	if err := s.file.Close(); !errors.Is(err, os.ErrClosed) {
		return err
	}
	return nil
}
