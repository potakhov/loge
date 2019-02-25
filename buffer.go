package loge

import (
	"sync"
	"time"

	"github.com/potakhov/cache"
)

// TransactionList defines a generalized interface to the transaction list
// for various transports
type TransactionList interface {
	Get(id uint64, autofree bool) (*Transaction, bool)
	Free(id uint64)
}

// Transport defines the output transport
type Transport interface {
	NewTransaction(uint64)
	Stop()
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

	outputs  []Transport
	refcount int
}

// Transaction is a set of records to commit to the output transport
type Transaction struct {
	ID         uint64
	Items      []*BufferElement
	references int
}

func newBuffer(logger *logger) *buffer {
	return &buffer{
		nextTransactionID: 1,
		logger:            logger,
		transactionFlush:  make(chan bool, 1),
		stop:              make(chan struct{}),
		backlog:           cache.CreateLine(logger.configuration.BacklogExpirationTimeout),
	}
}

func (b *buffer) start(outputs []Transport) {
	b.outputs = outputs
	b.refcount = len(outputs)

	go b.loop()
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
		t.Stop()
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

	trans := &Transaction{
		ID:         b.nextTransactionID,
		references: b.refcount,
		Items:      tr,
	}

	b.backlogLock.Lock()
	b.backlog.Store(b.nextTransactionID, trans)
	b.backlogLock.Unlock()

	for _, t := range b.outputs {
		t.NewTransaction(b.nextTransactionID)
	}

	b.nextTransactionID++
}

// Get returns the transaction by ID. It can optionally decrease the reference count if
// caller does not need to wait for delivery confirmation
func (b *buffer) Get(id uint64, autofree bool) (*Transaction, bool) {
	b.backlogLock.Lock()
	defer b.backlogLock.Unlock()

	t, err := b.backlog.Get(id)
	if err == nil {
		trans := t.(*Transaction)

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

// Free decreases the reference count for the transaction after it has been used in the transport
func (b *buffer) Free(id uint64) {
	b.backlogLock.Lock()
	defer b.backlogLock.Unlock()

	t, err := b.backlog.Get(id)
	if err == nil {
		trans := t.(*Transaction)
		trans.references--
		if trans.references == 0 {
			b.backlog.Delete(id)
		}
	}
}
