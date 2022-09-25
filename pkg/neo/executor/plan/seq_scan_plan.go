package plan

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

type SeqScanPlan interface {
	AbstractPlan
	Predicate() expression.Expression
	DBOid() uint64
	TableOid() uint64
	GetEvalCtx() ast.EvaluateCtx
}

type seqScanPlan struct {
	AbstractPlan
	predicate expression.Expression
	tableOid  uint64
	dbOid     uint64
	ctx       ast.EvaluateCtx
	//TODO(weny): add scan prefix here
}

func (s seqScanPlan) GetEvalCtx() ast.EvaluateCtx {
	return s.ctx
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

func (s seqScanPlan) Predicate() expression.Expression {
	return s.predicate
}

func NewSeqScanPlan(schema bschema.Reader, predicate expression.Expression, ctx ast.EvaluateCtx, dbOid, tableOid uint64) SeqScanPlan {
	return &seqScanPlan{
		AbstractPlan: NewAbstractPlan(schema, nil),
		predicate:    predicate,
		dbOid:        dbOid,
		tableOid:     tableOid,
		ctx:          ctx,
	}
}

func (s seqScanPlan) String() string {
	childStr := make([]string, 0, len(s.GetChildren()))
	for _, child := range s.GetChildren() {
		childStr = append(childStr, child.String())
	}
	return utils.TreeFormat(fmt.Sprintf("SeqScanPlan | Predicate: %s", s.Predicate().String()), childStr...)
}
