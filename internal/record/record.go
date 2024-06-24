package record

import "time"

type Record struct {
	Value string
	Exp   time.Time
}

// Returns new empty record
func NewRecord() *Record {
	return &Record{}
}
