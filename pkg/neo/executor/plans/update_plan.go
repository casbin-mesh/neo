package plans

import "github.com/casbin-mesh/neo/pkg/primitive/btuple"

type UpdatePlan interface {
	AbstractPlan
	TableOid() uint64
	GetUpdateAttrs() map[int]btuple.Elem
}

type updatePlan struct {
	AbstractPlan
	tableOid uint64
	// map column idx-> tuple element
	updateAttrs map[int]btuple.Elem
}

func (u updatePlan) GetType() PlanType {
	return UpdatePlanType
}

func (u updatePlan) TableOid() uint64 {
	return u.tableOid
}

func (u updatePlan) GetUpdateAttrs() map[int]btuple.Elem {
	return u.updateAttrs
}

func NewUpdatePlan(children []AbstractPlan, tableOid uint64, updateAttrs map[int]btuple.Elem) UpdatePlan {
	return &updatePlan{
		AbstractPlan: NewAbstractPlan(nil, children),
		tableOid:     tableOid,
		updateAttrs:  updateAttrs,
	}
}
