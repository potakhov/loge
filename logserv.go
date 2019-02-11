package logserv

import (
	"log"
	"os"
	"time"
)

// Various log output modes
const (
	OutputConsole     uint32 = 1 // OutputConsole outputs to the stderr
	OutputFile        uint32 = 2 // OutputFile adds a file output
	OutputFileRotate  uint32 = 4 // OutputFileRotate adds an automatic file rotation based on current date
	OutputIncludeLine uint32 = 8 // Include file and line into the output
)

// Configuration defines the logger startup configuration
type Configuration struct {
	Mode               uint32        // work mode
	Path               string        // output path for the file mode
	Filename           string        // log file name (ignored if rotation is enabled)
	URLs               []string      // server addresses
	TransactionSize    int           // transaction size limit in bytes (default 10KB)
	TransactionTimeout time.Duration // transaction length limit (default 3 seconds)
}

var std *logger

const (
	defaultTransactionSize   = 10 * 1024
	defaultTransactionLength = time.Second * 3
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

	if ((l.configuration.Mode & OutputFile) != 0) || ((l.configuration.Mode & OutputConsole) != 0) {
		t = time.Now()
		dumpTimeToBuffer(&l.buf, t)
	}

	if (l.configuration.Mode & OutputConsole) != 0 {
		os.Stderr.Write(l.buf)
		os.Stderr.Write(d)
	}

	if (l.configuration.Mode & OutputFile) != 0 {
		l.buffer.write(
			NewBufferElement(t, l.buf, d),
		)
	}

	return len(d), nil
}
