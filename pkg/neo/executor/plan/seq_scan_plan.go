package plan

import (
	"github.com/casbin-mesh/neo/pkg/neo/executor/expression"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

type SeqScanPlan interface {
	AbstractPlan
	Predicate() expression.AbstractExpression
	DBOid() uint64
	TableOid() uint64
}

type seqScanPlan struct {
	AbstractPlan
	predicate expression.AbstractExpression
	tableOid  uint64
	dbOid     uint64
	//TODO(weny): add scan prefix here
}

func (s seqScanPlan) TableOid() uint64 {
	return s.tableOid
}

func (s seqScanPlan) DBOid() uint64 {
	return s.dbOid
}

func (s seqScanPlan) GetType() PlanType {
	return SeqScanPlanType
}

func (s seqScanPlan) Predicate() expression.AbstractExpression {
	return s.predicate
}

func NewSeqScanPlan(schema bschema.Reader, predicate expression.AbstractExpression, dbOid, tableOid uint64) SeqScanPlan {
	return &seqScanPlan{
		AbstractPlan: NewAbstractPlan(schema, nil),
		predicate:    predicate,
		dbOid:        dbOid,
		tableOid:     tableOid,
	}
}
