package index

import "sync"

type TSController struct {
	nextTxnTS  uint64
	sync.Mutex //to protect nextTxnTS from datarace
}

func (TSC *TSController) TS_init() {
	TSC.Lock()
	defer TSC.Unlock()
	TSC.nextTxnTS = 1
}

func (TSC *TSController) get_TS() uint64 {
	TSC.Lock()
	defer TSC.Unlock()
	ts := TSC.nextTxnTS
	TSC.nextTxnTS++
	return ts
}
