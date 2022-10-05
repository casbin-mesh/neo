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

package executor

import (
	"context"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type matcherExecutor struct {
	baseExecutor
	plan     plan.MatcherPlan
	done     bool
	children []Executor
}

func (m *matcherExecutor) Init() {
	for _, child := range m.children {
		child.Init()
	}
}

var (
	True  = btuple.NewModifier([]btuple.Elem{[]byte{1}})
	False = btuple.NewModifier([]btuple.Elem{[]byte{0}})
)

func (m *matcherExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	if m.done {
		return false, nil
	}

	switch m.plan.EffectType() {
	case plan.AllowOverride:
		next, err = m.children[0].Next(ctx, tuple, rid)
		if err != nil {
			return
		}
		if !next { // no match policies
			*tuple = False
		} else {
			*tuple = True
		}
		m.done = true
	}
	return true, nil
}

func (m *matcherExecutor) Close() error {
	var errs []error
	for _, child := range m.children {
		err := child.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

func NewMatcherExecutor(ctx session.Context, plan plan.MatcherPlan, children []Executor) Executor {
	return &matcherExecutor{
		baseExecutor: newBaseExecutor(ctx),
		plan:         plan,
		children:     children,
	}
}
