package loge

import (
	"encoding/json"
	"time"
)

// BufferElement is a single buffered log entry
type BufferElement struct {
	Timestamp  time.Time                  `json:"time"`
	Timestring [dateTimeStringLength]byte `json:"-"`
	Message    string                     `json:"msg"`
}

// NewBufferElement creates a new log entry
func NewBufferElement(t time.Time, buf []byte, msg []byte) *BufferElement {
	b := &BufferElement{
		Timestamp: t.UTC(),                  // time in UTC for the buffer
		Message:   string(msg[:len(msg)-1]), // it is required because log.Output always adds a new line
	}
	copy(b.Timestring[:], buf) // timestamp in local machine time for file output

	return b
}

// Marshal marshals the record into json format
func (be *BufferElement) Marshal() ([]byte, error) {
	return json.Marshal(be)
}

// Size returns the record size in bytes
func (be *BufferElement) Size() int {
	return dateTimeStringLength + len(be.Message)
}
