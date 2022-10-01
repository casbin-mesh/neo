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
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
)

var (
	ErrUnsupportedFilter = errors.New("unsupported filter")
	ErrFailedInsertValue = errors.New("failed to insert value")
)

type table struct {
	engine      Engine
	tableName   string
	dbName      string
	db          *model.DBInfo
	table       *model.TableInfo
	matcherPlan plan.AbstractPlan
}

func (n *table) Update(ctx context.Context, filter interface{}, update interface{}, opts ...*UpdateOptions) (int, error) {
	opt := MergeUpdateOptions(opts...)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)
	if err != nil {
		return 0, err
	}
	if err = n.Init(sessCtx); err != nil {
		return 0, err
	}

	builder := executor.NewExecutorBuilder(sessCtx)
	queryPlan, err := n.filter2OptimizedQueryPlan(ctx, filter)
	if err != nil {
		return 0, err
	}

	// retrieves data
	exec, err := builder.TryBuild(queryPlan)
	result, tids, err := executor.Execute(exec, ctx)
	if err != nil {
		return 0, err
	}
	updateInfo, err := generateUpdateAttrInfoFromInterface(n.table, update)
	if err != nil {
		return 0, err
	}

	// update plan
	updatePlan := plan.NewUpdatePlan([]plan.AbstractPlan{plan.NewMiddlePlan(result, tids)}, n.table.ID, n.db.ID, updateInfo)

	// builds executor
	exec, err = builder.Build(updatePlan), builder.Error()
	if err != nil {
		return 0, err
	}
	// executes
	_, _, err = executor.Execute(exec, ctx)
	if err != nil {
		return 0, err
	}

	if opt.AutoCommit() {
		err = n.engine.commitSession(ctx, sessCtx)
		if err != nil {
			return 0, err
		}
	}

	return len(tids), nil
}

func (n *table) Delete(ctx context.Context, filter interface{}, opts ...*DeleteOptions) ([]primitive.ObjectID, error) {
	opt := MergeDeleteOptions(opts...)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)
	if err != nil {
		return nil, err
	}
	if err = n.Init(sessCtx); err != nil {
		return nil, err
	}

	builder := executor.NewExecutorBuilder(sessCtx)
	queryPlan, err := n.filter2OptimizedQueryPlan(ctx, filter)
	if err != nil {
		return nil, err
	}

	// retrieves data
	exec, err := builder.TryBuild(queryPlan)
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

	return tids, nil
}

func (n *table) filter2OptimizedQueryPlan(ctx context.Context, filter interface{}) (queryPlan plan.AbstractPlan, err error) {
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
			return nil, ErrUnsupportedFilter
		}
	}
	return queryPlan, nil
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

	builder := executor.NewExecutorBuilder(sessCtx)
	queryPlan, err := n.filter2OptimizedQueryPlan(ctx, filter)
	if err != nil {
		return nil, err
	}
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

func (n table) InsertOne(ctx context.Context, tid primitive.ObjectID, data A, opts ...*InsertOptions) (A, error) {
	opt := MergeInsertOptions(opts...)
	opt.SetUpdateTxn(true)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)
	if err != nil {
		return nil, err
	}
	if err = n.Init(sessCtx); err != nil {
		return nil, err
	}
	builder := executor.NewExecutorBuilder(sessCtx)
	exec, err := builder.Build(plan.NewRawInsertPlan([]primitive.ObjectID{tid}, []value.Values{A2Values(data)}, n.db.ID, n.table.ID)), builder.Error()
	if err != nil {
		return nil, err
	}
	result, _, err := executor.Execute(exec, ctx)

	if err != nil {
		return nil, err
	}

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

func (n table) InsertMany(ctx context.Context, tids []primitive.ObjectID, data []A, opts ...*InsertOptions) ([]A, error) {
	opt := MergeInsertOptions(opts...)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)
	if err = n.Init(sessCtx); err != nil {
		return nil, err
	}
	builder := executor.NewExecutorBuilder(sessCtx)
	exec := builder.Build(plan.NewRawInsertPlan(tids, A2ValuesArray(data), n.db.ID, n.table.ID))
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
	opt.SetUpdateTxn(true)
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

var (
	ErrInvalidUpdate              = errors.New("invalid update")
	ErrUnsupportedUpdate          = errors.New("unsupported update")
	ErrUnsupportedUpdateOperation = errors.New("unsupported update operation")
	SetOperation                  = "$set"
)

func buildSetOperation(table *model.TableInfo, update map[string]interface{}) plan.UpdateAttrsInfo {
	info := plan.UpdateAttrsInfo{}
	for key, v := range update {
		idx := table.Field(key)
		if idx == -1 {
			// ignore not exists fields
			continue
		}
		info[idx] = plan.NewModifier(plan.ModifierSet, value.NewValueFromInterface(v))
	}
	return info
}

func generateUpdateAttrInfoFromInterface(table *model.TableInfo, update interface{}) (plan.UpdateAttrsInfo, error) {
	if update == nil {
		return nil, ErrInvalidUpdate
	}
	switch values := update.(type) {
	case map[string]interface{}:
		for s, m := range values {
			switch s {
			case SetOperation:
				mm, ok := m.(map[string]interface{})
				if !ok {
					return nil, ErrUnsupportedUpdate
				}
				return buildSetOperation(table, mm), nil
			default:
				return nil, ErrUnsupportedUpdateOperation
			}
		}
		return nil, ErrUnsupportedUpdate
	default:
		return nil, ErrUnsupportedUpdate
	}
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
	opt.SetUpdateTxn(true)
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

func (n *table) Analyze(ctx context.Context, data A) (string, error) {
	sessCtx, err := n.engine.getSessionCtx(ctx, &BaseOptions{})
	defer n.engine.discardSession(ctx, sessCtx)

	if err != nil {
		return "", err
	}
	if err = n.Init(sessCtx); err != nil {
		return "", err
	}
	// generates accessor from data array
	accessor, err := EnforceValue2Accessor(n.table, data)
	if err != nil {
		return "", err
	}
	// by default, using the 0 matcher
	matcher := n.db.MatcherInfo[0]
	optimizerCtx := NewBaseCtx(matcher, n.db, n.table)
	optimizerCtx.SetReqAccessor(accessor)

	mg := optimizer.NewMatcherGenerator(optimizerCtx)
	o := optimizer.NewOptimizer(optimizerCtx)

	matcherPlan := o.Optimizer(mg.Generate(matcher.Predicate))
	return matcherPlan.String(), nil
}

func (n *table) EnforceOne(ctx context.Context, data A, opts ...*EnforceOptions) (bool, error) {
	opt := MergeEnforceOptions(opts...)
	opt.SetUpdateTxn(true)
	sessCtx, err := n.engine.getSessionCtx(ctx, opt.BaseOptions)
	defer n.engine.discardSession(ctx, sessCtx)

	if n.matcherPlan == nil {
		if err != nil {
			return false, err
		}
		if err = n.Init(sessCtx); err != nil {
			return false, err
		}
		// generates accessor from data array
		accessor, err := EnforceValue2Accessor(n.table, data)
		if err != nil {
			return false, err
		}
		// by default, using the 0 matcher
		matcher := n.db.MatcherInfo[0]
		optimizerCtx := NewBaseCtx(matcher, n.db, n.table)
		optimizerCtx.SetReqAccessor(accessor)

		mg := optimizer.NewMatcherGenerator(optimizerCtx)
		o := optimizer.NewOptimizer(optimizerCtx)
		n.matcherPlan = o.Optimizer(mg.Generate(matcher.Predicate))
	}

	builder := executor.NewExecutorBuilder(sessCtx)

	exec, err := builder.TryBuild(n.matcherPlan)

	result, _, err := executor.Execute(exec, ctx)
	if err != nil {
		return false, err
	}
	return result[0] == executor.True, nil
}

func (n table) EnforceMany(ctx context.Context, data []A, opts ...*EnforceOptions) ([]bool, error) {
	//TODO implement me
	panic("implement me")
}
