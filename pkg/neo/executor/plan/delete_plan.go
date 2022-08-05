package plan

type DeletePlan interface {
	AbstractPlan
	TableOid() uint64
	DbOid() uint64
}

type deletePlan struct {
	AbstractPlan
	tableOid uint64
	dbOid    uint64
}

func (d deletePlan) DbOid() uint64 {
	return d.dbOid
}

func (d deletePlan) TableOid() uint64 {
	return d.tableOid
}

func (d deletePlan) GetType() PlanType {
	return UpdatePlanType
}

func NewDeletePlan(children []AbstractPlan, tableOid uint64, dbOid uint64) DeletePlan {
	return &deletePlan{
		AbstractPlan: NewAbstractPlan(nil, children),
		tableOid:     tableOid,
		dbOid:        dbOid,
	}
}
