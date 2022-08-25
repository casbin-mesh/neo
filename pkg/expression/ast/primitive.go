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

func (e *Error) getChildAt(idx int) Evaluable {
	return nil
}

func (e *Error) childrenLen() int {
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

func (e *Accessor) getChildAt(idx int) Evaluable {
	if idx == 0 {
		return e.Ancestor
	} else {
		return e.Ident
	}
}

func (e *Accessor) childrenLen() int {
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

func (e *ScalarFunction) getChildAt(idx int) Evaluable {
	return nil
}

func (e *ScalarFunction) childrenLen() int {
	return 0
}

type Primitive struct {
	Typ   Type
	Value interface{}
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

func (p *Primitive) Clone() *Primitive {
	np := *p
	return &np
}

func (e *Primitive) getChildAt(idx int) Evaluable {
	return nil
}

func (e *Primitive) childrenLen() int {
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
