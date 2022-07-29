package plans

import (
	"github.com/casbin-mesh/neo/pkg/neo/executor/expression"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

type SeqScanPlan interface {
	AbstractPlan
	Predicate() expression.AbstractExpression
}

type seqScanPlan struct {
	AbstractPlan
	predicate expression.AbstractExpression
	tableOid  uint64
}

func (s seqScanPlan) GetType() PlanType {
	return SeqScanPlanType
}

func (s seqScanPlan) Predicate() expression.AbstractExpression {
	return s.predicate
}

func NewSeqScanPlan(schema bschema.Reader, predicate expression.AbstractExpression, tableOid uint64) SeqScanPlan {
	return &seqScanPlan{
		AbstractPlan: NewAbstractPlan(schema, nil),
		predicate:    predicate,
		tableOid:     tableOid,
	}
}
