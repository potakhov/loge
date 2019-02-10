package logserv

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Various log output modes
const (
	OutputConsole     = 1  // OutputConsole outputs to the stderr
	OutputFile        = 2  // OutputFile adds a file output
	OutputFileRotate  = 4  // OutputFileRotate adds an automatic file rotation based on current date
	OutputIncludeFile = 8  // Include file and line into the output
	OutputServer      = 16 // OutputServer starts the server output
)

// Configuration defines the logger startup configuration
type Configuration struct {
	Mode     int      // work mode
	Path     string   // output path for the file mode
	Filename string   // log file name (ignored if rotation is enabled)
	URLs     []string // server addresses
}

var std *logger

type logger struct {
	configuration   Configuration
	writingFile     bool
	currentFilename string
	file            *os.File
	buf             []byte
}

// Init initializes the library and returns the shutdown handler to defer
func Init(c Configuration) func() {
	l := &logger{
		configuration: c,
	}

	flag := 0
	if (c.Mode & OutputIncludeFile) != 0 {
		flag |= log.Lshortfile
	}

	if (c.Mode & OutputFile) != 0 {
		if fileInfo, err := os.Stat(c.Path); !os.IsNotExist(err) {
			if fileInfo.IsDir() {
				l.writingFile = true
			}
		}
	}

	log.SetFlags(flag)
	log.SetOutput(l)

	return l.shutdown
}

func (l *logger) shutdown() {
	if l.writingFile {
		l.file.Close()
	}
}

func (l *logger) Write(d []byte) (int, error) {
	var t time.Time

	if l.writingFile || ((l.configuration.Mode & OutputConsole) != 0) {
		t = time.Now()
		l.dumpTimeToBuffer(t)
	}

	if (l.configuration.Mode & OutputConsole) != 0 {
		os.Stderr.Write(l.buf)
		os.Stderr.Write(d)
	}

	/*if l.writingFile {
		l.file.Write(l.buf)
		l.file.Write(d)
	}*/

	return len(d), nil
}

func (l *logger) getLogName() string {
	t := time.Now()
	ret := fmt.Sprintf("%d%02d%02d.log", t.Year(), t.Month(), t.Day())
	return ret
}

func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *logger) dumpTimeToBuffer(t time.Time) {
	l.buf = l.buf[:0]
	year, month, day := t.Date()
	itoa(&l.buf, year, 4)
	l.buf = append(l.buf, '/')
	itoa(&l.buf, int(month), 2)
	l.buf = append(l.buf, '/')
	itoa(&l.buf, day, 2)
	l.buf = append(l.buf, ' ')

	hour, min, sec := t.Clock()
	itoa(&l.buf, hour, 2)
	l.buf = append(l.buf, ':')
	itoa(&l.buf, min, 2)
	l.buf = append(l.buf, ':')
	itoa(&l.buf, sec, 2)
	l.buf = append(l.buf, '.')
	itoa(&l.buf, t.Nanosecond()/1e3, 6)
	l.buf = append(l.buf, ' ')
}

func (l *logger) createFile() {
	if (l.configuration.Mode & OutputFileRotate) != 0 {
		l.currentFilename = filepath.Join(l.configuration.Path, l.getLogName())
	} else {
		l.currentFilename = filepath.Join(l.configuration.Path, l.configuration.Filename)
	}

	var err error
	l.file, err = os.OpenFile(l.currentFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		l.file = nil
	}
}
