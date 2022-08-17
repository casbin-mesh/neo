package plan

import "github.com/casbin-mesh/neo/pkg/primitive/bschema"

type MultiIndexScan interface {
	AbstractPlan
	FetchTuple() bool
	DBOid() uint64
	TableOid() uint64
}

type multiIndexPlan struct {
	AbstractPlan
	tableOid   uint64
	dbOid      uint64
	fetchTuple bool
}

func (p multiIndexPlan) FetchTuple() bool {
	return p.fetchTuple
}

func (p multiIndexPlan) TableOid() uint64 {
	return p.tableOid
}

func (p multiIndexPlan) DBOid() uint64 {
	return p.dbOid
}

func NewMultiIndexScan(children []AbstractPlan, schema bschema.Reader, fetchTuple bool, dbOid, tableOid uint64) MultiIndexScan {
	return &multiIndexPlan{
		AbstractPlan: NewAbstractPlan(schema, children),
		fetchTuple:   fetchTuple,
		tableOid:     tableOid,
		dbOid:        dbOid,
	}
}
