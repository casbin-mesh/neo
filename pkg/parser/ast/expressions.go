package ast

import (
	"errors"
	"regexp"
)

type Type int

const (
	INT Type = iota + 1
	FLOAT64
	BOOL
	STRING
	VARIABLE
	TUPLE
	NULL
)

type Op int

const (
	ADD Op = iota + 1
	SUB
	MUL
	DIV
	MOD
	UMINUS
	SEPARATOR

	NULL_COALESCENCE
	TERNARY_FALSE
	TERNARY_TRUE

	POW
	LSHIFT
	RSHIFT

	OR
	AND
	XOR
	NOT
	GT
	GTE
	LT
	LTE
	EQ
	NE
	RE
	NRE
	BETWEEN
	BOOL_NOT
	AND_AND
	OR_OR
)

type EvaluateCtx interface {
	GetParameters() Parameters
	GetFunctions() Functions
}

type Evaluable interface {
	Evaluate(ctx EvaluateCtx) (*Primitive, error)
}

// BinaryOperationExpr is for binary operation like `1 + 1`, `1 - 1`, etc.
type BinaryOperationExpr struct {
	// Op is the operator code for BinaryOperation.
	Op
	// L is the left expression in BinaryOperation.
	L Evaluable
	// R is the right expression in BinaryOperation.
	R Evaluable
}

func hasType(typ Type, ps ...*Primitive) []int {
	var result []int
	for i, p := range ps {
		if p.Typ == typ {
			result = append(result, i)
		}
	}
	return result
}

func hasInt(l, r *Primitive) (bool, bool) {
	return l.Typ == INT, r.Typ == INT
}

func hasFloat64(l, r *Primitive) (bool, bool) {
	return l.Typ == FLOAT64, r.Typ == FLOAT64
}

func getArithmeticRetValue(ctx EvaluateCtx, op Op, evalMap ArithmeticEvalMap, l, r *Primitive) *Primitive {
	lInt, rInt := hasInt(l, r)
	lFloat, rFloat := hasFloat64(l, r)
	evalInt := evalMap[op].evalInt
	evalFloat := evalMap[op].evalFloat

	if lInt && rInt {
		l.Typ = INT
		l.Value = evalInt(l.Value.(int), r.Value.(int))
		return l
	} else if lFloat && rFloat {
		l.Typ = FLOAT64
		l.Value = evalFloat(l.Value.(float64), r.Value.(float64))
		return l
	} else {
		if lFloat && rInt {
			l.Typ = FLOAT64
			l.Value = evalFloat(l.Value.(float64), float64(r.Value.(int)))
		} else if lInt && rFloat {
			l.Typ = FLOAT64
			l.Value = evalFloat(float64(l.Value.(int)), r.Value.(float64))
		}
		return l
	}
}

func getConditionalExprRetValue(ctx EvaluateCtx, op Op, evalMap BoolEvalMap, l, r *Primitive) *Primitive {

	evalGroup := evalMap[op]
	if l.Typ != r.Typ {
		l.Typ = BOOL
		l.Value = false
		return l
	}
	switch l.Typ {
	case INT:
		l.Typ = BOOL
		l.Value = evalGroup.evalInt(l.Value.(int), r.Value.(int))
		return l
	case FLOAT64:
		l.Typ = BOOL
		l.Value = evalGroup.evalFloat(l.Value.(float64), r.Value.(float64))
		return l
	case STRING:
		l.Typ = BOOL
		l.Value = evalGroup.evalString(l.Value.(string), r.Value.(string))
		return l
	case BOOL:
		l.Typ = BOOL
		l.Value = evalGroup.evalBool(l.Value.(bool), r.Value.(bool))
		return l
	case VARIABLE:

		// TODO: eval variable
	}

	return BoolFalse
}

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
	for lhs.Typ == VARIABLE {
		lhs, err = lhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Short Circuital
	if lhs.Typ == BOOL && !lhs.AsBool(ctx) {
		lhs.Value = false
		return lhs, nil
	}

	if rhs, err = r.Evaluate(ctx); err != nil {
		return nil, err
	}
	for rhs.Typ == VARIABLE {
		rhs, err = rhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	lVal, rVal := lhs.AsBool(ctx), rhs.AsBool(ctx)

	if lhs.Typ == BOOL && rhs.Typ == BOOL {
		lhs.Typ = BOOL
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

// getLogicalOrRetValue If expr1 can be converted to true, returns expr1; else, returns expr2.
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Logical_OR
func getLogicalOrRetValue(ctx EvaluateCtx, l, r Evaluable) (*Primitive, error) {

	var (
		lhs, rhs *Primitive
		err      error
	)

	if lhs, err = l.Evaluate(ctx); err != nil {
		return nil, err
	}
	for lhs.Typ == VARIABLE {
		lhs, err = lhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Short Circuital
	if lhs.Typ == BOOL && lhs.AsBool(ctx) {
		lhs.Typ = BOOL
		lhs.Value = true
		return lhs, nil
	}

	if rhs, err = r.Evaluate(ctx); err != nil {
		return nil, err
	}
	for rhs.Typ == VARIABLE {
		rhs, err = rhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	lVal, rVal := lhs.AsBool(ctx), rhs.AsBool(ctx)

	if lhs.Typ == BOOL && rhs.Typ == BOOL {
		lhs.Typ = BOOL
		lhs.Value = lVal && rVal
		return lhs, nil
	} else if lVal {
		return lhs, nil
	}
	return rhs, nil
}

func getLogicalExprRetValue(ctx EvaluateCtx, op Op, l, r Evaluable) (*Primitive, error) {
	switch op {
	case AND_AND:
		return getLogicalAndRetValue(ctx, l, r)
	case OR_OR:
		return getLogicalOrRetValue(ctx, l, r)
	}
	return nil, nil
}

// getNullishCoalescingOperationExprRetValue The nullish coalescing operator (??) is a logical operator,
// returns its right-hand side operand when its left-hand side operand is null or undefined,
// otherwise returns its left-hand side operand.
func getNullishCoalescingOperationExprRetValue(ctx EvaluateCtx, l, r *Primitive) *Primitive {
	if l.IsNil() {
		return r
	}
	return l
}

func getBetweenOperationExprRetValue(ctx EvaluateCtx, l, r *Primitive) *Primitive {
	if r.IsNil() {
		return BoolFalse
	}
	for _, elem := range r.Value.([]*Primitive) {
		if l.Equal(elem, ctx) {
			return BoolTrue
		}
	}
	return BoolFalse
}

func (p *BinaryOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	var (
		lhs, rhs *Primitive
		err      error
	)

	switch p.Op {
	case AND_AND, OR_OR:
		return getLogicalExprRetValue(ctx, p.Op, p.L, p.R)
	}

	if lhs, err = p.L.Evaluate(ctx); err != nil {
		return nil, err
	}
	for lhs.Typ == VARIABLE {
		lhs, err = lhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}
	if rhs, err = p.R.Evaluate(ctx); err != nil {
		return nil, err
	}
	for rhs.Typ == VARIABLE {
		rhs, err = rhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	switch p.Op {
	case ADD, SUB, DIV, MUL, MOD, POW:
		return getArithmeticRetValue(ctx, p.Op, defaultArithmeticEvalMap, lhs, rhs), nil
	case EQ, NE, LT, LTE, GT, GTE:
		return getConditionalExprRetValue(ctx, p.Op, defaultBoolEvalMap, lhs, rhs), nil
	case NULL_COALESCENCE:
		return getNullishCoalescingOperationExprRetValue(ctx, lhs, rhs), nil
	case BETWEEN:
		return getBetweenOperationExprRetValue(ctx, lhs, rhs), nil
	}

	return nil, nil
}

type RegexOperationExpr struct {
	Typ     Op
	Target  string
	Pattern *regexp.Regexp
}

func (p *RegexOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	if p.Typ == NRE {
		return &Primitive{Typ: BOOL, Value: !p.Pattern.MatchString(p.Target)}, nil
	}
	return &Primitive{Typ: BOOL, Value: p.Pattern.MatchString(p.Target)}, nil
}

type ScalarFunction struct {
	Ident string
	Args  []*Primitive
}

var (
	ErrFunctionNotExists = errors.New("function not exists")
)

type Args []*Primitive

func (args Args) ToNaiveValues() (output []interface{}) {
	for _, arg := range args {
		output = append(output, arg.Value)
	}
	return output
}

func ConvertNaiveToPrimitive(input interface{}) *Primitive {
	switch v := input.(type) {
	case int:
		return &Primitive{Typ: INT, Value: v}
	case float64:
		return &Primitive{Typ: FLOAT64, Value: v}
	case bool:
		return &Primitive{Typ: BOOL, Value: v}
	case string:
		return &Primitive{Typ: STRING, Value: v}
		//TODO: tuple ?
	}
	return nil
}

func (p *ScalarFunction) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	f := ctx.GetFunctions().Get(p.Ident)
	if f == nil {
		return nil, ErrFunctionNotExists
	}

	switch fn := f.(type) {
	case Function:
		return fn.Eval(p.Args...)
	case FunctionWithCtx:
		return fn.Eval(ctx, p.Args...)
	case NaiveFunction:
		args := Args(p.Args).ToNaiveValues()
		naiveRet, err := fn.NaiveEval(args...)
		return ConvertNaiveToPrimitive(naiveRet), err
	case NaiveFunctionWithCtx:
		args := Args(p.Args).ToNaiveValues()
		naiveRet, err := fn.NaiveEval(ctx, args...)
		return ConvertNaiveToPrimitive(naiveRet), err
	}

	return null, nil
}

type UnaryOperationExpr struct {
	Child Evaluable
	Op
}

func (p *UnaryOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	v, err := p.Child.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for v.Typ == VARIABLE {
		v, err = v.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	switch p.Op {
	case BOOL_NOT:
		if v.Typ == BOOL {
			v.Value = !v.Value.(bool)
			return v, nil
		}
		return &Primitive{Typ: BOOL, Value: !v.AsBool(ctx)}, nil
	case SUB:
		switch v.Typ {
		case INT:
			v.Value = -v.Value.(int)
			return v, nil
		case FLOAT64:
			v.Value = -v.Value.(float64)
			return v, nil
		}
	}
	return nil, nil
}

type TernaryOperationExpr struct {
	Cond  Evaluable
	True  Evaluable
	False Evaluable
}

func (p *TernaryOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	v, err := p.Cond.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for v.Typ == VARIABLE {
		v, err = v.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	flag := false
	if v.Typ == BOOL {
		flag = v.Value.(bool)
	} else {
		flag = v.AsBool(ctx)
	}
	if flag {
		v, err = p.True.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		for v.Typ == VARIABLE {
			v, err = v.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
		}
		return v, nil
	} else {
		v, err = p.False.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		for v.Typ == VARIABLE {
			v, err = v.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
		}
		return v, nil
	}

	return nil, nil
}

type Primitive struct {
	Typ   Type
	Value interface{}
	Text  string
}

// AsBool https://developer.mozilla.org/en-US/docs/Glossary/Truthy
func (p *Primitive) AsBool(ctx EvaluateCtx) bool {
	switch p.Typ {
	case NULL:
		return false
	case INT:
		return p.Value.(int) != 0
	case FLOAT64:
		return p.Value.(float64) != 0.0
	case BOOL:
		return p.Value.(bool)
	case STRING:
		return len(p.Value.(string)) != 0
	case TUPLE:
		return p.Value != nil
	case VARIABLE:
		value := ctx.GetParameters().Get(p.Value.(string))
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
	case FLOAT64:
		return compareFloat(p.Value.(float64), another.Value.(float64)) == 0
	case BOOL:
		return compareBool(p.Value.(bool), another.Value.(bool)) == 0
	case STRING:
		return compareString(p.Value.(string), another.Value.(string)) == 0
	case VARIABLE:
		value := ctx.GetParameters().Get(p.Value.(string))
		if value == nil {
			return false
		}
		// it can convert to parameter
		if param, ok := value.(Parameter); ok {
			if param.GetPrimitive() != nil {
				return param.GetPrimitive().Equal(another, ctx)
			}
		}
		// naive value
		return ConvertNaiveToPrimitive(value).Equal(another, ctx)
	}
	return false
}

func (p *Primitive) IsNil() bool {
	return p.Value == nil
}

var null = &Primitive{Typ: NULL}
var BoolFalse = &Primitive{Typ: BOOL, Value: false}
var BoolTrue = &Primitive{Typ: BOOL, Value: true}

func (p *Primitive) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	if p.Typ == VARIABLE {
		value := ctx.GetParameters().Get(p.Value.(string))
		if value == nil {
			return null, nil
		}
		// it can convert to parameter
		if param, ok := value.(Parameter); ok {
			if pr := param.GetPrimitive(); pr != nil {
				return pr, nil
			}
			return null, nil
		}
		if pr := ConvertNaiveToPrimitive(value); pr != nil {
			return pr, nil
		}
		return null, nil
	}
	return p, nil
}
