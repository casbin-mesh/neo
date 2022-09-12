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
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
)

type MatcherPlanType int

const (
	AllowOverride MatcherPlanType = iota
	DenyOverride
	AllowAndDeny
	Priority
	PriorityBaseOnRole
)

type PhysicalPlan struct {
	Plan     plan.AbstractPlan
	Children []PhysicalPlan
	Property
}

type LogicalSeqScan struct {
	plan.AbstractPlan
	Predicate Predicate
	DbId      uint64
	TableId   uint64
}

func (p *LogicalSeqScan) String() string {
	return fmt.Sprintf("%s | DB: %d | Table: %d | Predicate: %s", "LogicalSeqScan", p.DbId, p.TableId, p.Predicate.String())
}

type LogicalIndexLookupReader struct {
	plan.AbstractPlan
	Build     plan.AbstractPlan
	Probe     plan.AbstractPlan
	Predicate Predicate
}

func (p *LogicalIndexLookupReader) String() string {
	childStr := make([]string, 0, 2)
	childStr = append(childStr, "(Build)"+p.Build.String())
	childStr = append(childStr, "(Probe)"+p.Probe.String())
	return treeFormat(fmt.Sprintf("LogicalIndexLookupReader | Predicate: %s", p.Predicate.String()), childStr...)
}

type LogicalRowIdScan struct {
	plan.AbstractPlan
	TableId uint64
}

func (p *LogicalRowIdScan) String() string {
	return fmt.Sprintf("%s | Table: %d", "LogicalRowIdScan", p.TableId)
}

type LogicalIndexReader struct {
	plan.AbstractPlan
	Indexes   []*model.IndexInfo
	Predicate Predicate
}

func (p *LogicalIndexReader) String() string {
	return fmt.Sprintf("LogicalIndexReader | Predicate: %s", p.Predicate.String())
}

type LogicalMatcherPlan struct {
	Type       MatcherPlanType
	Predicates []Predicate
	Children   []plan.AbstractPlan
}

var eftPolicy2Str = []string{
	"AllowOverride",
	"DenyOverride",
	"AllowAndDeny",
	"Priority",
	"PriorityBaseOnRole",
}

func (p *LogicalMatcherPlan) String() string {
	childStr := make([]string, 0, len(p.Children))
	for _, child := range p.Children {
		childStr = append(childStr, child.String())
	}
	return treeFormat(fmt.Sprintf("LogicalMatcherPlan | Type: %s", eftPolicy2Str[p.Type]), childStr...)
}

type LogicalConst struct {
	plan.AbstractPlan
	Predicate Predicate
}

func (p *LogicalConst) String() string {
	return treeFormat(fmt.Sprintf("LogicalConst | Predicate: %s", p.Predicate.String()))
}

type LogicalPredicate struct {
	plan.AbstractPlan
	Predicate Predicate
}

func (p *LogicalPredicate) String() string {
	return treeFormat(fmt.Sprintf("LogicalPredicate | Predicate: %s", p.Predicate.String()))
}

type LogicalAndPlan struct {
	plan.AbstractPlan
	Const    []plan.AbstractPlan
	NonConst []plan.AbstractPlan
}

func (p *LogicalAndPlan) String() string {
	childStr := make([]string, 0, len(p.Const)+len(p.NonConst))
	for _, child := range p.Const {
		childStr = append(childStr, fmt.Sprintf("(Const)%s", child.String()))
	}
	for _, child := range p.NonConst {
		childStr = append(childStr, fmt.Sprintf("(Non-Const)%s", child.String()))
	}
	return treeFormat("LogicalAndPlan", childStr...)
}

type LogicalOrPlan struct {
	plan.AbstractPlan
	Const    []plan.AbstractPlan
	NonConst []plan.AbstractPlan
}

func (p *LogicalOrPlan) String() string {
	childStr := make([]string, 0, len(p.Const)+len(p.NonConst))
	for _, child := range p.Const {
		childStr = append(childStr, fmt.Sprintf("(Const)%s", child.String()))
	}
	for _, child := range p.NonConst {
		childStr = append(childStr, fmt.Sprintf("(Non-Const)%s", child.String()))
	}
	return treeFormat("LogicalOrPlan", childStr...)
}

type Property struct {
	Indexes       []*model.IndexInfo
	SecondScan    bool
	FullTableScan bool
	Cardinality   uint64
}
