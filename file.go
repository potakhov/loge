package loge

import (
	"os"
	"path/filepath"
	"sync"
)

type fileOutputTransport struct {
	buffer          *buffer
	currentFilename string
	file            *os.File
	done            chan struct{}
	wg              sync.WaitGroup

	signal chan struct{}

	trans       []uint64
	transLocker sync.Mutex
}

func newFileTransport(buffer *buffer) *fileOutputTransport {
	ft := &fileOutputTransport{
		buffer: buffer,
		done:   make(chan struct{}),
		signal: make(chan struct{}, 1),
		trans:  make([]uint64, 0),
	}

	go ft.loop()
	return ft
}

func (ft *fileOutputTransport) loop() {
	ft.wg.Add(1)
	defer ft.wg.Done()

	for {
		select {
		case <-ft.done:
			ft.flushAll()
			return
		case <-ft.signal:
			ft.flushAll()
		}
	}
}

func (ft *fileOutputTransport) newTransaction(id uint64) {
	ft.transLocker.Lock()
	ft.trans = append(ft.trans, id)
	ft.transLocker.Unlock()

	select {
	case ft.signal <- struct{}{}:
	default:
	}
}

func (ft *fileOutputTransport) stop() {
	close(ft.done)
	ft.wg.Wait()
}

func (ft *fileOutputTransport) flushAll() {

}

func (ft *fileOutputTransport) createFile() {
	if (ft.buffer.logger.configuration.Mode & OutputFileRotate) != 0 {
		ft.currentFilename = filepath.Join(ft.buffer.logger.configuration.Path, getLogName())
	} else {
		ft.currentFilename = filepath.Join(ft.buffer.logger.configuration.Path, ft.buffer.logger.configuration.Filename)
	}

	var err error
	ft.file, err = os.OpenFile(ft.currentFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		ft.file = nil
	}
}
