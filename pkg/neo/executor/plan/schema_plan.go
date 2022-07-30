package plan

import "github.com/casbin-mesh/neo/pkg/neo/model"

type SchemaPlan interface {
	AbstractPlan
	GetDBInfo() *model.DBInfo
}

type createDBPlan struct {
	AbstractPlan
	db *model.DBInfo
	tp PlanType
}

func (c *createDBPlan) GetDBInfo() *model.DBInfo {
	return c.db
}

func (c *createDBPlan) GetType() PlanType {
	return c.tp
}

func NewCreateDBPlan(db *model.DBInfo) SchemaPlan {
	return &createDBPlan{
		AbstractPlan: NewAbstractPlan(nil, nil),
		db:           db,
		tp:           CreateDBPlanType,
	}
}
