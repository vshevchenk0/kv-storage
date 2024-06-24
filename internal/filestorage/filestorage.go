package filestorage

import "github.com/vshevchenk0/kv-storage/internal/record"

type FileStorage interface {
	Load(target map[string]*record.Record) error
	Save(key string, value *record.Record) error
	SaveAsync(key string, record *record.Record)
	Delete(key string) error
	DeleteAsync(key string)
	Flush() error
	Stop() error
}
