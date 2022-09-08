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

package builtin

import (
	"errors"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
)

var (
	ErrInvalidArgs    = errors.New("invalid arguments")
	ErrInvalidArgType = errors.New("invalid argument's type")
)

type CoupleStrFn struct {
	fn func(a, b string) string
}

type TripleStrFn struct {
	fn func(a, b, c string) string
}

type CoupleStrFnRetBool struct {
	fn func(a, b string) bool
}

type CoupleStrFnRetBoolAndError struct {
	fn func(a, b string) (bool, error)
}

func (c CoupleStrFnRetBoolAndError) Eval(ctx ast.EvaluateCtx, args ...ast.Evaluable) (*ast.Primitive, error) {
	if len(args) != 2 {
		return nil, ErrInvalidArgs
	}
	a1, err := args[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a1.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a2, err := args[1].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a2.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a1.Typ = ast.BOOLEAN
	a1.Value, err = c.fn(a1.Value.(string), a2.Value.(string))
	return a1, err
}

func (s CoupleStrFnRetBool) Eval(ctx ast.EvaluateCtx, args ...ast.Evaluable) (*ast.Primitive, error) {
	if len(args) != 2 {
		return nil, ErrInvalidArgs
	}
	a1, err := args[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a1.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a2, err := args[1].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a2.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a1.Typ = ast.BOOLEAN
	a1.Value = s.fn(a1.Value.(string), a2.Value.(string))
	return a1, nil
}

func (t TripleStrFn) Eval(ctx ast.EvaluateCtx, args ...ast.Evaluable) (*ast.Primitive, error) {
	if len(args) != 3 {
		return nil, ErrInvalidArgs
	}
	a1, err := args[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a1.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a2, err := args[1].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a2.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a3, err := args[2].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a3.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a1.Typ = ast.STRING
	a1.Value = t.fn(a1.Value.(string), a2.Value.(string), a3.Value.(string))
	return a1, nil
}

func (s CoupleStrFn) Eval(ctx ast.EvaluateCtx, args ...ast.Evaluable) (*ast.Primitive, error) {
	if len(args) != 2 {
		return nil, ErrInvalidArgs
	}
	a1, err := args[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a1.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a2, err := args[1].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if a2.Typ != ast.STRING {
		return nil, ErrInvalidArgType
	}
	a1.Typ = ast.STRING
	a1.Value = s.fn(a1.Value.(string), a2.Value.(string))
	return a1, nil
}

func NewCoupleStrFnRetBoolAndError(fn func(a, b string) (bool, error)) ast.FunctionWithCtx {
	return &CoupleStrFnRetBoolAndError{fn: fn}
}

func NewCoupleStrFn(fn func(a, b string) string) ast.FunctionWithCtx {
	return &CoupleStrFn{fn: fn}
}

func NewCoupleStrRetBoolFn(fn func(a, b string) bool) ast.FunctionWithCtx {
	return &CoupleStrFnRetBool{fn: fn}
}

func NewTripleStrFn(fn func(a, b, c string) string) ast.FunctionWithCtx {
	return &TripleStrFn{fn: fn}
}

var (
	BuildinFnSet = map[string]ast.FunctionWithCtx{
		"keyGet":    NewCoupleStrFn(KeyGet),
		"keyGet2":   NewTripleStrFn(KeyGet2),
		"keyGet3":   NewTripleStrFn(KeyGet3),
		"keyMatch":  NewCoupleStrRetBoolFn(KeyMatch),
		"keyMatch2": NewCoupleStrRetBoolFn(KeyMatch2),
		"keyMatch3": NewCoupleStrRetBoolFn(KeyMatch3),
		"keyMatch4": NewCoupleStrRetBoolFn(KeyMatch4),
		"keyMatch5": NewCoupleStrRetBoolFn(KeyMatch5),
		"ipMatch":   NewCoupleStrRetBoolFn(IPMatch),
		"globMatch": NewCoupleStrFnRetBoolAndError(GlobMatch),
	}
)
