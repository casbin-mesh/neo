package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	badgerAdapter "github.com/casbin-mesh/neo/pkg/db/adapter/badger"
	"github.com/casbin-mesh/neo/pkg/neo/index"
	"github.com/casbin-mesh/neo/pkg/neo/meta"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/schema"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/y"
	"github.com/dgraph-io/ristretto/z"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockDB struct {
	db        db.DB
	metaIndex index.Index[any]
	infoIndex index.Index[*model.DBInfo]
	txnMark   y.WaterMark
	closer    *z.Closer
}

func (db *mockDB) NewTxnAt(readTs uint64, update bool) session.Context {
	txn := db.db.NewTransactionAt(readTs, update)
	metaTxn := db.metaIndex.NewTransactionAt(readTs, update)
	infoTxn := db.infoIndex.NewTransactionAt(readTs, update)
	return session.NewSessionCtx(txn, meta.NewInMemMeta(metaTxn), schema.New(infoTxn), &db.txnMark)
}

func (db *mockDB) Close() error {
	db.closer.SignalAndWait()
	return db.db.Close()
}

func OpenMockDB(t *testing.T, path string) *mockDB {
	db, err := badgerAdapter.OpenManaged(badger.DefaultOptions(path))
	assert.Nil(t, err)
	metaIndex := index.New[any](index.Options{})
	infoIndex := index.New[*model.DBInfo](index.Options{})
	closer := z.NewCloser(1)
	mark := y.WaterMark{}
	mark.Init(closer)
	return &mockDB{
		db:        db,
		metaIndex: metaIndex,
		infoIndex: infoIndex,
		txnMark:   mark,
		closer:    closer,
	}
}

func (db *mockDB) WaitForMark(ctx context.Context, ts uint64) error {
	return db.txnMark.WaitForMark(ctx, ts)
}
