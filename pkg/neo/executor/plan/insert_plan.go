package plan

import "github.com/casbin-mesh/neo/pkg/primitive/btuple"

type InsertPlan interface {
	AbstractPlan
	RawValues() []btuple.Modifier
	RawValuesSize() int
	TableOid() uint64
}

type insertPlan struct {
	AbstractPlan
	rawValues []btuple.Modifier
	tableOid  uint64
}

func (i insertPlan) RawValuesSize() int {
	return len(i.rawValues)
}

func (i insertPlan) RawValues() []btuple.Modifier {
	return i.rawValues
}

func (i insertPlan) TableOid() uint64 {
	return i.tableOid
}

func (i insertPlan) GetType() PlanType {
	return InsertPlanType
}

func NewRawInsertPlan(rawValues []btuple.Modifier, tableOid uint64) InsertPlan {
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
