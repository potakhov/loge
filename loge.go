package loge

import (
	"io"
	"log"
	"os"
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

// Configuration defines the logger startup configuration
type Configuration struct {
	Mode                     uint32        // work mode
	Path                     string        // output path for the file mode
	Filename                 string        // log file name (ignored if rotation is enabled)
	URLs                     []string      // server addresses
	TransactionSize          int           // transaction size limit in bytes (default 10KB)
	TransactionTimeout       time.Duration // transaction length limit (default 3 seconds)
	ConsoleOutput            io.Writer     // output writer for console (default os.Stderr)
	BacklogExpirationTimeout time.Duration // transaction backlog expiration timeout (default is time.Hour)
}

var std *logger

const (
	defaultTransactionSize   = 10 * 1024
	defaultTransactionLength = time.Second * 3
	defaultBacklogTimeout    = time.Hour
)

type logger struct {
	configuration Configuration
	buf           []byte
	buffer        *buffer
}

// Init initializes the library and returns the shutdown handler to defer
func Init(c Configuration) func() {
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

	l.buffer = newBuffer(l)

	log.SetFlags(flag)
	log.SetOutput(l)

	return l.shutdown
}

func (l *logger) shutdown() {
	l.buffer.shutdown()
}

func (l *logger) Write(d []byte) (int, error) {
	var t time.Time
	var be *BufferElement

	if ((l.configuration.Mode & OutputFile) != 0) || ((l.configuration.Mode & OutputConsole) != 0) {
		t = time.Now()
		dumpTimeToBuffer(&l.buf, t)
	}

	if (l.configuration.Mode & OutputConsole) != 0 {
		if (l.configuration.Mode & OutputConsoleInJSONFormat) != 0 {
			be = NewBufferElement(t, l.buf, d)

			json, err := be.Marshal()
			if err == nil {
				l.configuration.ConsoleOutput.Write(json)
				l.configuration.ConsoleOutput.Write([]byte("\n"))
			}
		} else {
			l.configuration.ConsoleOutput.Write(l.buf)
			l.configuration.ConsoleOutput.Write(d)
		}
	}

	if (l.configuration.Mode & OutputFile) != 0 {
		if be == nil {
			be = NewBufferElement(t, l.buf, d)
		}

		l.buffer.write(
			be,
		)
	}

	return len(d), nil
}
