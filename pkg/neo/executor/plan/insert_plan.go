package plan

import "github.com/casbin-mesh/neo/pkg/primitive/btuple"

type InsertPlan interface {
	AbstractPlan
	RawValues() []btuple.Builder
	RawValuesSize() int
	TableOid() uint64
}

type insertPlan struct {
	AbstractPlan
	rawValues []btuple.Builder
	tableOid  uint64
}

func (i insertPlan) RawValuesSize() int {
	return len(i.rawValues)
}

func (i insertPlan) RawValues() []btuple.Builder {
	return i.rawValues
}

func (i insertPlan) TableOid() uint64 {
	return i.tableOid
}

func (i insertPlan) GetType() PlanType {
	return InsertPlanType
}

func NewRawInsertPlan(rawValues []btuple.Builder, tableOid uint64) InsertPlan {
	return &insertPlan{
		AbstractPlan: NewAbstractPlan(nil, nil),
		rawValues:    rawValues,
		tableOid:     tableOid,
	}
}

func NewInsertPlan(children []AbstractPlan, tableOid uint64) InsertPlan {
	return &insertPlan{
		AbstractPlan: NewAbstractPlan(nil, children),
		tableOid:     tableOid,
	}
}
