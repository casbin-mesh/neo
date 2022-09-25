package catalog

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/meta"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/schema"
)

type Catalog interface {
	GetDBInfoByName(name string) (*model.DBInfo, error)
	GetDBInfoByDBId(did uint64) (*model.DBInfo, error)
	CreateDBInfo(ctx context.Context, info *model.DBInfo) (dbId uint64, err error)
	GetMetaRW() meta.ReaderWriter
}

type catalog struct {
	meta   meta.ReaderWriter
	schema schema.ReaderWriter
	txn    db.Txn
}

func (c *catalog) GetTxn() db.Txn {
	return c.txn
}

func (c *catalog) GetMetaRW() meta.ReaderWriter {
	return c.meta
}

func (c *catalog) GetSchemaRW() schema.ReaderWriter {
	return c.schema
}

func (c *catalog) CreateDBInfo(ctx context.Context, info *model.DBInfo) (dbId uint64, err error) {
	rw := c.GetMetaRW()
	if dbId, err = rw.NewDb(info.Name.L); err != nil {
		return
	}

	for _, matcherInfo := range info.MatcherInfo {
		if _, err = c.createMatcher(ctx, dbId, matcherInfo); err != nil {
			return dbId, err
		}
	}

	for _, tableInfo := range info.TableInfo {
		if _, err = c.createTable(ctx, dbId, tableInfo); err != nil {
			return dbId, err
		}
	}

	schemaRW := c.GetSchemaRW()
	txn := c.GetTxn()
	key := codec.DBInfoKey(dbId)
	if err = txn.Set(key, codec.EncodeDBInfo(info)); err != nil {
		return 0, err
	}
	if err = schemaRW.Set(codec.DBInfoKey(dbId), info); err != nil {
		return 0, err
	}

	return
}

func (c *catalog) createTable(ctx context.Context, did uint64, info *model.TableInfo) (tableId uint64, err error) {
	rw := c.GetMetaRW()
	if tableId, err = rw.NewTable(did, info.Name.L); err != nil {
		return
	}
	info.ID = tableId

	txn := c.GetTxn()
	if err = txn.Set(codec.TableInfoKey(tableId), codec.EncodeTableInfo(info)); err != nil {
		return 0, err
	}

	for _, column := range info.Columns {
		if _, err = c.createColumn(ctx, tableId, column); err != nil {
			return
		}
	}

	for _, index := range info.Indices {
		if _, err = c.createIndex(ctx, tableId, index); err != nil {
			return
		}
	}

	//TODO(weny) :foreign keys

	return
}

func (c *catalog) createColumn(ctx context.Context, tid uint64, info *model.ColumnInfo) (columnId uint64, err error) {
	rw := c.GetMetaRW()
	if columnId, err = rw.NewColumn(tid, info.ColName.L); err != nil {
		return
	}
	info.ID = columnId

	txn := c.GetTxn()
	if err = txn.Set(codec.ColumnInfoKey(columnId), codec.EncodeColumnInfo(info)); err != nil {
		return 0, err
	}

	return
}

func (c *catalog) createIndex(ctx context.Context, tid uint64, info *model.IndexInfo) (indexId uint64, err error) {
	rw := c.meta
	if indexId, err = rw.NewIndex(tid, info.Name.L); err != nil {
		return
	}
	info.ID = indexId

	txn := c.GetTxn()
	if err = txn.Set(codec.IndexInfoKey(indexId), codec.EncodeIndexInfo(info)); err != nil {
		return 0, err
	}

	return
}

func (c *catalog) createMatcher(ctx context.Context, did uint64, info *model.MatcherInfo) (matcherId uint64, err error) {
	rw := c.meta
	if matcherId, err = rw.NewMatcher(did, info.Name.L); err != nil {
		return
	}
	info.ID = matcherId

	txn := c.GetTxn()
	if err = txn.Set(codec.MatcherInfoKey(matcherId), codec.EncodeMatcherInfo(info)); err != nil {
		return 0, err
	}

	return
}

func (c *catalog) GetDBInfoByDBId(did uint64) (*model.DBInfo, error) {
	db, err := c.schema.Get(codec.DBInfoKey(did))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (c *catalog) GetDBInfoByName(name string) (*model.DBInfo, error) {
	did, err := c.meta.GetDBId(name)
	if err != nil {
		return nil, err
	}
	return c.GetDBInfoByDBId(did)
}

func NewCatalogWithMetaRW(meta meta.ReaderWriter, schema schema.ReaderWriter, txn db.Txn) Catalog {
	return &catalog{meta: meta, schema: schema, txn: txn}
}

func NewCatalog(schema schema.ReaderWriter, txn db.Txn) Catalog {
	return &catalog{meta: meta.NewDbMeta(txn), schema: schema, txn: txn}
}
