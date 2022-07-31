// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
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
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type schemaExec struct {
	baseExecutor
	schemaPlan plan.SchemaPlan
	done       bool
}

func (s *schemaExec) Init() {
	s.done = false
}

func (s schemaExec) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (bool, error) {
	if s.done {
		return false, nil
	}

	switch s.schemaPlan.GetType() {
	case plan.CreateDBPlanType:
		_, err := s.GetSessionCtx().GetCatalog().CreateDBInfo(context.TODO(), s.schemaPlan.GetDBInfo())
		if err != nil {
			return false, err
		}
	}

	s.done = true
	return false, nil
}

func NewSchemaExec(ctx session.Context, plan plan.SchemaPlan) Executor {
	return &schemaExec{
		baseExecutor: newBaseExecutor(ctx),
		schemaPlan:   plan,
	}
}
