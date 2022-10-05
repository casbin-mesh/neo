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
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type middleExecutor struct {
	plan *plan.MiddlePlan
	iter int
}

func (m *middleExecutor) Init() {
}

func (m *middleExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (bool, error) {
	if m.iter < len(m.plan.Tuples()) {
		*tuple = m.plan.Tuples()[m.iter]
		*rid = m.plan.Tids()[m.iter]
		m.iter++
		return true, nil
	}
	return false, nil
}

func (m *middleExecutor) Close() error {
	return nil
}

func NewMiddleExecutor(plan *plan.MiddlePlan) Executor {
	return &middleExecutor{plan: plan}
}
