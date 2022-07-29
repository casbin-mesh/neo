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
	"os"
	"testing"
)

var basic_model = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

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

type asserter struct {
	dbs       []*model.DBInfo
	dbId      uint64
	tableId   uint64
	indexId   uint64
	colId     uint64
	matcherId uint64
}

func builderAsserter(infos ...*model.DBInfo) *asserter {
	a := &asserter{
		dbs: make([]*model.DBInfo, len(infos)),
	}
	for i, info := range infos {
		db := info.Clone()
		a.dbs[i] = db
		a.dbId++
		db.ID = a.dbId

		for _, matcherInfo := range db.MatcherInfo {
			a.matcherId++
			matcherInfo.ID = a.matcherId
		}

		for _, tableInfo := range db.TableInfo {
			a.tableId++
			tableInfo.ID = a.tableId

			for _, indexInfo := range tableInfo.Indices {
				a.indexId++
				indexInfo.ID = a.indexId
			}
			for _, column := range tableInfo.Columns {
				a.colId++
				column.ID = a.colId
			}
			//TODO: foreigne keys
		}

	}
	return a
}

func (a *asserter) Check(t *testing.T, sc session.Context) {
	meta := sc.GetMetaReaderWriter()
	for _, dbInfo := range a.dbs {
		id, err := meta.GetDBId(dbInfo.Name.L)
		assert.Nil(t, err)
		assert.Equal(t, dbInfo.ID, id)

		for _, matcherInfo := range dbInfo.MatcherInfo {
			id, err = meta.GetMatcherId(dbInfo.ID, matcherInfo.Name.L)
			assert.Nil(t, err)
			assert.Equal(t, matcherInfo.ID, id)
		}

		for _, tableInfo := range dbInfo.TableInfo {
			id, err = meta.GetTableId(dbInfo.ID, tableInfo.Name.L)
			assert.Nil(t, err)
			assert.Equal(t, tableInfo.ID, id)

			for _, indexInfo := range tableInfo.Indices {
				id, err = meta.GetIndexId(tableInfo.ID, indexInfo.Name.L)
				assert.Nil(t, err)
				assert.Equal(t, indexInfo.ID, id)
			}
			for _, column := range tableInfo.Columns {
				id, err = meta.GetColumnId(tableInfo.ID, column.Name.L)
				assert.Nil(t, err)
				assert.Equal(t, column.ID, id)
			}
			//TODO: foreigne keys
		}

	}
}

func TestSchemaExec_CreateDatabase(t *testing.T) {
	p := "./__test_tmp__/schema_exec"
	mockDb := OpenMockDB(t, p)
	defer func() {
		mockDb.Close()
		os.RemoveAll(p)
	}()
	sc := mockDb.NewTxnAt(1, true)
	info := &model.DBInfo{
		// ID: 1,
		Name: model.CIStr{
			O: "Test",
			L: "test",
		},
		TableInfo: []*model.TableInfo{
			{
				// ID: 1,
				Name: model.CIStr{
					O: "policy",
					L: "policy",
				},
				Columns: []*model.ColumnInfo{
					{
						// ID: 1,
						Name: model.CIStr{
							O: "subject",
							L: "subject",
						},
					},
					{
						// ID: 2,
						Name: model.CIStr{
							O: "object",
							L: "object",
						},
					},
					{
						// ID: 3,
						Name: model.CIStr{
							O: "action",
							L: "action",
						},
					},
					{
						// ID: 4,
						Name: model.CIStr{
							O: "effect",
							L: "effect",
						},
						DefaultValueBit: []byte("allow"),
					},
				},
			},
			{
				// ID: 2,
				Name: model.CIStr{
					O: "group",
					L: "group",
				},
				Columns: []*model.ColumnInfo{
					{
						// ID: 5,
						Name: model.CIStr{
							O: "member",
							L: "member",
						},
					},
					{
						// ID: 6,
						Name: model.CIStr{
							O: "group",
							L: "group",
						},
					},
					{
						// ID: 7,
						Name: model.CIStr{
							O: "domain",
							L: "domain",
						},
						DefaultValueBit: []byte("default"),
					},
				},
			},
		},
		MatcherInfo: []*model.MatcherInfo{
			{
				ID: 1,
				Name: model.CIStr{
					O: "matcher",
					L: "matcher",
				},
				Raw: basic_model,
			},
		},
	}
	checker := builderAsserter(info)

	exec := NewSchemaExec(sc)
	_, err := exec.createDatabase(context.TODO(), info)
	assert.Nil(t, err)
	err = sc.CommitTxn(context.TODO(), 2)
	assert.Nil(t, err)

	// waits all transactions finished
	assert.Nil(t, mockDb.WaitForMark(context.TODO(), 2))

	// check
	sc = mockDb.NewTxnAt(3, false)
	checker.Check(t, sc)

}