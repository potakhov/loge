package loge

import (
	"sync"
)

// TransactionHandler provides simplified interface to a transaction processor
// taking care of managing goroutines and events
type TransactionHandler interface {
	WriteOutTransaction(tr *Transaction)
	FlushTransactions()
}

// WrappedTransport wraps the TransactionHandler
type WrappedTransport struct {
	buffer      TransactionList
	signal      chan struct{}
	done        chan struct{}
	trans       []uint64
	transLocker sync.Mutex
	wg          sync.WaitGroup
	terminated  bool

	handler TransactionHandler
}

// WrapTransport creates a wrapped transaction handler
func WrapTransport(buffer TransactionList, handler TransactionHandler) *WrappedTransport {
	ft := &WrappedTransport{
		buffer:  buffer,
		handler: handler,
		done:    make(chan struct{}),
		signal:  make(chan struct{}, 1),
		trans:   make([]uint64, 0),
	}

	go ft.loop()
	return ft
}

func (ft *WrappedTransport) loop() {
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

// NewTransaction /Transport handler
func (ft *WrappedTransport) NewTransaction(id uint64) {
	ft.transLocker.Lock()
	ft.trans = append(ft.trans, id)
	ft.transLocker.Unlock()

	select {
	case ft.signal <- struct{}{}:
	default:
	}
}

// Stop /Transport handler
func (ft *WrappedTransport) Stop() {
	close(ft.done)
	ft.wg.Wait()
}

func (ft *WrappedTransport) flushAll() {
	if ft.terminated {
		return
	}

	ft.transLocker.Lock()
	if len(ft.trans) == 0 {
		ft.transLocker.Unlock()
		return
	}

	ids := ft.trans
	ft.trans = make([]uint64, 0)
	ft.transLocker.Unlock()

	for _, id := range ids {
		tr, ok := ft.buffer.Get(id, true)
		if ok {
			ft.handler.WriteOutTransaction(tr)
		}
	}

	ft.handler.FlushTransactions()
}
