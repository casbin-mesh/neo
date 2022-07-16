package schema

import (
	"github.com/casbin-mesh/neo/pkg/neo/index"
	"github.com/casbin-mesh/neo/pkg/neo/model"
)

type Reader interface {
	Get(key []byte) (*model.DBInfo, error)
}

type ReaderWriter interface {
	Reader
	Set(key []byte, info *model.DBInfo) error

	CommitAt(commitTs uint64) error
	Rollback()
}

type inMemSchema struct {
	index.Txn[*model.DBInfo]
}

func (i inMemSchema) Get(key []byte) (*model.DBInfo, error) {
	return i.Txn.Get(key)
}

func (i inMemSchema) Set(key []byte, info *model.DBInfo) error {
	return i.Txn.Set(key, info)
}

func (i inMemSchema) CommitAt(commitTs uint64) error {
	return i.Txn.CommitAt(commitTs, nil)
}

func (i inMemSchema) Rollback() {
	//TODO implement me
}

func New(txn index.Txn[*model.DBInfo]) ReaderWriter {
	return &inMemSchema{Txn: txn}
}
