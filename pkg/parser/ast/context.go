package ast

type Parameter interface {
	GetPrimitive() *Primitive
}

type Parameters interface {
	Get(key string) interface{}
}

type Function interface {
	Eval(args ...*Primitive) (*Primitive, error)
}

type NaiveFunction interface {
	NaiveEval(args ...interface{}) (interface{}, error)
}

type FunctionWithCtx interface {
	Eval(ctx EvaluateCtx, args ...*Primitive) (*Primitive, error)
}

type NaiveFunctionWithCtx interface {
	NaiveEval(ctx EvaluateCtx, args ...interface{}) (interface{}, error)
}

type Functions interface {
	Get(key string) interface{}
}

type FunctionSet map[string]interface{}

func (fs FunctionSet) Get(key string) interface{} {
	return fs[key]
}

func (fs FunctionSet) AddFunction(k string, f Function) {
	fs[k] = f
}

func (fs FunctionSet) AddFunctionWithCtx(k string, f FunctionWithCtx) {
	fs[k] = f
}

func (fs FunctionSet) AddNaiveFunction(k string, f NaiveFunction) {
	fs[k] = f
}

func (fs FunctionSet) AddNaiveFunctionWithCtx(k string, f NaiveFunctionWithCtx) {
	fs[k] = f
}

type ParameterSet map[string]interface{}

func (ps ParameterSet) Get(key string) interface{} {
	return ps[key]
}

func (ps ParameterSet) AddNaiveParameter(k string, p interface{}) {
	ps[k] = p
}

func (ps ParameterSet) AddParameter(k string, p Primitive) {
	ps[k] = p
}

type Context struct {
	functions  FunctionSet
	parameters ParameterSet
}

func (c Context) GetParameters() Parameters {
	return c.parameters
}

func (c Context) GetFunctions() Functions {
	return c.functions
}

func NewContext() EvaluateCtx {
	return &Context{
		functions:  make(FunctionSet),
		parameters: make(ParameterSet),
	}
}
