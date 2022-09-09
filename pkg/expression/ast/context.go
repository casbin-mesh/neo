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

type AccessorValue interface {
	GetMember(ident string) *Primitive
}

type Parameter interface {
	GetPrimitive() *Primitive
}

type FunctionWithCtx interface {
	Eval(ctx EvaluateCtx, args ...Evaluable) (*Primitive, error)
}

type Variables struct {
	funcIdentifier bool
}

type Context struct {
	fc FirstClass
	Variables
}

func (ctx *Context) GetVars() *Variables {
	return &ctx.Variables
}

type FirstClass map[string]interface{}

func (ctx Context) Get(key string) interface{} {
	return ctx.fc[key]
}

func (ctx Context) AddAccessor(k string, a AccessorValue) {
	ctx.fc[k] = a
}

func (ctx Context) AddFunctionWithCtx(k string, f FunctionWithCtx) {
	ctx.fc[k] = f
}

func (ctx Context) AddParameter(k string, p Primitive) {
	ctx.fc[k] = p
}

func NewContext() *Context {
	return &Context{
		fc: make(FirstClass),
	}
}
