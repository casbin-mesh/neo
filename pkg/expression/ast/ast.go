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

package ast

type EvaluateCtx interface {
	Get(key string) interface{}
	GetVars() *Variables
}

type Evaluable interface {
	Evaluate(ctx EvaluateCtx) (*Primitive, error)
	// GetChildAt returns am immutable ast.Evaluable
	GetChildAt(idx int) Evaluable
	// GetMutChildAt returns a mutable ref of ast.Evaluable
	GetMutChildAt(idx int) *Evaluable
	ChildrenLen() int
	Clone() Evaluable
}
