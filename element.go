package logserv

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
		Timestamp: t,
		Message:   string(msg),
	}
	copy(b.Timestring[:], buf)

	return b
}

// Marshal marshals the record into json format
func (be *BufferElement) Marshal() ([]byte, error) {
	return json.Marshal(be)
}
