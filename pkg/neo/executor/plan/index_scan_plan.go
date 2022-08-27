package plan

import (
	"github.com/casbin-mesh/neo/pkg/neo/executor/expression"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

type IndexScanPlan interface {
	AbstractPlan
	Predicate() expression.AbstractExpression
	DBOid() uint64
	TableOid() uint64
	Prefix() []byte
}

type indexScanPlan struct {
	AbstractPlan
	tableOid  uint64
	dbOid     uint64
	prefix    []byte
	predicate expression.AbstractExpression
}

func (s indexScanPlan) Prefix() []byte {
	return s.prefix
}

func (s indexScanPlan) Predicate() expression.AbstractExpression {
	return s.predicate
}

func (s indexScanPlan) TableOid() uint64 {
	return s.tableOid
}

func (s indexScanPlan) DBOid() uint64 {
	return s.dbOid
}

func NewIndexScanPlan(schema bschema.Reader, prefix []byte, predicate expression.AbstractExpression, dbOid, tableOid uint64) IndexScanPlan {
	return &indexScanPlan{
		AbstractPlan: NewAbstractPlan(schema, nil),
		prefix:       prefix,
		predicate:    predicate,
		dbOid:        dbOid,
		tableOid:     tableOid,
	}
}
