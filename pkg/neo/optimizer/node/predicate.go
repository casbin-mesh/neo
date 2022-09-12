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

package node

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"strings"
)

type PredicateType int

const (
	Or PredicateType = iota + 1
	And
	Other
)

type Predicate struct {
	Type PredicateType
	Args []Predicate
	Expr ast.Evaluable
}

func (p *Predicate) HasPredicates() bool {
	return len(p.Args) > 0 || p.Expr != nil
}

func (p *Predicate) String() string {
	switch p.Type {
	case And:
		conds := make([]string, 0, len(p.Args))
		for _, arg := range p.Args {
			conds = append(conds, arg.String())
		}

		return fmt.Sprintf("( %s )", strings.Join(conds, " && "))
	case Or:
		conds := make([]string, 0, len(p.Args))
		for _, arg := range p.Args {
			conds = append(conds, arg.String())
		}
		return fmt.Sprintf("( %s )", strings.Join(conds, " || "))
	case Other:
		return p.Expr.String()
	default:
		return "unknown predicate"
	}
}

func (p *Predicate) Clone() Predicate {
	np := *p
	np.Args = make([]Predicate, len(p.Args))
	for i, arg := range p.Args {
		np.Args[i] = arg.Clone()
	}
	if p.Expr != nil {
		np.Expr = p.Expr.Clone()
	}
	return np
}

func GetPredicateAccessorMembers(p Predicate) (result []string) {
	nameSet := make(map[string]struct{})
	switch p.Type {
	case Or, And:
		for _, arg := range p.Args {
			for _, name := range GetPredicateAccessorMembers(arg) {
				nameSet[name] = struct{}{}
			}
		}
	case Other:
		return expression.GetAccessorMembers(p.Expr)
	}
	result = make([]string, 0, len(nameSet))
	for name, _ := range nameSet {
		result = append(result, name)
	}
	return result
}
