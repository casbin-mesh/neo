package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	badgerAdapter "github.com/casbin-mesh/neo/pkg/db/adapter/badger"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/index"
	"github.com/casbin-mesh/neo/pkg/neo/meta"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/schema"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"github.com/casbin-mesh/neo/pkg/utils"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/y"
	"github.com/dgraph-io/ristretto/z"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	basicModelText = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

	mockDBInfo1 = &model.DBInfo{
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
						ColName: model.CIStr{
							O: "subject",
							L: "subject",
						},
					},
					{
						// ID: 2,
						ColName: model.CIStr{
							O: "object",
							L: "object",
						},
					},
					{
						// ID: 3,
						ColName: model.CIStr{
							O: "action",
							L: "action",
						},
					},
					{
						// ID: 4,
						ColName: model.CIStr{
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
						ColName: model.CIStr{
							O: "member",
							L: "member",
						},
					},
					{
						// ID: 6,
						ColName: model.CIStr{
							O: "group",
							L: "group",
						},
					},
					{
						// ID: 7,
						ColName: model.CIStr{
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
				Raw: basicModelText,
			},
		},
	}
	mockDBDataSet []btuple.Modifier
)

func IdsAsserter(t *testing.T, expected []primitive.ObjectID, got []primitive.ObjectID) {
	assert.Equal(t, len(expected), len(got))
	for i, modifier := range expected {
		assert.Equal(t, modifier, got[i])
	}
}

func TuplesAsserter(t *testing.T, expected []btuple.Modifier, got []btuple.Modifier) {
	assert.Equal(t, len(expected), len(got))
	for i, modifier := range expected {
		assert.Equal(t, modifier, got[i])
	}
}

func CloneTupleSet(set []btuple.Modifier) (output []btuple.Modifier) {
	for _, modifier := range set {
		output = append(output, modifier.Clone())
	}
	return
}

func UpdateValue(set []btuple.Modifier, schema bschema.Reader, updateAttrs map[int]plan.Modifier) {
	for _, s := range set {
		for i := 0; i < schema.FieldsLen(); i++ {
			if m, ok := updateAttrs[i]; ok {
				switch m.Type() {
				case plan.ModifierSet:
					s.Set(i, m.Value().([]byte))
				}
			}
		}
	}
}

func MergeDefaultValue(set []btuple.Modifier, schema bschema.Reader) {
	for _, mo := range set {
		err := mo.MergeDefaultValue(schema)
		PanicIfErr(err)
	}
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	var err error
	mockDBDataSet, err = utils.CsvToTuples("../../../examples/assets/policy/basic_policy.csv")
	PanicIfErr(err)
}

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

func (db *mockDB) CreateDB(t *testing.T, sc session.Context, info *model.DBInfo) {
	exec := NewSchemaExec(sc, plan.NewCreateDBPlan(info))
	exec.Init()
	_, err := exec.Next(context.TODO(), nil, nil)
	assert.Nil(t, err)
}

func (db *mockDB) InsertTuples(t *testing.T, sc session.Context, dbOid, tableOid uint64, tuples []btuple.Modifier) (result []btuple.Modifier, ids []primitive.ObjectID, err error) {
	builder := executorBuilder{ctx: sc}
	executor := builder.Build(plan.NewRawInsertPlan(tuples, dbOid, tableOid))
	assert.Nil(t, builder.Error())
	result, ids, err = Execute(executor, context.TODO())
	return
}

func (db *mockDB) SeqScan(t *testing.T, sc session.Context, dbOid, tableOid uint64, schema bschema.Reader) (result []btuple.Modifier, ids []primitive.ObjectID, err error) {
	builder := executorBuilder{ctx: sc}
	executor, err := builder.Build(plan.NewSeqScanPlan(schema, nil, dbOid, tableOid)), builder.Error()
	assert.Nil(t, err)
	result, ids, err = Execute(executor, context.TODO())
	return
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
				id, err = meta.GetColumnId(tableInfo.ID, column.ColName.L)
				assert.Nil(t, err)
				assert.Equal(t, column.ID, id)
			}
			//TODO: foreigne keys
		}

	}
}
