package ast

import "strings"

type compareFn[T any] func(l, r T) int

func gte[T any](cmp compareFn[T]) EvalBool[T] {
	return func(l, r T) bool {
		return cmp(l, r) >= 0
	}
}

func gt[T any](cmp compareFn[T]) EvalBool[T] {
	return func(l, r T) bool {
		return cmp(l, r) > 0
	}
}

func lt[T any](cmp compareFn[T]) EvalBool[T] {
	return func(l, r T) bool {
		return cmp(l, r) < 0
	}
}

func lte[T any](cmp compareFn[T]) EvalBool[T] {
	return func(l, r T) bool {
		return cmp(l, r) <= 0
	}
}

func ne[T any](cmp compareFn[T]) EvalBool[T] {
	return func(l, r T) bool {
		return cmp(l, r) != 0
	}
}

func eq[T any](cmp compareFn[T]) EvalBool[T] {
	return func(l, r T) bool {
		return cmp(l, r) == 0
	}
}

func compareString(l, r string) int {
	return strings.Compare(l, r)
}

func compareBool(l, r bool) int {
	if l == r {
		return 0
	}
	if !l {
		return -1
	}
	return 1
}

func compareFloat(l, r float64) int {
	if l == r {
		return 0
	} else if l < r {
		return -1
	}
	return 1
}

func compareInt(l, r int) int {
	if l == r {
		return 0
	} else if l < r {
		return -1
	}
	return 1
}

type EvalBool[T any] func(l, r T) bool

type BoolEvalMap map[Op]BoolEvalFnGroup

type BoolEvalFnGroup struct {
	evalInt    EvalBool[int]
	evalFloat  EvalBool[float64]
	evalBool   EvalBool[bool]
	evalString EvalBool[string]
}

var defaultBoolEvalMap = map[Op]BoolEvalFnGroup{
	EQ: {
		evalInt:    eq(compareInt),
		evalFloat:  eq(compareFloat),
		evalBool:   eq(compareBool),
		evalString: eq(compareString),
	},
	NE: {
		evalInt:    ne(compareInt),
		evalFloat:  ne(compareFloat),
		evalBool:   ne(compareBool),
		evalString: ne(compareString),
	},
	LT: {
		evalInt:    lt(compareInt),
		evalFloat:  lt(compareFloat),
		evalBool:   lt(compareBool),
		evalString: lt(compareString),
	},
	LTE: {
		evalInt:    lte(compareInt),
		evalFloat:  lte(compareFloat),
		evalBool:   lte(compareBool),
		evalString: lte(compareString),
	},
	GT: {
		evalInt:    gt(compareInt),
		evalFloat:  gt(compareFloat),
		evalBool:   gt(compareBool),
		evalString: gt(compareString),
	},
	GTE: {
		evalInt:    gte(compareInt),
		evalFloat:  gte(compareFloat),
		evalBool:   gte(compareBool),
		evalString: gte(compareString),
	},
}
