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

type Error struct {
	error
}

func (e *Error) GetMutChildAt(idx int) *Evaluable {
	return nil
}

func (e *Error) Clone() Evaluable {
	ne := *e
	return &ne
}

func (e *Error) GetChildAt(idx int) Evaluable {
	return nil
}

func (e *Error) ChildrenLen() int {
	return 0
}

func (e *Error) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	return nil, e.error
}

type Accessor struct {
	Typ      Type
	Ancestor Evaluable
	Ident    Evaluable
}

func (e *Accessor) GetMutChildAt(idx int) *Evaluable {
	if idx == 0 {
		return &e.Ancestor
	} else {
		return &e.Ident
	}
}

func (e *Accessor) Clone() Evaluable {
	return &Accessor{
		Typ:      e.Typ,
		Ancestor: e.Ancestor.Clone(),
		Ident:    e.Ident.Clone(),
	}
}

func (e *Accessor) GetChildAt(idx int) Evaluable {
	if idx == 0 {
		return e.Ancestor
	} else {
		return e.Ident
	}
}

func (e *Accessor) ChildrenLen() int {
	return 2
}

func (a Accessor) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	//TODO implement me
	panic("implement me")
}

type ScalarFunction struct {
	Ident Evaluable
	Args  []Evaluable
}

func (e ScalarFunction) GetMutChildAt(idx int) *Evaluable {
	if idx == 0 {
		return &e.Ident
	} else {
		// TODO:
		panic("unsupported")
		//return Args(e.Args)
	}
}

type Args []Evaluable
type ArgsRef []*Evaluable

func (a Args) Clone() Evaluable {
	return Args(CloneSlice(a))
}

func (a Args) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	return nil, nil
}

func (a Args) GetChildAt(idx int) Evaluable {
	return a[idx]
}

func (a Args) GetMutChildAt(idx int) *Evaluable {
	return &a[idx]
}

func (a Args) ChildrenLen() int {
	return len(a)
}

func CloneSlice(a []Evaluable) []Evaluable {
	newArgs := make([]Evaluable, 0, len(a))
	for _, evaluable := range a {
		newArgs = append(newArgs, evaluable.Clone())
	}
	return newArgs
}

func (a ScalarFunction) Clone() Evaluable {
	return &ScalarFunction{
		Ident: a.Ident.Clone(),
		Args:  CloneSlice(a.Args),
	}
}

func (s ScalarFunction) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	ctx.GetVars().funcIdentifier = true
	ident, err := s.Ident.Evaluate(ctx)
	ctx.GetVars().funcIdentifier = false
	if err != nil {
		return nil, err
	}
	f := ctx.Get(ident.Value.(string))

	switch fn := f.(type) {
	case FunctionWithCtx:
		return fn.Eval(ctx, s.Args...)
	default:
		return null, nil
	}
}

func (e *ScalarFunction) GetChildAt(idx int) Evaluable {
	if idx == 0 {
		return e.Ident
	} else {
		return Args(e.Args)
	}
}

func (e *ScalarFunction) ChildrenLen() int {
	return 2
}

type Primitive struct {
	Typ   Type
	Value interface{}
}

func (p *Primitive) GetMutChildAt(idx int) *Evaluable {
	return nil
}

// AsBool https://developer.mozilla.org/en-US/docs/Glossary/Truthy
func (p *Primitive) AsBool(ctx EvaluateCtx) bool {
	switch p.Typ {
	case NULL:
		return false
	case INT:
		return p.Value.(int) != 0
	case FLOAT:
		return p.Value.(float64) != 0.0
	case BOOLEAN:
		return p.Value.(bool)
	case STRING:
		return len(p.Value.(string)) != 0
	case TUPLE:
		return p.Value != nil
	case IDENTIFIER:
		value := ctx.Get(p.Value.(string))
		return value != nil
	default:
		return p.Value != nil
	}
}

func (p *Primitive) Equal(another *Primitive, ctx EvaluateCtx) bool {
	if another.Typ != p.Typ {
		return false
	}
	switch p.Typ {
	case INT:
		return compareInt(p.Value.(int), another.Value.(int)) == 0
	case FLOAT:
		return compareFloat(p.Value.(float64), another.Value.(float64)) == 0
	case BOOLEAN:
		return compareBool(p.Value.(bool), another.Value.(bool)) == 0
	case STRING:
		return compareString(p.Value.(string), another.Value.(string)) == 0
	case IDENTIFIER:
		value := ctx.Get(p.Value.(string))
		if value == nil {
			return false
		}
		// if it can convert to primitive
		if param, ok := value.(*Primitive); ok && param != nil {
			return param.Equal(another, ctx)
		}

		//TODO: handles when the identifier is a function
	}
	return false
}

func (p *Primitive) IsNil() bool {
	return p.Value == nil
}

func (p *Primitive) Clone() Evaluable {
	np := *p
	return &np
}

func (e *Primitive) GetChildAt(idx int) Evaluable {
	return nil
}

func (e *Primitive) ChildrenLen() int {
	return 0
}

func (p *Primitive) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	if p.Typ == IDENTIFIER && !ctx.GetVars().funcIdentifier {
		value := ctx.Get(p.Value.(string))
		if value == nil {
			return null, nil
		}
		// it can convert to parameter
		if param, ok := value.(Primitive); ok {
			return &param, nil
		}
		//TODO: handles when the identifier is a function
		return null, nil
	}
	return p, nil
}
