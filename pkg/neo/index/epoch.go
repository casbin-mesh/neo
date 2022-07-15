package index

import "sync"

type epoch[T any] struct {
	txns          []txn
	txnsWriteSet  []Value
	activeTxnsCnt uint64
	startTS       uint64
	endTS         uint64
	sync.Mutex    //to avoid datarace
}

type epochManager[T any] struct {
	currentEpoch      epoch
	epoches           []epoch
	last_active_epoch *epoch
	sync.Mutex        //to avoid datarace
}

func (em *epochManager) add_into_epoch(t txn) {
	em.Lock()
	defer em.Unlock()
	em.currentEpoch.txns = append(em.currentEpoch.txns, t)
	em.currentEpoch.activeTxnsCnt++
}

func (em *epochManager) new_epoch() epoch {
	newEp := epoch{
		activeTxnsCnt: 0,
	}
	em.epoches = append(em.epoches, newEp)
	return newEp
}
