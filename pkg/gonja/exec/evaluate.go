package exec

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

var (
	rtListOfAny = reflect.TypeOf([]any{})
	rtError     = reflect.TypeOf((*error)(nil)).Elem()

	evaluatorPool = sync.Pool{
		New: func() interface{} {
			return &Evaluator{}
		},
	}
)

type Evaluator struct {
	*EvalConfig
	Ctx      *Context
	Resolver *Resolver
	Current  parse.Node
}

func (r *Renderer) Evaluator() *Evaluator {
	e := evaluatorPool.Get().(*Evaluator)
	e.EvalConfig = r.EvalConfig
	e.Ctx = r.Ctx
	e.Resolver = r.Resolver
	return e
}

func (r *Renderer) Eval(node parse.Expression) *Value {
	e := r.Evaluator()
	defer func() {
		e.EvalConfig = nil
		e.Ctx = nil
		e.Current = nil
		e.Resolver = nil
		evaluatorPool.Put(e)
	}()
	// enrich runtime errors with token position
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(errors.TemplateRuntimeError); ok {
				rerr.Enrich(parse.AsErrorToken(e.Current.Position()))
				panic(rerr)
			} else {
				panic(r)
			}
		}
	}()

	return e.Eval(node)
}

func (e *Evaluator) Eval(node parse.Expression) *Value {
	e.Current = node
	switch n := node.(type) {
	case *parse.StringNode:
		return AsValue(n.Val)
	case *parse.IntegerNode:
		return AsValue(n.Val)
	case *parse.FloatNode:
		return AsValue(n.Val)
	case *parse.BoolNode:
		return AsValue(n.Val)
	case *parse.ListNode:
		return e.evalList(n)
	case *parse.TupleNode:
		return e.evalTuple(n)
	case *parse.DictNode:
		return e.evalDict(n)
	case *parse.PairNode:
		return e.evalPair(n)
	case *parse.NameNode:
		return e.evalName(n)
	case *parse.CallNode:
		return e.evalCall(n)
	case *parse.GetItemNode:
		return e.evalGetItem(n)
	case *parse.NegationNode:
		return AsValue(!e.Eval(n.Term).IsTrue())
	case *parse.BinaryExpressionNode:
		return e.evalBinaryExpression(n)
	case *parse.UnaryExpressionNode:
		return e.evalUnaryExpression(n)
	case *parse.FilteredExpression:
		return e.EvaluateFiltered(n)
	case *parse.TestExpression:
		return e.EvalTest(n)
	}

	panic(fmt.Errorf("BUG: unknown expression type '%T'", node))
}

func (e *Evaluator) evalBinaryExpression(node *parse.BinaryExpressionNode) *Value {
	var (
		left  *Value
		right *Value
	)
	left = e.Eval(node.Left)

	// lazy right expression evaluation for 'and' and 'or' operations
	if node.Operator.Type != parse.OperatorAnd && node.Operator.Type != parse.OperatorOr {
		right = e.Eval(node.Right)
	}

	switch node.Operator.Type {
	case parse.OperatorAdd:
		if left.IsList() {
			if !right.IsList() {
				e.Current = node.Left
				errors.ThrowTemplateRuntimeError("unable to concatenate list to '%s'", node.Right)
			}
			leftList := reflect.ValueOf(left.IndVal)
			rightList := reflect.ValueOf(right.IndVal)
			newList := reflect.MakeSlice(rtListOfAny, 0, leftList.Len()+rightList.Len())
			for ix := 0; ix < leftList.Len(); ix++ {
				newList = reflect.Append(newList, leftList.Index(ix))
			}
			for ix := 0; ix < rightList.Len(); ix++ {
				newList = reflect.Append(newList, rightList.Index(ix))
			}
			return AsValue(newList.Interface())
		}
		if left.IsFloat() || right.IsFloat() {
			// Result will be a float
			return AsValue(left.Float() + right.Float())
		}
		// Result will be an integer
		return AsValue(left.Integer() + right.Integer())
	case parse.OperatorSub:
		if left.IsFloat() || right.IsFloat() {
			// Result will be a float
			return AsValue(left.Float() - right.Float())
		}
		// Result will be an integer
		return AsValue(left.Integer() - right.Integer())
	case parse.OperatorMul:
		if left.IsFloat() || right.IsFloat() {
			// Result will be float
			return AsValue(left.Float() * right.Float())
		}
		if left.IsString() {
			return AsValue(strings.Repeat(left.String(), right.Integer()))
		}
		// Result will be int
		return AsValue(left.Integer() * right.Integer())
	case parse.OperatorDiv:
		// Float division
		return AsValue(left.Float() / right.Float())
	case parse.OperatorFloordiv:
		// Int division
		return AsValue(int(left.Float() / right.Float()))
	case parse.OperatorMod:
		// Result will be int
		return AsValue(left.Integer() % right.Integer())
	case parse.OperatorPower:
		return AsValue(math.Pow(left.Float(), right.Float()))
	case parse.OperatorConcat:
		return AsValue(strings.Join([]string{left.String(), right.String()}, ""))
	case parse.OperatorAnd:
		if !left.IsTrue() {
			return AsValue(false)
		}
		right = e.Eval(node.Right)
		return AsValue(right.IsTrue())
	case parse.OperatorOr:
		if left.IsTrue() {
			return AsValue(true)
		}
		right = e.Eval(node.Right)
		return AsValue(right.IsTrue())
	case parse.OperatorLteq:
		if left.IsFloat() || right.IsFloat() {
			return AsValue(left.Float() <= right.Float())
		}
		return AsValue(left.Integer() <= right.Integer())
	case parse.OperatorGteq:
		if left.IsFloat() || right.IsFloat() {
			return AsValue(left.Float() >= right.Float())
		}
		return AsValue(left.Integer() >= right.Integer())
	case parse.OperatorEq:
		return AsValue(left.EqualValueTo(right))
	case parse.OperatorGt:
		if left.IsFloat() || right.IsFloat() {
			return AsValue(left.Float() > right.Float())
		}
		return AsValue(left.Integer() > right.Integer())
	case parse.OperatorLt:
		if left.IsFloat() || right.IsFloat() {
			return AsValue(left.Float() < right.Float())
		}
		return AsValue(left.Integer() < right.Integer())
	case parse.OperatorNe:
		return AsValue(!left.EqualValueTo(right))
	case parse.OperatorIn:
		return AsValue(right.Contains(left))
	case parse.OperatorIs:
		return nil
	}

	panic(fmt.Errorf("BUG: unknown operator '%s'", node.Operator.Token))
}

func (e *Evaluator) evalUnaryExpression(expr *parse.UnaryExpressionNode) *Value {
	result := e.Eval(expr.Term)
	if expr.Negative {
		if result.IsNumber() {
			switch {
			case result.IsFloat():
				return AsValue(-1 * result.Float())
			case result.IsInteger():
				return AsValue(-1 * result.Integer())
			default:
				errors.ThrowTemplateRuntimeError("Operation between a number and a non-(float/integer) is not possible")
			}
		} else {
			errors.ThrowTemplateRuntimeError("negative sign on a non-number expression '%s'", expr.Position())
		}
	}
	return result
}

func (e *Evaluator) evalList(node *parse.ListNode) *Value {
	values := ValuesList{}
	for _, val := range node.Val {
		value := e.Eval(val)
		values = append(values, value)
	}
	e.Current = node
	return AsValue(values)
}

func (e *Evaluator) evalTuple(node *parse.TupleNode) *Value {
	values := ValuesList{}
	for _, val := range node.Val {
		value := e.Eval(val)
		values = append(values, value)
	}
	e.Current = node
	return AsValue(values)
}

func (e *Evaluator) evalDict(node *parse.DictNode) *Value {
	pairs := []*Pair{}
	for _, pair := range node.Pairs {
		p := e.evalPair(pair)
		pairs = append(pairs, p.Interface().(*Pair))
	}
	e.Current = node
	return AsValue(&Dict{pairs})
}

func (e *Evaluator) evalPair(node *parse.PairNode) *Value {
	key := e.Eval(node.Key)
	e.Current = node
	value := e.Eval(node.Value)
	e.Current = node
	return AsValue(&Pair{key, value})
}

func (e *Evaluator) evalName(node *parse.NameNode) *Value {
	if node.Name.Val == "none" || node.Name.Val == "None" {
		return AsValue(nil)
	}
	return e.Ctx.Get(node.Name.Val)
}

func (e *Evaluator) evalGetItem(node *parse.GetItemNode) *Value {
	value := e.Eval(node.Node)
	e.Current = node
	if node.Arg != "" {
		item := e.Resolver.Get(value, node.Arg)
		e.Current = node
		return item
	}
	item := e.Resolver.Get(value, node.Index)
	e.Current = node
	return item
}

func (e *Evaluator) evalCall(node *parse.CallNode) *Value {
	fn := e.Eval(node.Func)
	if !fn.IsCallable() {
		errors.ThrowTemplateRuntimeError("'%s' is not callable", node.Func)
	}

	fnType := fn.IndVal.Type()
	numParamsOut := fnType.NumOut()
	if !(numParamsOut == 1 || numParamsOut == 2) {
		errors.ThrowTemplateRuntimeError(
			"function %s must have one (value) or two (value, error) return parameters, not %d",
			node.Func,
			numParamsOut,
		)
	} else if numParamsOut == 2 && fnType.Out(1) != rtError {
		errors.ThrowTemplateRuntimeError(
			"function %s must have an error as second return parameter, not %s",
			node.Func,
			fnType.Out(1),
		)
	}

	var params []reflect.Value
	if fnType.NumIn() == 1 && fnType.In(0) == reflect.TypeOf(&VarArgs{}) {
		params = e.evalVarArgs(node)
	} else {
		params = e.evalParams(node, fn)
	}

	// Call it and get first return parameter back
	values := fn.IndVal.Call(params)
	rv := values[0]
	if numParamsOut == 2 {
		e := values[1].Interface()
		if e != nil {
			err := e.(error)
			if err != nil {
				errors.ThrowTemplateRuntimeError(
					"call of function %s failed: %s",
					node.Func,
					err,
				)
			}
		}
	}

	if rv.Type() == rtValue {
		return rv.Interface().(*Value)
	}
	return AsValue(rv.Interface())
}

func (e *Evaluator) evalVarArgs(node *parse.CallNode) []reflect.Value {
	e.Current = node
	params := NewVarArgs()
	for _, param := range node.Args {
		value := e.Eval(param)
		params.Args = append(params.Args, value)
	}

	for key, param := range node.Kwargs {
		value := e.Eval(param)
		params.SetKwarg(key, value)
	}

	return []reflect.Value{reflect.ValueOf(params)}
}

func (e *Evaluator) evalParams(node *parse.CallNode, fn *Value) []reflect.Value {
	e.Current = node
	args := node.Args
	fnType := fn.IndVal.Type()

	if len(args) != fnType.NumIn() && !(len(args) >= fnType.NumIn()-1 && fnType.IsVariadic()) {
		errors.ThrowTemplateRuntimeError(
			"function input argument count (%d) of '%s' must be equal to the calling argument count (%d)",
			fnType.NumIn(),
			node.String(),
			len(args),
		)
	}

	// Output arguments
	if fnType.NumOut() != 1 && fnType.NumOut() != 2 {
		errors.ThrowTemplateRuntimeError(
			"'%s' must have exactly 1 or 2 output arguments, the second argument must be of type error",
			node.String(),
		)
	}

	// Evaluate all parameters
	var parameters []reflect.Value

	wantNumParams := fnType.NumIn()
	isVariadic := fnType.IsVariadic()
	var wantType reflect.Type

	for idx, arg := range args {
		param := e.Eval(arg)

		if isVariadic && idx >= wantNumParams-1 {
			wantType = fnType.In(wantNumParams - 1).Elem()
		} else {
			wantType = fnType.In(idx)
		}

		// wants the *Value type
		if wantType == rtValue {
			parameters = append(parameters, reflect.ValueOf(param))
			continue
		}

		// wants something else
		paramType := param.Val.Type()
		if wantType != paramType {
			errors.ThrowTemplateRuntimeError(
				"parameter %d of function %s must be of type %s, got %s",
				idx,
				node.String(),
				wantType.String(),
				paramType.String(),
			)
		}
		parameters = append(parameters, param.Val)
	}

	// check if any of the values are invalid
	for idx, param := range parameters {
		if param.Kind() == reflect.Invalid {
			errors.ThrowTemplateRuntimeError(
				"parameter %d of function %s has an invalid value",
				idx,
				node.String(),
			)
		}
	}

	return parameters
}

// GetItem returns the item of the given index or key.
func (e *Evaluator) GetItem(value, key *Value) *Value {
	return e.Resolver.Get(value, key)
}
