package filestorage

import "github.com/vshevchenk0/kv-storage/internal/record"

type EmptyStorage struct{}

// Creates empty storage.
// Useful for tests or if you don't need a persistence layer.
func NewEmptyStorage() *EmptyStorage {
	return &EmptyStorage{}
}

func (s *EmptyStorage) Load(target map[string]*record.Record) error {
	return nil
}

func (s *EmptyStorage) Save(key string, value *record.Record) error {
	return nil
}

func (s *EmptyStorage) SaveAsync(key string, value *record.Record) {}

func (s *EmptyStorage) Delete(key string) error {
	return nil
}

func (s *EmptyStorage) DeleteAsync(key string) {}

func (s *EmptyStorage) Flush() error {
	return nil
}

func (s *EmptyStorage) Stop() error {
	return nil
}
