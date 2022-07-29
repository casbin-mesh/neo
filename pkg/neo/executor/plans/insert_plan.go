package plans

import "github.com/casbin-mesh/neo/pkg/primitive/btuple"

type InsertPlan interface {
	AbstractPlan
	RawValues() []btuple.Reader
	TableOid() uint64
}

type insertPlan struct {
	AbstractPlan
	rawValues []btuple.Reader
	tableOid  uint64
}

func (i insertPlan) RawValues() []btuple.Reader {
	return i.rawValues
}

func (i insertPlan) TableOid() uint64 {
	return i.tableOid
}

func (i insertPlan) GetType() PlanType {
	return InsertPlanType
}

func NewRawInsertPlan(rawValues []btuple.Reader, tableOid uint64) InsertPlan {
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
