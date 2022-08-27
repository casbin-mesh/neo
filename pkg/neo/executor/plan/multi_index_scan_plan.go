package plan

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/utils"
)

type MultiIndexScan interface {
	AbstractPlan
	DBOid() uint64
	TableOid() uint64
	LeftOutputSchema() bschema.Reader
	RightOutputSchema() bschema.Reader
}

type multiIndexPlan struct {
	AbstractPlan
	left     bschema.Reader
	right    bschema.Reader
	tableOid uint64
	dbOid    uint64
}

func (p multiIndexPlan) LeftOutputSchema() bschema.Reader {
	return p.left
}

func (p multiIndexPlan) RightOutputSchema() bschema.Reader {
	return p.right
}

func (p multiIndexPlan) TableOid() uint64 {
	return p.tableOid
}

func (p multiIndexPlan) DBOid() uint64 {
	return p.dbOid
}

func NewMultiIndexScan(children []AbstractPlan, dbOid, tableOid uint64) MultiIndexScan {
	return &multiIndexPlan{
		//TODO: merge schema
		AbstractPlan: NewAbstractPlan(
			utils.MergeSchema(children[0].OutputSchema(), children[1].OutputSchema()),
			children),
		left:     children[0].OutputSchema(),
		right:    children[1].OutputSchema(),
		tableOid: tableOid,
		dbOid:    dbOid,
	}
}
