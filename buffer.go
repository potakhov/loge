package loge

import (
	"sync"
	"time"

	"github.com/potakhov/cache"
)

type transport interface {
	newTransaction(uint64)
	stop()
}

type buffer struct {
	logger            *logger
	stop              chan struct{}
	wg                sync.WaitGroup
	nextTransactionID uint64

	currentTransaction     []*BufferElement
	currentTransactionSize int
	currentTransactionLock sync.Mutex

	transactionFlush chan bool
	flushSent        bool

	backlog     *cache.Line
	backlogLock sync.Mutex

	outputs  []transport
	refcount int
}

type transaction struct {
	id         uint64
	references int
	items      []*BufferElement
}

func newBuffer(logger *logger) *buffer {
	b := &buffer{
		nextTransactionID: 1,
		logger:            logger,
		transactionFlush:  make(chan bool, 1),
		stop:              make(chan struct{}),
		backlog:           cache.CreateLine(logger.configuration.BacklogExpirationTimeout),
	}

	b.outputs = make([]transport, 1)
	b.outputs[0] = newFileTransport(b)
	b.refcount = 1

	go b.loop()

	return b
}

func (b *buffer) loop() {
	b.wg.Add(1)
	defer b.wg.Done()

	tm := time.NewTimer(b.logger.configuration.TransactionTimeout)
	for {
		select {
		case <-b.stop:
			b.flush()
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
	flush := false

	b.currentTransactionLock.Lock()
	b.currentTransaction = append(b.currentTransaction, el)
	b.currentTransactionSize += el.Size()
	if !b.flushSent {
		if b.currentTransactionSize >= b.logger.configuration.TransactionSize {
			flush = true
			b.flushSent = true
		}
	}
	b.currentTransactionLock.Unlock()

	if flush {
		select {
		case b.transactionFlush <- true:
		default:
		}
	}
}

func (b *buffer) shutdown() {
	close(b.stop)
	b.wg.Wait()

	for _, t := range b.outputs {
		t.stop()
	}
}

func (b *buffer) flush() {
	b.currentTransactionLock.Lock()
	b.flushSent = false
	if len(b.currentTransaction) == 0 {
		b.currentTransactionLock.Unlock()
		return
	}

	tr := b.currentTransaction
	b.currentTransaction = make([]*BufferElement, 0)
	b.currentTransactionSize = 0
	b.currentTransactionLock.Unlock()

	trans := &transaction{
		id:         b.nextTransactionID,
		references: b.refcount,
		items:      tr,
	}

	b.backlogLock.Lock()
	b.backlog.Store(b.nextTransactionID, trans)
	b.backlogLock.Unlock()

	for _, t := range b.outputs {
		t.newTransaction(b.nextTransactionID)
	}

	b.nextTransactionID++
}

func (b *buffer) get(id uint64, autofree bool) (*transaction, bool) {
	b.backlogLock.Lock()
	defer b.backlogLock.Unlock()

	t, err := b.backlog.Get(id)
	if err == nil {
		trans := t.(*transaction)

		if autofree {
			trans.references--
			if trans.references == 0 {
				b.backlog.Delete(id)
			}
		}

		return trans, true
	}

	return nil, false
}

func (b *buffer) free(id uint64) {
	b.backlogLock.Lock()
	defer b.backlogLock.Unlock()

	t, err := b.backlog.Get(id)
	if err == nil {
		trans := t.(*transaction)
		trans.references--
		if trans.references == 0 {
			b.backlog.Delete(id)
		}
	}
}
