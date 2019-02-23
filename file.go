package loge

import (
	"os"
	"path/filepath"
)

type fileOutputTransport struct {
	logger          *logger
	currentFilename string
	file            *os.File
}

func newFileTransport(logger *logger) *fileOutputTransport {
	return &fileOutputTransport{
		logger: logger,
	}
}

func (ft *fileOutputTransport) newTransaction(id uint64) {
}

func (ft *fileOutputTransport) stop() {

}

func (ft *fileOutputTransport) createFile() {
	if (ft.logger.configuration.Mode & OutputFileRotate) != 0 {
		ft.currentFilename = filepath.Join(ft.logger.configuration.Path, getLogName())
	} else {
		ft.currentFilename = filepath.Join(ft.logger.configuration.Path, ft.logger.configuration.Filename)
	}

	var err error
	ft.file, err = os.OpenFile(ft.currentFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		ft.file = nil
	}
}
