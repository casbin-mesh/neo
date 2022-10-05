package plan

import (
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

type PlanType int

const (
	InsertPlanType PlanType = iota + 1
	UpdatePlanType
	SeqScanPlanType
	CreateDBPlanType
)

type Plan interface {
	GetType() PlanType
}

type AbstractPlan interface {
	OutputSchema() bschema.Reader
	HasChildren() bool
	GetChildAt(idx int) AbstractPlan
	GetChildren() []AbstractPlan
	GetType() PlanType
	String() string
	FindBestPlan(ctx session.OptimizerCtx) AbstractPlan
}

type abstractPlan struct {
	schema   bschema.Reader
	children []AbstractPlan
}

func (a abstractPlan) FindBestPlan(ctx session.OptimizerCtx) AbstractPlan {
	panic("unimplemented")
}

func (a abstractPlan) String() string {
	return "unimplemented"
}

func (a abstractPlan) OutputSchema() bschema.Reader {
	return a.schema
}

func (a abstractPlan) HasChildren() bool {
	return len(a.children) > 0
}

func (a abstractPlan) GetChildAt(idx int) AbstractPlan {
	return a.children[idx]
}

func (a abstractPlan) GetChildren() []AbstractPlan {
	return a.children
}

func (a abstractPlan) GetType() PlanType {
	panic("unreachable")
}

func NewAbstractPlan(schema bschema.Reader, children []AbstractPlan) AbstractPlan {
	return &abstractPlan{
		schema:   schema,
		children: children,
	}
}
