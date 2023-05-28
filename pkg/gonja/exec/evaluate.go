package exec

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"

	debug "github.com/aisbergg/gonja/internal/debug/exec"
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
	Ctx          *Context
	ValueFactory *ValueFactory
	Current      parse.Node
}

// func (r *Renderer) Evaluator() *Evaluator {
// 	e := evaluatorPool.Get().(*Evaluator)
// 	e.EvalConfig = r.EvalConfig
// 	e.Ctx = r.Ctx
// 	e.ValueFactory = r.ValueFactory
// 	return e
// }

func (r *Renderer) Evaluator() *Evaluator {
	e := &Evaluator{
		EvalConfig:   r.EvalConfig,
		Ctx:          r.Ctx,
		ValueFactory: r.ValueFactory,
	}
	return e
}

func (r *Renderer) Eval(node parse.Expression) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	e := r.Evaluator()
	// defer func() {
	// 	e.EvalConfig = nil
	// 	e.Ctx = nil
	// 	e.Current = nil
	// 	e.ValueFactory = nil
	// 	evaluatorPool.Put(e)
	// }()
	// enrich runtime errors with token position
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(errors.TemplateRuntimeError); ok {
				rerr.Enrich(e.Current.Position().ErrorToken())
				panic(rerr)
			} else {
				panic(r)
			}
		}
	}()

	return e.Eval(node)
}

func (e *Evaluator) Eval(node parse.Expression) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	e.Current = node
	switch n := node.(type) {
	case *parse.StringNode:
		return e.ValueFactory.Value(n.Val)
	case *parse.IntegerNode:
		return e.ValueFactory.Value(n.Val)
	case *parse.FloatNode:
		return e.ValueFactory.Value(n.Val)
	case *parse.BoolNode:
		return e.ValueFactory.Value(n.Val)
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
		return e.ValueFactory.Value(!e.Eval(n.Term).Bool())
	case *parse.BinaryExpressionNode:
		return e.evalBinary(n)
	case *parse.UnaryExpressionNode:
		return e.evalUnary(n)
	case *parse.FilteredExpression:
		return e.evalFiltered(n)
	case *parse.TestExpression:
		return e.evalTest(n)
	case *parse.InlineIfExpressionNode:
		return e.evalInlineIf(n)
	}

	panic(fmt.Errorf("[BUG] unknown expression type '%T'", node))
}

func (e *Evaluator) evalBinary(node *parse.BinaryExpressionNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	var (
		left  Value
		right Value
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

			leftList := indirectReflectValue(left.ReflectValue())
			rightList := indirectReflectValue(right.ReflectValue())
			newList := reflect.MakeSlice(rtListOfAny, 0, leftList.Len()+rightList.Len())
			for ix := 0; ix < leftList.Len(); ix++ {
				newList = reflect.Append(newList, leftList.Index(ix))
			}
			for ix := 0; ix < rightList.Len(); ix++ {
				newList = reflect.Append(newList, rightList.Index(ix))
			}
			return e.ValueFactory.Value(newList.Interface())
		}
		if left.IsFloat() || right.IsFloat() {
			// Result will be a float
			return e.ValueFactory.Value(left.Float() + right.Float())
		}
		// Result will be an integer
		return e.ValueFactory.Value(left.Integer() + right.Integer())
	case parse.OperatorSub:
		if left.IsFloat() || right.IsFloat() {
			// Result will be a float
			return e.ValueFactory.Value(left.Float() - right.Float())
		}
		// Result will be an integer
		return e.ValueFactory.Value(left.Integer() - right.Integer())
	case parse.OperatorMul:
		if left.IsFloat() || right.IsFloat() {
			// Result will be float
			return e.ValueFactory.Value(left.Float() * right.Float())
		}
		if left.IsString() {
			return e.ValueFactory.Value(strings.Repeat(left.String(), right.Integer()))
		}
		// Result will be int
		return e.ValueFactory.Value(left.Integer() * right.Integer())
	case parse.OperatorDiv:
		// Float division
		return e.ValueFactory.Value(left.Float() / right.Float())
	case parse.OperatorFloordiv:
		// Int division
		return e.ValueFactory.Value(int(left.Float() / right.Float()))
	case parse.OperatorMod:
		// Result will be int
		return e.ValueFactory.Value(left.Integer() % right.Integer())
	case parse.OperatorPower:
		return e.ValueFactory.Value(math.Pow(left.Float(), right.Float()))
	case parse.OperatorConcat:
		return e.ValueFactory.Value(strings.Join([]string{left.String(), right.String()}, ""))
	case parse.OperatorAnd:
		if !left.Bool() {
			return e.ValueFactory.Value(false)
		}
		right = e.Eval(node.Right)
		return e.ValueFactory.Value(right.Bool())
	case parse.OperatorOr:
		if left.Bool() {
			return e.ValueFactory.Value(true)
		}
		right = e.Eval(node.Right)
		return e.ValueFactory.Value(right.Bool())
	case parse.OperatorLteq:
		if left.IsFloat() || right.IsFloat() {
			return e.ValueFactory.Value(left.Float() <= right.Float())
		}
		return e.ValueFactory.Value(left.Integer() <= right.Integer())
	case parse.OperatorGteq:
		if left.IsFloat() || right.IsFloat() {
			return e.ValueFactory.Value(left.Float() >= right.Float())
		}
		return e.ValueFactory.Value(left.Integer() >= right.Integer())
	case parse.OperatorEq:
		return e.ValueFactory.Value(left.EqualValueTo(right))
	case parse.OperatorGt:
		if left.IsFloat() || right.IsFloat() {
			return e.ValueFactory.Value(left.Float() > right.Float())
		}
		return e.ValueFactory.Value(left.Integer() > right.Integer())
	case parse.OperatorLt:
		if left.IsFloat() || right.IsFloat() {
			return e.ValueFactory.Value(left.Float() < right.Float())
		}
		return e.ValueFactory.Value(left.Integer() < right.Integer())
	case parse.OperatorNe:
		return e.ValueFactory.Value(!left.EqualValueTo(right))
	case parse.OperatorIn:
		return e.ValueFactory.Value(right.Contains(left))
	case parse.OperatorIs:
		return nil
	}

	panic(fmt.Errorf("[BUG] unknown operator '%s'", node.Operator.Token))
}

func (e *Evaluator) evalUnary(expr *parse.UnaryExpressionNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", expr.String())

	result := e.Eval(expr.Term)
	if expr.Negative {
		if result.IsNumber() {
			switch {
			case result.IsFloat():
				return e.ValueFactory.Value(-1 * result.Float())
			case result.IsInteger():
				return e.ValueFactory.Value(-1 * result.Integer())
			default:
				errors.ThrowTemplateRuntimeError("Operation between a number and a non-(float/integer) is not possible")
			}
		} else {
			errors.ThrowTemplateRuntimeError("negative sign on a non-number expression '%s'", expr.Position())
		}
	}
	return result
}

func (e *Evaluator) evalList(node *parse.ListNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	values := ValuesList{}
	for _, val := range node.Val {
		value := e.Eval(val)
		values = append(values, value)
	}
	e.Current = node
	return e.ValueFactory.Value(values)
}

func (e *Evaluator) evalTuple(node *parse.TupleNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	values := ValuesList{}
	for _, val := range node.Val {
		value := e.Eval(val)
		values = append(values, value)
	}
	e.Current = node
	return e.ValueFactory.Value(values)
}

func (e *Evaluator) evalDict(node *parse.DictNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	pairs := []*Pair{}
	for _, pair := range node.Pairs {
		p := e.evalPair(pair)
		pairs = append(pairs, p.Interface().(*Pair))
	}
	e.Current = node
	return e.ValueFactory.Value(&Dict{pairs})
}

func (e *Evaluator) evalPair(node *parse.PairNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	key := e.Eval(node.Key)
	e.Current = node
	value := e.Eval(node.Value)
	e.Current = node
	return e.ValueFactory.Value(&Pair{key, value})
}

func (e *Evaluator) evalName(node *parse.NameNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	switch node.Name.Val {
	case "None", "none", "Nil", "nil":
		return NewNilValue()
	}
	return e.Ctx.Get(node.Name.Val)
}

func (e *Evaluator) evalGetItem(node *parse.GetItemNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	value := e.Eval(node.Node)
	e.Current = node
	if node.Arg != "" {
		item := value.GetItem(node.Arg)
		e.Current = node
		return item
	}
	item := value.GetItem(node.Index)
	e.Current = node
	return item
}

func (e *Evaluator) evalCall(node *parse.CallNode) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	fn := e.Eval(node.Func)
	if !fn.IsCallable() {
		errors.ThrowTemplateRuntimeError("'%s' is not callable", fn.String())
	}

	fnVal := indirectReflectValue(fn.ReflectValue())
	fnType := fnVal.Type()
	numParamsOut := fnType.NumOut()
	if !(numParamsOut == 1 || numParamsOut == 2) {
		fnName := ""
		if nameNode, ok := node.Func.(*parse.NameNode); ok {
			fnName = nameNode.Name.Val
		} else {
			fnName = fn.String()
		}
		errors.ThrowTemplateRuntimeError(
			"function %s must have one (value) or two (value, error) return parameters, not %d",
			fnName,
			numParamsOut,
		)
	} else if numParamsOut == 2 && fnType.Out(1) != rtError {
		fnName := ""
		if nameNode, ok := node.Func.(*parse.NameNode); ok {
			fnName = nameNode.Name.Val
		} else {
			fnName = fn.String()
		}
		errors.ThrowTemplateRuntimeError(
			"function %s must have an error as second return parameter, not %s",
			fnName,
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
	values := fnVal.Call(params)
	rv := values[0]
	if numParamsOut == 2 {
		e := values[1].Interface()
		if e != nil {
			err := e.(error)
			if err != nil {
				fnName := ""
				if nameNode, ok := node.Func.(*parse.NameNode); ok {
					fnName = nameNode.Name.Val
				} else {
					fnName = fn.String()
				}
				errors.ThrowTemplateRuntimeError(
					"call of function %s failed: %s",
					fnName,
					err,
				)
			}
		}
	}

	if rv.Type() == rtValue {
		return rv.Interface().(Value)
	}
	return e.ValueFactory.Value(rv.Interface())
}

func (e *Evaluator) evalVarArgs(node *parse.CallNode) []reflect.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	e.Current = node
	params := NewVarArgs(e.ValueFactory)
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

func (e *Evaluator) evalParams(node *parse.CallNode, fn Value) []reflect.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	e.Current = node
	args := node.Args
	fnType := indirectReflectValue(fn.ReflectValue()).Type()

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

		// if the parameter is variadic (...type), the last parameters are all
		// of the same type
		if isVariadic && idx >= wantNumParams-1 {
			wantType = fnType.In(wantNumParams - 1).Elem()
		} else {
			wantType = fnType.In(idx)
		}

		// wants the Value type
		if wantType == rtValue {
			parameters = append(parameters, reflect.ValueOf(param))
			continue
		}

		// wants something else
		paramRV := param.ReflectValue()
		paramType := paramRV.Type()

		if wantType != paramType {
			// try to convert the parameter to the wanted type
			for {
				if paramType.ConvertibleTo(wantType) {
					paramRV = paramRV.Convert(wantType)
					paramType = paramRV.Type()
					break
				}
				if paramType.Kind() == reflect.Interface || paramType.Kind() == reflect.Ptr {
					// try to unwrap
					paramRV = paramRV.Elem()
					paramType = paramRV.Type()
					continue
				}
				break
			}
		}

		if wantType != paramType {
			errors.ThrowTemplateRuntimeError(
				"parameter %d of function %s must be of type %s, got %s",
				idx,
				node.String(),
				wantType.String(),
				paramType.String(),
			)
		}
		parameters = append(parameters, paramRV)
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

// evalInlineIf evaluates an inline if expression.
func (e *Evaluator) evalInlineIf(expr *parse.InlineIfExpressionNode) Value {
	condition := e.Eval(expr.Condition)
	if condition.Bool() {
		return e.Eval(expr.TrueExpr)
	}
	return e.Eval(expr.FalseExpr)
}
