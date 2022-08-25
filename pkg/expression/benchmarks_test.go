package expression

import (
	"github.com/Knetic/govaluate"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/parser"
	"strings"
	"testing"
)

/*
  Serves as a "water test" to give an idea of the general overhead of parsing
*/
func BenchmarkSingleParse_govaluate(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		govaluate.NewEvaluableExpression("1")
	}
}

func BenchmarkSingleParse(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		s := parser.NewLexer(strings.NewReader("1"))
		parser.Parse(s)
	}
}

/*
  Benchmarks the bare-minimum evaluation time
*/
func BenchmarkEvaluationSingle_govaluate(bench *testing.B) {
	expression, _ := govaluate.NewEvaluableExpression("1")
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

func BenchmarkEvaluationSingle(bench *testing.B) {
	s := parser.NewLexer(strings.NewReader("1"))
	parser.Parse(s)
	expression := parser.GetParseResult(s).(ast.Evaluable)
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

/*
  Benchmarks evaluation times of literals (no variables, no modifiers)
*/
func BenchmarkEvaluationNumericLiteral_govaluate(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("(2) > (1)")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

func BenchmarkEvaluationNumericLiteral(bench *testing.B) {
	s := parser.NewLexer(strings.NewReader("(2) > (1)"))
	parser.Parse(s)
	expression := parser.GetParseResult(s).(ast.Evaluable)
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

/*
  Benchmarks evaluation times of literals with modifiers
*/
func BenchmarkEvaluationLiteralModifiers_govaluate(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("(2) + (2) == (4)")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

func BenchmarkEvaluationLiteralModifiers(bench *testing.B) {

	s := parser.NewLexer(strings.NewReader("(2) + (2) == (4)"))
	parser.Parse(s)
	expression := parser.GetParseResult(s).(ast.Evaluable)
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

/*
  Benchmarks evaluation times of parameters
*/
func BenchmarkEvaluationParameters_govaluate(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("requests_made > requests_succeeded")
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkEvaluationParameters(bench *testing.B) {

	s := parser.NewLexer(strings.NewReader("requests_made > requests_succeeded"))
	parser.Parse(s)
	expression := parser.GetParseResult(s).(ast.Evaluable)
	ctx := ast.NewContext()
	ctx.AddParameter("requests_made", ast.Primitive{Typ: ast.INT, Value: 99})
	ctx.AddParameter("requests_succeeded", ast.Primitive{Typ: ast.INT, Value: 90})

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(ctx)
	}
}

/*
  Benchmarks evaluation times of parameters + literals with modifiers
*/
func BenchmarkEvaluationParametersModifiers_govaluate(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkEvaluationParametersModifiers(bench *testing.B) {

	s := parser.NewLexer(strings.NewReader("(requests_made * requests_succeeded / 100) >= 90"))
	parser.Parse(s)
	expression := parser.GetParseResult(s).(ast.Evaluable)
	ctx := ast.NewContext()
	ctx.AddParameter("requests_made", ast.Primitive{Typ: ast.INT, Value: 99})
	ctx.AddParameter("requests_succeeded", ast.Primitive{Typ: ast.INT, Value: 90})

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(ctx)
	}
}

/*
  Benchmarks evaluation times of functions + parameters + literals with modifiers
*/
func BenchmarkEvaluationFunction_govaluate(bench *testing.B) {

	functions := map[string]govaluate.ExpressionFunction{}
	ret := 90
	functions["get_requests_made"] = func(arguments ...interface{}) (interface{}, error) {
		return ret, nil
	}
	expression, _ := govaluate.NewEvaluableExpressionWithFunctions("(get_requests_made() * requests_succeeded / 100) >= 90", functions)
	parameters := map[string]interface{}{
		"requests_succeeded": 90.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

type mockFunc struct {
	fn func(args ...ast.Evaluable) (*ast.Primitive, error)
}

func (f *mockFunc) Eval(ctx ast.EvaluateCtx, args ...ast.Evaluable) (*ast.Primitive, error) {
	return f.fn(args...)
}

func BenchmarkEvaluationFunction(bench *testing.B) {
	s := parser.NewLexer(strings.NewReader("(get_requests_made() * requests_succeeded / 100) >= 90"))
	parser.Parse(s)
	expression := parser.GetParseResult(s).(ast.Evaluable)
	ctx := ast.NewContext()
	ctx.AddParameter("requests_succeeded", ast.Primitive{Typ: ast.INT, Value: 90})

	ret := &ast.Primitive{Typ: ast.INT, Value: 99}
	fn1 := mockFunc{
		fn: func(args ...ast.Evaluable) (*ast.Primitive, error) {
			return ret, nil
		},
	}
	ctx.AddFunctionWithCtx("get_requests_made", &fn1)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(ctx)
	}
}

func BenchmarkEvaluationFunction_Naive(bench *testing.B) {
	ret := 99
	get_requests_made := func() int {
		return ret
	}
	requests_succeeded := 90
	expression := func() bool {
		return (get_requests_made() * requests_succeeded / 100) >= 90
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression()
	}
}

func BenchmarkEvaluationFunctionOnly(bench *testing.B) {
	s := parser.NewLexer(strings.NewReader("(get_requests_made() * requests_succeeded() / 100) >= 90"))
	parser.Parse(s)
	expression := parser.GetParseResult(s).(ast.Evaluable)
	ctx := ast.NewContext()

	ret := &ast.Primitive{Typ: ast.INT, Value: 99}
	ret2 := &ast.Primitive{Typ: ast.INT, Value: 90}
	fn1 := mockFunc{
		fn: func(args ...ast.Evaluable) (*ast.Primitive, error) {
			return ret, nil
		},
	}

	fn2 := mockFunc{
		fn: func(args ...ast.Evaluable) (*ast.Primitive, error) {
			return ret2, nil
		},
	}
	ctx.AddFunctionWithCtx("get_requests_made", &fn1)
	ctx.AddFunctionWithCtx("requests_succeeded", &fn2)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(ctx)
	}
}
