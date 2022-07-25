package index

import "sync"

type epoch[T any] struct {
	txns          []txn[T]
	txnsWriteSet  []Value[T]
	activeTxnsCnt uint64
	startTS       uint64
	endTS         uint64
	sync.Mutex    //to avoid datarace
}

type epochManager[T any] struct {
	currentEpoch      *epoch[T]
	epoches           []epoch[T]
	last_active_epoch *epoch[T]
	sync.Mutex        //to avoid datarace
}

func (em *epochManager[T]) add_into_epoch(t txn[T]) {
	em.Lock()
	defer em.Unlock()
	em.currentEpoch.txns = append(em.currentEpoch.txns, t)
	em.currentEpoch.activeTxnsCnt++
}

func (em *epochManager[T]) new_epoch() epoch[T] {
	newEp := epoch[T]{
		activeTxnsCnt: 0,
	}
	em.epoches = append(em.epoches, newEp)
	return newEp
}
