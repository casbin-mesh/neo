package plan

import "github.com/casbin-mesh/neo/pkg/primitive/bschema"

type TableRowIdScan struct {
	AbstractPlan
	tableOid uint64
	dbOid    uint64
}

func (s TableRowIdScan) TableOid() uint64 {
	return s.tableOid
}

func (s TableRowIdScan) DBOid() uint64 {
	return s.dbOid
}

func NewTableRowIdScan(schema bschema.Reader, dbOid, tableOid uint64, child AbstractPlan) *TableRowIdScan {
	return &TableRowIdScan{
		AbstractPlan: NewAbstractPlan(schema, []AbstractPlan{child}),
		tableOid:     tableOid,
		dbOid:        dbOid,
	}
}
