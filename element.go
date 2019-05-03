package loge

import (
	"encoding/json"
	"fmt"
	"time"
)

// BufferElement is a single buffered log entry
type BufferElement struct {
	Timestamp   time.Time                  `json:"time"`
	Timestring  [dateTimeStringLength]byte `json:"-"`
	Message     string                     `json:"msg"`
	Level       uint32                     `json:"-"`
	Levelstring string                     `json:"level,omitempty"`
	Data        map[string]interface{}     `json:"data,omitempty"`

	l *logger
}

func inPlaceBufferElement(l *logger) *BufferElement {
	return &BufferElement{
		l:    l,
		Data: make(map[string]interface{}),
	}
}

func (be *BufferElement) fill(t time.Time, buf []byte, msg []byte, level uint32) {
	be.Levelstring = levelToString(level)
	be.Level = level
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
func NewBufferElement(t time.Time, buf []byte, msg []byte, level uint32) *BufferElement {
	b := &BufferElement{}
	b.fill(t, buf, msg, level)
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
		be.l.submit(be, fmt.Sprintf(format, v...), 0)
	}
}

// Println creates creates a new log entry
func (be *BufferElement) Println(v ...interface{}) {
	if be.l != nil {
		be.l.submit(be, fmt.Sprintln(v...), 0)
	}
}

// Info creates creates a new "info" log entry
func (be *BufferElement) Info(format string, v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelInfo) != 0) {
		be.l.submit(be, fmt.Sprintf(format, v...), LogLevelInfo)
	}
}

// Debug creates creates a new "debug" log entry
func (be *BufferElement) Debug(format string, v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelDebug) != 0) {
		be.l.submit(be, fmt.Sprintf(format, v...), LogLevelDebug)
	}
}

// Trace creates creates a new "trace" log entry
func (be *BufferElement) Trace(format string, v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelTrace) != 0) {
		be.l.submit(be, fmt.Sprintf(format, v...), LogLevelTrace)
	}
}

// Warn creates creates a new "warning" log entry
func (be *BufferElement) Warn(format string, v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelWarning) != 0) {
		be.l.submit(be, fmt.Sprintf(format, v...), LogLevelWarning)
	}
}

// Error creates creates a new "error" log entry
func (be *BufferElement) Error(format string, v ...interface{}) {
	if (be.l != nil) && ((be.l.configuration.LogLevels & LogLevelError) != 0) {
		be.l.submit(be, fmt.Sprintf(format, v...), LogLevelError)
	}
}
