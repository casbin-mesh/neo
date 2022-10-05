package plan

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

type IndexScanPlan interface {
	AbstractPlan
	Predicate() expression.Expression
	DBOid() uint64
	TableOid() uint64
	Prefix() []byte
	GetEvalCtx() ast.EvaluateCtx
}

type indexScanPlan struct {
	AbstractPlan
	tableOid  uint64
	dbOid     uint64
	prefix    []byte
	predicate expression.Expression
	ctx       ast.EvaluateCtx
}

func (s indexScanPlan) GetEvalCtx() ast.EvaluateCtx {
	return s.ctx
}

func (s indexScanPlan) Prefix() []byte {
	return s.prefix
}

func (s indexScanPlan) Predicate() expression.Expression {
	return s.predicate
}

func (s indexScanPlan) TableOid() uint64 {
	return s.tableOid
}

func (s indexScanPlan) DBOid() uint64 {
	return s.dbOid
}

func NewIndexScanPlan(schema bschema.Reader, prefix []byte, predicate expression.Expression, ctx ast.EvaluateCtx, dbOid, tableOid uint64) IndexScanPlan {
	return &indexScanPlan{
		AbstractPlan: NewAbstractPlan(schema, nil),
		prefix:       prefix,
		predicate:    predicate,
		dbOid:        dbOid,
		tableOid:     tableOid,
		ctx:          ctx,
	}
}

func (s indexScanPlan) String() string {
	childStr := make([]string, 0, len(s.GetChildren()))
	for _, child := range s.GetChildren() {
		childStr = append(childStr, child.String())
	}
	return utils.TreeFormat(fmt.Sprintf("IndexScanPlan | Predicate: %s", s.Predicate().String()), childStr...)
}
