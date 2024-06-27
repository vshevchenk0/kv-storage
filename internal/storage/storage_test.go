package storage

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/vshevchenk0/kv-storage/internal/filestorage"
	"github.com/vshevchenk0/kv-storage/internal/record"
)

var s *storage

func TestMain(m *testing.M) {
	s = &storage{
		store:       make(map[string]*record.Record),
		fileStorage: filestorage.NewEmptyStorage(),
		mu:          &sync.RWMutex{},
	}
	m.Run()
}

func TestSetExpiringRecord(t *testing.T) {
	cases := []struct {
		name        string
		key         string
		value       string
		ttl         uint64
		syncCommit  bool
		expectedTtl time.Duration
	}{
		{
			name:        "should set expiring record to storage with sync commit",
			key:         "expiring",
			value:       "expiring",
			ttl:         60,
			syncCommit:  true,
			expectedTtl: time.Minute * 1,
		},
		{
			name:        "should extend record ttl with async commit",
			key:         "expiring",
			value:       "expiring",
			ttl:         60,
			syncCommit:  false,
			expectedTtl: time.Minute * 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := s.Set(tc.key, tc.value, tc.ttl, tc.syncCommit); err != nil {
				t.Errorf("error setting record: %v", err)
			}
			record, ok := s.store[tc.key]
			if !ok {
				t.Error("record was not set")
			}
			if record.Value != tc.value {
				t.Errorf("record has wrong value. has: %v, want: %v", record.Value, tc.value)
			}
			ttl := time.Until(record.Exp).Round(time.Minute)
			if ttl != tc.expectedTtl {
				t.Errorf("record has wrong ttl. has: %s, want :%s", ttl, tc.expectedTtl)
			}
		})
	}
}

func TestSetNonExpiringRecord(t *testing.T) {
	key := "non-expiring"
	value := "non-expiring"
	ttl := uint64(0)
	syncCommit := false
	if err := s.Set(key, value, ttl, syncCommit); err != nil {
		t.Errorf("error setting record: %v", err)
	}
	record, ok := s.store[key]
	if !ok {
		t.Error("record was not set")
	}
	if record.Value != value {
		t.Errorf("record has wrong value. has: %v, want: %v", record.Value, value)
	}
	if !record.Exp.IsZero() {
		t.Errorf("record should be non-expiring but has expiration date")
	}
}

func TestSetConvertNonExpiringRecordToExpiring(t *testing.T) {
	key := "non-expiring"
	value := "expiring"
	ttl := uint64(60)
	expectedTtl := time.Minute * 1
	syncCommit := true
	if err := s.Set(key, value, ttl, syncCommit); err != nil {
		t.Errorf("error setting record: %v", err)
	}
	record, ok := s.store[key]
	if !ok {
		t.Error("record was not set")
	}
	if record.Value != value {
		t.Errorf("record has wrong value. has: %v, want: %v", record.Value, value)
	}
	if record.Exp.IsZero() {
		t.Errorf("record should be expiring but has no expiration date")
	}
	actualTtl := time.Until(record.Exp).Round(time.Minute)
	if actualTtl != expectedTtl {
		t.Errorf("record has wrong ttl. has: %s, want :%s", actualTtl, expectedTtl)
	}
}

func TestGetValue(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
		ttl   uint64
	}{
		{
			name:  "should return value of expiring record",
			key:   "get_expiring",
			value: "get_expiring",
			ttl:   uint64(1),
		},
		{
			name:  "should return value of non-expiring record",
			key:   "get_non-expiring",
			value: "get_non-expiring",
			ttl:   uint64(0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := s.Set(tc.key, tc.value, tc.ttl, true); err != nil {
				t.Errorf("error setting record: %v", err)
			}
			value, err := s.Get(tc.key)
			if err != nil {
				t.Errorf("error getting record: %v", err)
			}
			if value != tc.value {
				t.Errorf("record has wrong value. has: %v, want: %v", value, tc.value)
			}
		})
	}
}

func TestGetExpiredRecord(t *testing.T) {
	key := "expired"
	value := "expired"
	ttl := uint64(1)
	syncCommit := true
	if err := s.Set(key, value, ttl, syncCommit); err != nil {
		t.Errorf("error setting record: %v", err)
	}
	<-time.After(time.Second * 1)
	_, err := s.Get(key)
	if err == nil {
		t.Errorf("should receive error")
	}
	if !errors.Is(err, ErrRecordExpired) {
		t.Errorf("received wrong error: %v", err)
	}
}

func TestGetNonExistingRecord(t *testing.T) {
	_, err := s.Get("non-existing")
	if err == nil {
		t.Errorf("should receive error")
	}
	if !errors.Is(err, ErrRecordDoesNotExist) {
		t.Errorf("received wrong error: %v", err)
	}
}

func TestDelete(t *testing.T) {
	cases := []struct {
		name       string
		key        string
		value      string
		ttl        uint64
		syncCommit bool
	}{
		{
			name:       "should delete record with sync commit",
			key:        "delete_me_sync",
			value:      "delete_me_sync",
			ttl:        uint64(0),
			syncCommit: true,
		},
		{
			name:       "should delete record with async commit",
			key:        "delete_me_async",
			value:      "delete_me_async",
			ttl:        uint64(2),
			syncCommit: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := s.Set(tc.key, tc.value, tc.ttl, true); err != nil {
				t.Errorf("error setting record: %v", err)
			}
			if err := s.Delete(tc.key, tc.syncCommit); err != nil {
				t.Errorf("error deleting record: %v", err)
			}
			if _, ok := s.store[tc.key]; ok {
				t.Errorf("delete was successful, but record is still in store")
			}
		})
	}
}
