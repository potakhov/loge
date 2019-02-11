package logserv

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type buffer struct {
	logger          *logger
	operational     bool
	currentFilename string
	file            *os.File
	stop            chan struct{}
	wg              sync.WaitGroup
}

func newBuffer(logger *logger) *buffer {
	b := &buffer{
		logger: logger,
	}

	if (b.logger.configuration.Mode & OutputFile) != 0 {
		b.operational = true
		b.stop = make(chan struct{})
		go b.loop()
	}

	return b
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

func (b *buffer) loop() {
	b.wg.Add(1)
	defer b.wg.Done()

	for {
		select {
		case <-b.stop:
			return
		}
	}
}

func (b *buffer) write(el *BufferElement) {
	m, _ := el.Marshal()
	fmt.Printf("%s", string(m))
}

func (b *buffer) shutdown() {
	if b.operational {
		close(b.stop)
		b.wg.Wait()
	}
}
