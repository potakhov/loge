package loge

import (
	"encoding/json"
	"fmt"
	"time"
)

// BufferElement is a single buffered log entry
type BufferElement struct {
	Timestamp  time.Time                  `json:"time"`
	Timestring [dateTimeStringLength]byte `json:"-"`
	Message    string                     `json:"msg"`
	Level      string                     `json:"level,omitempty"`
	Data       map[string]interface{}     `json:"data,omitempty"`

	l *logger
}

func inPlaceBufferElement(l *logger) *BufferElement {
	return &BufferElement{
		l:    l,
		Data: make(map[string]interface{}),
	}
}

func (be *BufferElement) fill(t time.Time, buf []byte, msg []byte) {
	be.Timestamp = t.UTC()      // time in UTC for the buffer
	copy(be.Timestring[:], buf) // timestamp in local machine time for file output
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		be.Message = string(msg[:len(msg)-1]) // it is required because log.Output always adds a new line
	} else {
		be.Message = string(msg)
	}
}

func (be *BufferElement) serializeData() string {
	var serializedData string
	for key, arg := range be.Data {
		if len(serializedData) > 0 {
			serializedData += ", "
		}

		serializedData += fmt.Sprintf("%s: %v", key, arg)
	}

	if serializedData != "" {
		serializedData = "<" + serializedData + "> "
	}

	return serializedData
}

// NewBufferElement creates a new log entry
func NewBufferElement(t time.Time, buf []byte, msg []byte) *BufferElement {
	b := &BufferElement{}
	b.fill(t, buf, msg)
	return b
}

// Marshal marshals the record into json format
func (be *BufferElement) Marshal() ([]byte, error) {
	return json.Marshal(be)
}

// Size returns the record size in bytes
func (be *BufferElement) Size() int {
	// we do not count optional data fields in overall size
	// for simplicity and speed
	return dateTimeStringLength + len(be.Message)
}

// With extends the log entry with optional parameters
func (be *BufferElement) With(key string, value interface{}) *BufferElement {
	be.Data[key] = value
	return be
}

// Printf creates creates a new log entry
func (be *BufferElement) Printf(format string, v ...interface{}) {
	if be.l != nil {
		be.l.submit(be, fmt.Sprintf(format, v...))
	}
}

// Println creates creates a new log entry
func (be *BufferElement) Println(v ...interface{}) {
	if be.l != nil {
		be.l.submit(be, fmt.Sprintln(v...))
	}
}

// Infof creates creates a new "info" log entry
func (be *BufferElement) Infof(format string, v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelInfo) != 0) {
		be.Level = "info"
		be.l.submit(be, fmt.Sprintf(format, v...))
	}
}

// Infoln creates creates a new "info" log entry
func (be *BufferElement) Infoln(v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelInfo) != 0) {
		be.Level = "info"
		be.l.submit(be, fmt.Sprintln(v...))
	}
}

// Debugf creates creates a new "debug" log entry
func (be *BufferElement) Debugf(format string, v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelDebug) != 0) {
		be.Level = "debug"
		be.l.submit(be, fmt.Sprintf(format, v...))
	}
}

// Debugln creates creates a new "debug" log entry
func (be *BufferElement) Debugln(v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelDebug) != 0) {
		be.Level = "debug"
		be.l.submit(be, fmt.Sprintln(v...))
	}
}
