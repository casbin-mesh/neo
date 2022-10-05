package plan

import (
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
)

type InsertPlan interface {
	AbstractPlan
	RawValues() []value.Values
	RawIds() []primitive.ObjectID
	RawValuesSize() int
	DBOid() uint64
	TableOid() uint64
}

type insertPlan struct {
	AbstractPlan
	rawValues []value.Values
	oids      []primitive.ObjectID
	dbOid     uint64
	tableOid  uint64
}

func (i insertPlan) RawIds() []primitive.ObjectID {
	return i.oids
}

func (i insertPlan) RawValuesSize() int {
	return len(i.rawValues)
}

func (i insertPlan) RawValues() []value.Values {
	return i.rawValues
}

func (i insertPlan) TableOid() uint64 {
	return i.tableOid
}

func (i insertPlan) DBOid() uint64 {
	return i.dbOid
}

func (i insertPlan) GetType() PlanType {
	return InsertPlanType
}

func NewRawInsertPlan(oids []primitive.ObjectID, rawValues []value.Values, dbOid, tableOid uint64) InsertPlan {
	return &insertPlan{
		AbstractPlan: NewAbstractPlan(nil, nil),
		rawValues:    rawValues,
		oids:         oids,
		dbOid:        dbOid,
		tableOid:     tableOid,
	}
}

func NewInsertPlan(children []AbstractPlan, dbOid, tableOid uint64) InsertPlan {
	return &insertPlan{
		AbstractPlan: NewAbstractPlan(nil, children),
		dbOid:        dbOid,
		tableOid:     tableOid,
	}
}
