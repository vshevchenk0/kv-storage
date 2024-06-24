package storage

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/vshevchenk0/kv-storage/internal/filestorage"
	"github.com/vshevchenk0/kv-storage/internal/record"
)

var (
	ErrRecordDoesNotExist = errors.New("record does not exist")
	ErrRecordExpired      = errors.New("record expired")
)

type Storage interface {
	Set(key string, value string, ttl uint64, syncCommit bool) error
	Get(key string) (string, error)
	Delete(key string, syncCommit bool) error
	Stop() error
}

type storage struct {
	logger      *slog.Logger
	store       map[string]*record.Record
	fileStorage filestorage.FileStorage
	mu          *sync.RWMutex
	ticker      *time.Ticker
}

func NewStorage(logger *slog.Logger, fileStorage filestorage.FileStorage, flushFrequency time.Duration) *storage {
	storage := &storage{
		logger:      logger,
		store:       make(map[string]*record.Record),
		fileStorage: fileStorage,
		mu:          &sync.RWMutex{},
		ticker:      time.NewTicker(flushFrequency),
	}
	storage.init()
	return storage
}

func (s *storage) init() {
	if err := s.fileStorage.Load(s.store); err != nil {
		s.logger.Warn("error reading from file, some data might be missing", slog.String("error", err.Error()))
	}
	go s.scheduleFlush()
}

func (s *storage) scheduleFlush() {
	for range s.ticker.C {
		s.fileStorage.Flush()
	}
}

func (s *storage) Stop() error {
	s.ticker.Stop()
	s.fileStorage.Flush()
	return s.fileStorage.Stop()
}

func (s *storage) Set(key string, value string, ttl uint64, syncCommit bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := record.NewRecord()
	record.Value = value
	existingRecord, ok := s.store[key]
	if !ok {
		if ttl != 0 {
			record.Exp = time.Now().Add(time.Duration(uint64(time.Second) * ttl))
		}
	} else {
		if existingRecord.Exp.IsZero() {
			record.Exp = time.Now().Add(time.Duration(uint64(time.Second) * ttl))
		} else {
			record.Exp = existingRecord.Exp.Add(time.Duration(uint64(time.Second) * ttl))
		}
	}

	if syncCommit {
		if err := s.fileStorage.Save(key, record); err != nil {
			return err
		}
	} else {
		s.fileStorage.SaveAsync(key, record)
	}
	s.store[key] = record
	return nil
}

func (s *storage) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.store[key]
	if !ok {
		return "", ErrRecordDoesNotExist
	}
	if record.Exp.IsZero() || time.Now().Before(record.Exp) {
		return record.Value, nil
	}
	delete(s.store, key)

	return "", ErrRecordExpired
}

func (s *storage) Delete(key string, syncCommit bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if syncCommit {
		if err := s.fileStorage.Delete(key); err != nil {
			return err
		}
	} else {
		s.fileStorage.DeleteAsync(key)
	}
	delete(s.store, key)
	return nil
}
