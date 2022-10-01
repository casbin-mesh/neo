package plan

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

type TableRowIdScan struct {
	AbstractPlan
	tableOid  uint64
	dbOid     uint64
	predicate expression.Expression
	ctx       ast.EvaluateCtx
}

func (s TableRowIdScan) Predicate() expression.Expression {
	return s.predicate
}

func (s TableRowIdScan) GetEvalCtx() ast.EvaluateCtx {
	return s.ctx
}

func (s TableRowIdScan) TableOid() uint64 {
	return s.tableOid
}

func (s TableRowIdScan) DBOid() uint64 {
	return s.dbOid
}

func NewTableRowIdScan(schema bschema.Reader, predicate expression.Expression, ctx ast.EvaluateCtx, dbOid, tableOid uint64, child AbstractPlan) *TableRowIdScan {
	return &TableRowIdScan{
		AbstractPlan: NewAbstractPlan(schema, []AbstractPlan{child}),
		tableOid:     tableOid,
		dbOid:        dbOid,
		predicate:    predicate,
		ctx:          ctx,
	}
}

func (s TableRowIdScan) String() string {
	childStr := make([]string, 0, len(s.GetChildren()))
	for _, child := range s.GetChildren() {
		childStr = append(childStr, child.String())
	}
	header := "TableRowIdScan"

	if s.predicate != nil {
		header = fmt.Sprintf("%s | Predicate: %s", header, s.predicate.String())
	}
	return utils.TreeFormat(header, childStr...)
}
