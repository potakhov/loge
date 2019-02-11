package loge

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/potakhov/cache"
)

type buffer struct {
	logger            *logger
	operational       bool
	currentFilename   string
	file              *os.File
	stop              chan struct{}
	wg                sync.WaitGroup
	nextTransactionID uint64

	currentTransaction     []*BufferElement
	currentTransactionSize int
	currentTransactionLock sync.Mutex

	transactionFlush chan bool

	backlog     *cache.Line
	backlogLock sync.RWMutex
}

type transaction struct {
	id    uint64
	items []*BufferElement
}

func newBuffer(logger *logger) *buffer {
	b := &buffer{
		nextTransactionID: 1,
		logger:            logger,
		transactionFlush:  make(chan bool, 1),
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

	tm := time.NewTimer(b.logger.configuration.TransactionTimeout)
	for {
		select {
		case <-b.stop:
			return
		case <-b.transactionFlush:
			if !tm.Stop() {
				<-tm.C
			}
			b.flush()
			tm.Reset(b.logger.configuration.TransactionTimeout)
		case <-tm.C:
			b.flush()
			tm.Reset(b.logger.configuration.TransactionTimeout)
		}
	}
}

func (b *buffer) write(el *BufferElement) {
	var size int

	b.currentTransactionLock.Lock()
	b.currentTransaction = append(b.currentTransaction, el)
	b.currentTransactionSize += el.Size()
	size = b.currentTransactionSize
	b.currentTransactionLock.Unlock()

	if size >= b.logger.configuration.TransactionSize {
		select {
		case b.transactionFlush <- true:
		default:
		}
	}
}

func (b *buffer) shutdown() {
	if b.operational {
		close(b.stop)
		b.wg.Wait()
	}
}

func (b *buffer) flush() {
	b.currentTransactionLock.Lock()
	if len(b.currentTransaction) == 0 {
		b.currentTransactionLock.Unlock()
		return
	}

	tr := b.currentTransaction
	b.currentTransaction = make([]*BufferElement, 0)
	b.currentTransactionSize = 0
	b.currentTransactionLock.Unlock()

	trans := &transaction{
		id:    b.nextTransactionID,
		items: tr,
	}

	b.backlogLock.Lock()
	b.backlog.Store(b.nextTransactionID, trans)
	b.backlogLock.Unlock()

	// TODO notify transports about the new transaction trans
}
