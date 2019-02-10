package logserv

import (
	"os"
	"path/filepath"
	"time"
)

type buffer struct {
	logger          *logger
	currentFilename string
	file            *os.File
}

type bufferElement struct {
	timestamp  time.Time
	timestring []byte
	message    []byte
}

func newBufferElement(t time.Time, buf []byte, msg []byte) *bufferElement {
	b := &bufferElement{
		timestamp: t,
	}

	b.timestring = make([]byte, len(buf))
	copy(b.timestring, buf)
	b.message = make([]byte, len(msg))
	copy(b.message, msg)

	return b
}

func newBuffer(logger *logger) *buffer {
	return &buffer{
		logger: logger,
	}
}

func (b *buffer) createFile() {
	if (b.logger.configuration.Mode & OutputFileRotate) != 0 {
		b.currentFilename = filepath.Join(b.logger.configuration.Path, getLogName())
	} else {
		b.currentFilename = filepath.Join(b.logger.configuration.Path, b.logger.configuration.Filename)
	}

	var err error
	b.file, err = os.OpenFile(b.currentFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		b.file = nil
	}
}

func (b *buffer) write(el *bufferElement) {

}

func (b *buffer) shutdown() {

}
