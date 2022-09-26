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

package plan

import (
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type MiddlePlan struct {
	AbstractPlan
	tuples []btuple.Modifier
	tids   []primitive.ObjectID
}

func (m MiddlePlan) Tuples() []btuple.Modifier {
	return m.tuples
}

func (m MiddlePlan) Tids() []primitive.ObjectID {
	return m.tids
}

func NewMiddlePlan(tuples []btuple.Modifier, tids []primitive.ObjectID) *MiddlePlan {
	return &MiddlePlan{
		AbstractPlan: nil,
		tuples:       tuples,
		tids:         tids,
	}
}