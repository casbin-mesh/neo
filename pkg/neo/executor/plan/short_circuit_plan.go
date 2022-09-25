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
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
)

type LogicalType int

const (
	AND LogicalType = iota + 1
	OR
)

var typ2Str = []string{
	"Unknown",
	"AND",
	"OR",
}

func (l LogicalType) String() string {
	return typ2Str[l]
}

type ShortCircuitPlan struct {
	AbstractPlan
	LogicalType
	NonConst []AbstractPlan
	Const    []AbstractPlan
}

func NewShortCircuitPlan(NonConst, Const []AbstractPlan, logicalType LogicalType) *ShortCircuitPlan {
	return &ShortCircuitPlan{
		AbstractPlan: nil,
		LogicalType:  logicalType,
		NonConst:     NonConst,
		Const:        Const,
	}
}

func (s *ShortCircuitPlan) String() string {
	childStr := make([]string, 0, len(s.Const)+len(s.NonConst))
	for _, child := range s.Const {
		childStr = append(childStr, fmt.Sprintf("(Const)%s", child.String()))
	}
	for _, child := range s.NonConst {
		childStr = append(childStr, fmt.Sprintf("(Non-Const)%s", child.String()))
	}
	return utils.TreeFormat(fmt.Sprintf("ShortCircuitPlan | Type: %s", s.LogicalType), childStr...)
}
