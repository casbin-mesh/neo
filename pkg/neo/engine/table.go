// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"context"
	"errors"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/executor"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
)

var (
	DefaultPolicyTableName = "p"
	ErrUnsupporttedFilter  = errors.New("unsupported filter")
	ErrFailedInsertValue   = errors.New("failed to insert value")
)

type table struct {
	engine    Engine
	tableName string
	dbName    string
	db        *model.DBInfo
	table     *model.TableInfo
}

func (n *table) FindOne(ctx context.Context, filter interface{}) (M, error) {
	//TODO implement me
	panic("implement me")
}

// Find a simple eq query
func (n *table) Find(ctx context.Context, filter interface{}, opts ...*FindOptions) ([]M, error) {
	opt := MergeFindOptions(opts...)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)

	if err != nil {
		return nil, err
	}
	if err = n.Init(sessCtx); err != nil {
		return nil, err
	}

	var queryPlan plan.AbstractPlan

	if filter == nil {
		queryPlan = plan.NewSeqScanPlan(n.table, nil, nil, n.table.ID, n.db.ID)
	} else {
		switch v := filter.(type) {
		case map[string]interface{}:
			// generates accessor from filter
			accessor, err := NewValueAccessorFromMap(n.table, v)
			if err != nil {
				return nil, err
			}
			cols := make([]string, 0, len(accessor.lookup))
			for s, _ := range accessor.lookup {
				cols = append(cols, s)
			}
			optimizerCtx := NewBaseCtx(nil, n.db, n.table)
			optimizerCtx.SetReqAccessor(accessor)
			queryAst := n.table.SelectAst(optimizerCtx.ReqAccessorAncestorName(), &model.SelectAstOptions{IncludedColumns: cols})
			sg := optimizer.NewSelectPlanGenerator(optimizerCtx)
			// generates select plan
			selectPlan := sg.Generate(queryAst)
			op := optimizer.NewOptimizer(optimizerCtx)

			// runs physical optimization
			queryPlan = op.Optimizer(selectPlan)

		default:
			return nil, ErrUnsupporttedFilter
		}
	}

	builder := executor.NewExecutorBuilder(sessCtx)
	// retrieves data
	exec, err := builder.Build(queryPlan), builder.Error()
	if err != nil {
		return nil, err
	}
	result, _, err := executor.Execute(exec, ctx)
	if opt.AutoCommit() {
		err = n.engine.commitSession(ctx, sessCtx)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return DecodeValues2Map(result, n.table)
}

type baseCtx struct {
	matcher                 *model.MatcherInfo
	db                      *model.DBInfo
	table                   *model.TableInfo
	reqAccessorAncestorName string
	allowIdent              string
	denyIdent               string
	effectColName           string
	policyTableName         string
	reqAccessor             ast.AccessorValue
}

func (b baseCtx) ReqAccessor() ast.AccessorValue {
	return b.reqAccessor
}

func (b baseCtx) GetTableStatic(name string) session.TableStatic {
	return nil
}

func (b baseCtx) PolicyTableName() string {
	return b.policyTableName
}

func (b baseCtx) EffectColName() string {
	return b.effectColName
}

func (b baseCtx) AllowIdent() string {
	return b.allowIdent
}

func (b baseCtx) DenyIdent() string {
	return b.denyIdent
}

func (b baseCtx) ReqAccessorAncestorName() string {
	return b.reqAccessorAncestorName
}

func (b baseCtx) Matcher() *model.MatcherInfo {
	return b.matcher
}

func (b baseCtx) DB() *model.DBInfo {
	return b.db
}

func (b baseCtx) Table() *model.TableInfo {
	return b.table
}

func (b *baseCtx) SetReqAccessor(accessor ast.AccessorValue) {
	b.reqAccessor = accessor
}

func NewBaseCtx(matcher *model.MatcherInfo, db *model.DBInfo, table *model.TableInfo) *baseCtx {
	return &baseCtx{
		matcher:                 matcher,
		db:                      db,
		table:                   table,
		reqAccessorAncestorName: "r",
		allowIdent:              "allow",
		denyIdent:               "deny",
		effectColName:           "eft",
		policyTableName:         "p",
	}
}

func (n *table) Init(sessCtx session.Context) (err error) {
	if n.db == nil {
		n.db, err = sessCtx.GetCatalog().GetDBInfoByName(n.dbName)
		if err != nil {
			return
		}
	}
	if n.table == nil {
		n.table, err = n.db.TableByLName(n.tableName)
		if err != nil {
			return
		}
	}
	return nil
}

func (n table) InsertOne(ctx context.Context, data A, opts ...*InsertOptions) (A, error) {
	opt := MergeInsertOptions(opts...)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)
	if err != nil {
		return nil, err
	}
	if err = n.Init(sessCtx); err != nil {
		return nil, err
	}
	builder := executor.NewExecutorBuilder(sessCtx)
	exec, err := builder.Build(plan.NewRawInsertPlan([]value.Values{A2Values(data)}, n.db.ID, n.table.ID)), builder.Error()
	if err != nil {
		return nil, err
	}
	result, _, err := executor.Execute(exec, ctx)

	if opt.AutoCommit() {
		err = n.engine.commitSession(ctx, sessCtx)
		if err != nil {
			return nil, err
		}
	}

	if len(result) != 1 {
		return nil, ErrFailedInsertValue
	}
	return DecodeValue(result[0], n.table)
}

func (n table) InsertMany(ctx context.Context, data []A, opts ...*InsertOptions) ([]A, error) {
	opt := MergeInsertOptions(opts...)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)
	if err = n.Init(sessCtx); err != nil {
		return nil, err
	}
	builder := executor.NewExecutorBuilder(sessCtx)
	exec := builder.Build(plan.NewRawInsertPlan(A2ValuesArray(data), n.db.ID, n.table.ID))
	result, _, err := executor.Execute(exec, ctx)

	if opt.AutoCommit() {
		err = n.engine.commitSession(ctx, sessCtx)
		if err != nil {
			return nil, err
		}
	}

	return DecodeValues(result, n.table)
}

func (n table) UpdateOne(ctx context.Context, data A, update A, opts ...*UpdateOptions) (A, error) {
	opt := MergeUpdateOptions(opts...)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)

	if err != nil {
		return nil, err
	}
	if err = n.Init(sessCtx); err != nil {
		return nil, err
	}
	// generates accessor from data array
	accessor, err := Value2Accessor(n.table, data)
	if err != nil {
		return nil, err
	}
	optimizerCtx := NewBaseCtx(nil, n.db, n.table)
	optimizerCtx.SetReqAccessor(accessor)
	// generate select ast
	queryAst := n.table.SelectAst(optimizerCtx.ReqAccessorAncestorName())
	sg := optimizer.NewSelectPlanGenerator(optimizerCtx)
	// generates select plan
	selectPlan := sg.Generate(queryAst)
	op := optimizer.NewOptimizer(optimizerCtx)
	// runs physical optimization
	optimizedSelect := op.Optimizer(selectPlan)

	builder := executor.NewExecutorBuilder(sessCtx)
	// retrieves data
	exec, err := builder.Build(optimizedSelect), builder.Error()
	result, tids, err := executor.Execute(exec, ctx)
	if err != nil {
		return nil, err
	}
	updateInfo := generateUpdateAttrInfo(data, update)
	// update plan
	updatePlan := plan.NewUpdatePlan([]plan.AbstractPlan{plan.NewMiddlePlan(result, tids)}, n.table.ID, n.db.ID, updateInfo)
	// builds executor
	exec, err = builder.Build(updatePlan), builder.Error()
	if err != nil {
		return nil, err
	}
	// executes
	_, _, err = executor.Execute(exec, ctx)
	if err != nil {
		return nil, err
	}

	if opt.AutoCommit() {
		err = n.engine.commitSession(ctx, sessCtx)
		if err != nil {
			return nil, err
		}
	}
	if len(result) == 0 {
		// returns anywhere
		return nil, nil
	}
	return DecodeValue(result[0], n.table)
}

func generateUpdateAttrInfo(data A, update A) plan.UpdateAttrsInfo {
	info := plan.UpdateAttrsInfo{}
	for i := 0; i < len(data); i++ {
		if data[i] != update[i] {
			info[i] = plan.NewModifier(plan.ModifierSet, value.NewValueFromInterface(update[i]))
		}
	}
	return info
}

func (n table) UpdateMany(ctx context.Context, data []A, update []A, opts ...*UpdateOptions) ([]A, error) {
	//TODO implement me
	panic("implement me")
}

func (n table) DeleteOne(ctx context.Context, data A, opts ...*DeleteOptions) (A, error) {
	opt := MergeDeleteOptions(opts...)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)
	if err != nil {
		return nil, err
	}
	if err = n.Init(sessCtx); err != nil {
		return nil, err
	}
	// generates accessor from data array
	accessor, err := Value2Accessor(n.table, data)
	if err != nil {
		return nil, err
	}
	optimizerCtx := NewBaseCtx(nil, n.db, n.table)
	optimizerCtx.SetReqAccessor(accessor)
	// generate select ast
	queryAst := n.table.SelectAst(optimizerCtx.ReqAccessorAncestorName())
	sg := optimizer.NewSelectPlanGenerator(optimizerCtx)
	// generates select plan
	selectPlan := sg.Generate(queryAst)
	op := optimizer.NewOptimizer(optimizerCtx)
	// runs physical optimization
	optimizedSelect := op.Optimizer(selectPlan)

	builder := executor.NewExecutorBuilder(sessCtx)
	// retrieves data
	exec, err := builder.Build(optimizedSelect), builder.Error()
	if err != nil {
		return nil, err
	}
	result, tids, err := executor.Execute(exec, ctx)
	if err != nil {
		return nil, err
	}

	deletePlan := plan.NewDeletePlan(
		[]plan.AbstractPlan{plan.NewMiddlePlan(result, tids)},
		n.table.ID, n.db.ID)

	// builds executor
	exec, err = builder.Build(deletePlan), builder.Error()
	if err != nil {
		return nil, err
	}
	// executes
	_, _, err = executor.Execute(exec, ctx)
	if err != nil {
		return nil, err
	}

	if opt.AutoCommit() {
		err = n.engine.commitSession(ctx, sessCtx)
		if err != nil {
			return nil, err
		}
	}
	if len(result) == 0 {
		// returns anywhere
		return nil, nil
	}
	return DecodeValue(result[0], n.table)
}

func (n table) DeleteMany(ctx context.Context, data []A, opts ...*DeleteOptions) ([]A, error) {
	//TODO implement me
	panic("implement me")
}

func (n table) EnforceOne(ctx context.Context, data A, opts ...*EnforceOptions) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (n table) EnforceMany(ctx context.Context, data []A, opts ...*EnforceOptions) ([]bool, error) {
	//TODO implement me
	panic("implement me")
}
