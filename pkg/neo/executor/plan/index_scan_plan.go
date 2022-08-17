package plan

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

type IndexScanPlan interface {
	AbstractPlan
	FetchTuple() bool
	IsValid(key []byte) bool
	DBOid() uint64
	TableOid() uint64
	Prefix() []byte
	PrimaryIndex() bool
}

type indexScanPlan struct {
	AbstractPlan
	tableOid     uint64
	dbOid        uint64
	prefix       []byte
	fetchTuple   bool
	isValid      func(key []byte) bool
	primaryIndex bool
}

func (s indexScanPlan) PrimaryIndex() bool {
	return s.primaryIndex
}

func (s indexScanPlan) FetchTuple() bool {
	return s.fetchTuple
}

func (s indexScanPlan) Prefix() []byte {
	return s.prefix
}

func (s indexScanPlan) IsValid(key []byte) bool {
	return s.isValid(key)
}

func (s indexScanPlan) TableOid() uint64 {
	return s.tableOid
}

func (s indexScanPlan) DBOid() uint64 {
	return s.dbOid
}

func NewIndexScanPlan(schema bschema.Reader, fetchTuple bool, primary bool, prefix []byte, isValid func(key []byte) bool, dbOid, tableOid uint64) IndexScanPlan {
	return &indexScanPlan{
		AbstractPlan: NewAbstractPlan(schema, nil),
		primaryIndex: primary,
		fetchTuple:   fetchTuple,
		prefix:       prefix,
		isValid:      isValid,
		dbOid:        dbOid,
		tableOid:     tableOid,
	}
}
