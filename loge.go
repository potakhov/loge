package loge

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Various log output modes
const (
	OutputConsole             uint32 = 1  // OutputConsole outputs to the stderr
	OutputFile                uint32 = 2  // OutputFile adds a file output
	OutputFileRotate          uint32 = 4  // OutputFileRotate adds an automatic file rotation based on current date
	OutputIncludeLine         uint32 = 8  // Include file and line into the output
	OutputConsoleInJSONFormat uint32 = 16 // Switch console output to JSON serialized format
)

// Various selectable log levels
const (
	LogLevelInfo  uint32 = 1
	LogLevelDebug uint32 = 2
)

// TransportCreator is an interface to create new optional transports when the log is initialized
type TransportCreator func(TransactionList) []Transport

// Configuration defines the logger startup configuration
type Configuration struct {
	Mode                     uint32           // work mode
	Path                     string           // output path for the file mode
	Filename                 string           // log file name (ignored if rotation is enabled)
	TransactionSize          int              // transaction size limit in bytes (default 10KB)
	TransactionTimeout       time.Duration    // transaction length limit (default 3 seconds)
	ConsoleOutput            io.Writer        // output writer for console (default os.Stderr)
	BacklogExpirationTimeout time.Duration    // transaction backlog expiration timeout (default is time.Hour)
	LogLevels                uint32           // selectable log levels
	Transports               TransportCreator // Optional transports creator
}

var std *logger

func init() {
	std = newLogger(
		Configuration{
			Mode:          OutputConsole,
			ConsoleOutput: os.Stderr,
		})
}

const (
	defaultTransactionSize   = 10 * 1024
	defaultTransactionLength = time.Second * 3
	defaultBacklogTimeout    = time.Minute * 15
)

type logger struct {
	configuration        Configuration
	writeTimestampBuffer []byte
	buffer               *buffer

	customTimestampBuffer []byte
	customTimestampLock   sync.Mutex
}

// Init initializes the library and returns the shutdown handler to defer
func Init(c Configuration) func() {
	std = newLogger(c)
	return std.shutdown
}

func newLogger(c Configuration) *logger {
	l := &logger{
		configuration: c,
	}

	flag := 0
	if (c.Mode & OutputIncludeLine) != 0 {
		flag |= log.Lshortfile
	}

	if (c.Mode & OutputFile) != 0 {
		validPath := false

		if fileInfo, err := os.Stat(c.Path); !os.IsNotExist(err) {
			if fileInfo.IsDir() {
				validPath = true
			}
		}

		if !validPath {
			l.configuration.Mode = l.configuration.Mode & (^OutputFile)
			os.Stderr.Write([]byte("Log path is invalid.  Log file output is disabled.\n"))
		}
	}

	if l.configuration.TransactionSize == 0 {
		l.configuration.TransactionSize = defaultTransactionSize
	}

	if l.configuration.TransactionTimeout == 0 {
		l.configuration.TransactionTimeout = defaultTransactionLength
	}

	if l.configuration.ConsoleOutput == nil {
		l.configuration.ConsoleOutput = os.Stderr
	}

	if l.configuration.BacklogExpirationTimeout == 0 {
		l.configuration.BacklogExpirationTimeout = defaultBacklogTimeout
	}

	if ((l.configuration.Mode & OutputFile) != 0) || (l.configuration.Transports != nil) {
		buffer := newBuffer(l)

		var outputs []Transport

		if (l.configuration.Mode & OutputFile) != 0 {
			outputs = make([]Transport, 1)
			outputs[0] = newFileTransport(buffer, c.Path, c.Filename, (c.Mode&OutputFileRotate) != 0, (c.Mode&OutputConsoleInJSONFormat) != 0)
		} else {
			outputs = make([]Transport, 0)
		}

		if l.configuration.Transports != nil {
			outputs = append(outputs, l.configuration.Transports(buffer)...)
		}

		if len(outputs) > 0 {
			l.buffer = buffer
			l.buffer.start(outputs)
		}
	}

	log.SetFlags(flag)
	log.SetOutput(l)

	return l
}

func (l *logger) shutdown() {
	if l.buffer != nil {
		l.buffer.shutdown()
	}
}

func (l *logger) Write(d []byte) (int, error) {
	if (l.buffer != nil) || ((l.configuration.Mode & OutputConsole) != 0) {
		t := time.Now()
		dumpTimeToBuffer(&l.writeTimestampBuffer, t) // don't have to lock this buf here because Write events are serialized
		l.write(
			NewBufferElement(t, l.writeTimestampBuffer, d),
		)
	}

	return len(d), nil
}

func (l *logger) write(be *BufferElement) {
	if (l.configuration.Mode & OutputConsole) != 0 {
		if (l.configuration.Mode & OutputConsoleInJSONFormat) != 0 {
			json, err := be.Marshal()
			if err == nil {
				l.configuration.ConsoleOutput.Write(json)
				l.configuration.ConsoleOutput.Write([]byte("\n"))
			}
		} else {
			l.configuration.ConsoleOutput.Write(be.Timestring[:])
			if be.Data != nil {
				l.configuration.ConsoleOutput.Write([]byte(be.serializeData()))
			}
			l.configuration.ConsoleOutput.Write([]byte(be.Message))
			l.configuration.ConsoleOutput.Write([]byte("\n"))
		}
	}

	if l.buffer != nil {
		l.buffer.write(
			be,
		)
	}
}

func (l *logger) writeLevel(level uint32, message string) {
	if (l.buffer != nil) || ((l.configuration.Mode & OutputConsole) != 0) {
		l.customTimestampLock.Lock()
		defer l.customTimestampLock.Unlock()
		t := time.Now()
		dumpTimeToBuffer(&l.customTimestampBuffer, t)
		be := NewBufferElement(t, l.customTimestampBuffer, []byte(message))
		switch level {
		case LogLevelInfo:
			be.Level = "info"
		case LogLevelDebug:
			be.Level = "debug"
		}
		l.write(be)
	}
}

// Printf creates creates a new log entry
func Printf(format string, v ...interface{}) {
	std.writeLevel(0, fmt.Sprintf(format, v...))
}

// Println creates creates a new log entry
func Println(v ...interface{}) {
	std.writeLevel(0, fmt.Sprintln(v...))
}

// Infof creates creates a new "info" log entry
func Infof(format string, v ...interface{}) {
	if (std.configuration.LogLevels & LogLevelInfo) != 0 {
		std.writeLevel(LogLevelInfo, fmt.Sprintf(format, v...))
	}
}

// Infoln creates creates a new "info" log entry
func Infoln(v ...interface{}) {
	if (std.configuration.LogLevels & LogLevelInfo) != 0 {
		std.writeLevel(LogLevelInfo, fmt.Sprintln(v...))
	}
}

// Debugf creates creates a new "debug" log entry
func Debugf(format string, v ...interface{}) {
	if (std.configuration.LogLevels & LogLevelDebug) != 0 {
		std.writeLevel(LogLevelDebug, fmt.Sprintf(format, v...))
	}
}

// Debugln creates creates a new "debug" log entry
func Debugln(v ...interface{}) {
	if (std.configuration.LogLevels & LogLevelDebug) != 0 {
		std.writeLevel(LogLevelDebug, fmt.Sprintln(v...))
	}
}

// With creates a new log entry with optional parameters
func With(key string, value interface{}) *BufferElement {
	be := inPlaceBufferElement(std)
	be.Data[key] = value
	return be
}

func (l *logger) submit(be *BufferElement, message string) {
	if (l.buffer != nil) || ((l.configuration.Mode & OutputConsole) != 0) {
		l.customTimestampLock.Lock()
		defer l.customTimestampLock.Unlock()
		t := time.Now()
		dumpTimeToBuffer(&l.customTimestampBuffer, t)
		be.fill(t, l.customTimestampBuffer, []byte(message))
		l.write(be)
	}
}
