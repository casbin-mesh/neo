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

// getLogicalAndRetValue if all values are truthy, the value of the last operand is returned.
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Logical_AND
func getLogicalAndRetValue(ctx EvaluateCtx, l, r Evaluable) (*Primitive, error) {
	var (
		lhs, rhs *Primitive
		err      error
	)

	if lhs, err = l.Evaluate(ctx); err != nil {
		return nil, err
	}
	for lhs.Typ == IDENTIFIER {
		lhs, err = lhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Short Circuital
	if lhs.Typ == BOOLEAN && !lhs.AsBool(ctx) {
		lhs.Value = false
		return lhs, nil
	}

	if rhs, err = r.Evaluate(ctx); err != nil {
		return nil, err
	}
	for rhs.Typ == IDENTIFIER {
		rhs, err = rhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	lVal, rVal := lhs.AsBool(ctx), rhs.AsBool(ctx)

	if lhs.Typ == BOOLEAN && rhs.Typ == BOOLEAN {
		lhs.Typ = BOOLEAN
		lhs.Value = lVal && rVal
		return lhs, nil
	} else {
		if lVal && rVal { // if all values are truthy, the value of the last operand is returned.
			return rhs, nil
		} else if !lVal { // returning immediately with the value of the first falsy operand it encounters;
			return lhs, nil
		} else if !rVal {
			return rhs, nil
		}
	}
	panic("unreachable")
}
