// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"context"
	"github.com/dgraph-io/badger/v3/y"
	"github.com/dgraph-io/ristretto/z"
	"sync"
)

type oracle struct {
	txnMark  *y.WaterMark
	readMark *y.WaterMark
	closer   *z.Closer

	sync.Mutex // For nextTxnTs and commits.

	setDiscard func(ts uint64)
	nextTxnTs  uint64
}

func newOracle() *oracle {
	orc := &oracle{
		readMark: &y.WaterMark{Name: ".PendingReads"},
		txnMark:  &y.WaterMark{Name: ".TxnTimestamp"},
		closer:   z.NewCloser(2),
	}
	orc.readMark.Init(orc.closer)
	orc.txnMark.Init(orc.closer)
	return orc
}

func (o *oracle) Stop() {
	o.closer.SignalAndWait()
}

func (o *oracle) readTs() uint64 {
	var readTs uint64
	o.Lock()
	readTs = o.nextTxnTs - 1
	o.readMark.Begin(readTs)
	o.Unlock()

	// Wait for all txns which have no conflicts, have been assigned a commit
	// timestamp and are going through the write to value log and LSM tree
	// process. Not waiting here could mean that some txns which have been
	// committed would not be read.
	y.Check(o.txnMark.WaitForMark(context.Background(), readTs))
	return readTs
}

func (o *oracle) nextTs() uint64 {
	o.Lock()
	defer o.Unlock()
	return o.nextTxnTs
}

func (o *oracle) incrementNextTs() {
	o.Lock()
	defer o.Unlock()
	o.nextTxnTs++
}

func (o *oracle) newCommitTs(readTs uint64) (commitTs uint64) {
	o.Lock()
	defer o.Unlock()
	o.readMark.Done(readTs)

	// targets cleanup committed transactions
	if o.setDiscard != nil {
		o.setDiscard(o.readMark.DoneUntil())
	}

	commitTs = o.nextTxnTs
	o.nextTxnTs++
	o.txnMark.Begin(commitTs)
	return commitTs
}

func (o *oracle) DoneRead(readTs uint64) {
	o.readMark.Done(readTs)
}

func (o *oracle) DoneCommit(commitTs uint64) {
	o.txnMark.Done(commitTs)
}

func (o *oracle) IncNextTs() {
	o.incrementNextTs()
}
