package plan

import "github.com/casbin-mesh/neo/pkg/primitive/value"

type UpdateType byte

const (
	ModifierSet UpdateType = iota
)

type Modifier interface {
	Value() value.Value
	Type() UpdateType
}

type modifier struct {
	typ   UpdateType
	value value.Value
}

func (m modifier) Value() value.Value {
	return m.value
}

func (m modifier) Type() UpdateType {
	return m.typ
}

func NewModifier(typ UpdateType, value value.Value) Modifier {
	return &modifier{typ, value}
}

type UpdateAttrsInfo map[int]Modifier

type UpdatePlan interface {
	AbstractPlan
	TableOid() uint64
	DBOid() uint64
	GetUpdateAttrs() UpdateAttrsInfo
}

type updatePlan struct {
	AbstractPlan
	tableOid uint64
	dbOid    uint64
	// map column idx-> tuple element
	updateAttrs UpdateAttrsInfo
}

func (u updatePlan) DBOid() uint64 {
	return u.dbOid
}

func (u updatePlan) GetType() PlanType {
	return UpdatePlanType
}

func (u updatePlan) TableOid() uint64 {
	return u.tableOid
}

func (u updatePlan) GetUpdateAttrs() UpdateAttrsInfo {
	return u.updateAttrs
}

func NewUpdatePlan(children []AbstractPlan, tableOid, dbOid uint64, updateAttrs UpdateAttrsInfo) UpdatePlan {
	return &updatePlan{
		AbstractPlan: NewAbstractPlan(nil, children),
		tableOid:     tableOid,
		dbOid:        dbOid,
		updateAttrs:  updateAttrs,
	}
}
