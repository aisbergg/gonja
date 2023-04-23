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

func (r *Renderer) Evaluator() *Evaluator {
	e := evaluatorPool.Get().(*Evaluator)
	e.EvalConfig = r.EvalConfig
	e.Ctx = r.Ctx
	e.ValueFactory = r.ValueVactory
	return e
}

func (r *Renderer) Eval(node parse.Expression) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("eval: %s", node.String())

	e := r.Evaluator()
	defer func() {
		e.EvalConfig = nil
		e.Ctx = nil
		e.Current = nil
		e.ValueFactory = nil
		evaluatorPool.Put(e)
	}()
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
		return e.ValueFactory.NewValue(n.Val, false)
	case *parse.IntegerNode:
		return e.ValueFactory.NewValue(n.Val, false)
	case *parse.FloatNode:
		return e.ValueFactory.NewValue(n.Val, false)
	case *parse.BoolNode:
		return e.ValueFactory.NewValue(n.Val, false)
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
		return e.ValueFactory.NewValue(!e.Eval(n.Term).IsTrue(), false)
	case *parse.BinaryExpressionNode:
		return e.evalBinaryExpression(n)
	case *parse.UnaryExpressionNode:
		return e.evalUnaryExpression(n)
	case *parse.FilteredExpression:
		return e.EvaluateFiltered(n)
	case *parse.TestExpression:
		return e.EvalTest(n)
	}

	panic(fmt.Errorf("[BUG] unknown expression type '%T'", node))
}

func (e *Evaluator) evalBinaryExpression(node *parse.BinaryExpressionNode) Value {
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
			return e.ValueFactory.NewValue(newList.Interface(), false)
		}
		if left.IsFloat() || right.IsFloat() {
			// Result will be a float
			return e.ValueFactory.NewValue(left.Float()+right.Float(), false)
		}
		// Result will be an integer
		return e.ValueFactory.NewValue(left.Integer()+right.Integer(), false)
	case parse.OperatorSub:
		if left.IsFloat() || right.IsFloat() {
			// Result will be a float
			return e.ValueFactory.NewValue(left.Float()-right.Float(), false)
		}
		// Result will be an integer
		return e.ValueFactory.NewValue(left.Integer()-right.Integer(), false)
	case parse.OperatorMul:
		if left.IsFloat() || right.IsFloat() {
			// Result will be float
			return e.ValueFactory.NewValue(left.Float()*right.Float(), false)
		}
		if left.IsString() {
			return e.ValueFactory.NewValue(strings.Repeat(left.String(), right.Integer()), false)
		}
		// Result will be int
		return e.ValueFactory.NewValue(left.Integer()*right.Integer(), false)
	case parse.OperatorDiv:
		// Float division
		return e.ValueFactory.NewValue(left.Float()/right.Float(), false)
	case parse.OperatorFloordiv:
		// Int division
		return e.ValueFactory.NewValue(int(left.Float()/right.Float()), false)
	case parse.OperatorMod:
		// Result will be int
		return e.ValueFactory.NewValue(left.Integer()%right.Integer(), false)
	case parse.OperatorPower:
		return e.ValueFactory.NewValue(math.Pow(left.Float(), right.Float()), false)
	case parse.OperatorConcat:
		return e.ValueFactory.NewValue(strings.Join([]string{left.String(), right.String()}, ""), false)
	case parse.OperatorAnd:
		if !left.IsTrue() {
			return e.ValueFactory.NewValue(false, false)
		}
		right = e.Eval(node.Right)
		return e.ValueFactory.NewValue(right.IsTrue(), false)
	case parse.OperatorOr:
		if left.IsTrue() {
			return e.ValueFactory.NewValue(true, false)
		}
		right = e.Eval(node.Right)
		return e.ValueFactory.NewValue(right.IsTrue(), false)
	case parse.OperatorLteq:
		if left.IsFloat() || right.IsFloat() {
			return e.ValueFactory.NewValue(left.Float() <= right.Float(), false)
		}
		return e.ValueFactory.NewValue(left.Integer() <= right.Integer(), false)
	case parse.OperatorGteq:
		if left.IsFloat() || right.IsFloat() {
			return e.ValueFactory.NewValue(left.Float() >= right.Float(), false)
		}
		return e.ValueFactory.NewValue(left.Integer() >= right.Integer(), false)
	case parse.OperatorEq:
		return e.ValueFactory.NewValue(left.EqualValueTo(right), false)
	case parse.OperatorGt:
		if left.IsFloat() || right.IsFloat() {
			return e.ValueFactory.NewValue(left.Float() > right.Float(), false)
		}
		return e.ValueFactory.NewValue(left.Integer() > right.Integer(), false)
	case parse.OperatorLt:
		if left.IsFloat() || right.IsFloat() {
			return e.ValueFactory.NewValue(left.Float() < right.Float(), false)
		}
		return e.ValueFactory.NewValue(left.Integer() < right.Integer(), false)
	case parse.OperatorNe:
		return e.ValueFactory.NewValue(!left.EqualValueTo(right), false)
	case parse.OperatorIn:
		return e.ValueFactory.NewValue(right.Contains(left), false)
	case parse.OperatorIs:
		return nil
	}

	panic(fmt.Errorf("[BUG] unknown operator '%s'", node.Operator.Token))
}

func (e *Evaluator) evalUnaryExpression(expr *parse.UnaryExpressionNode) Value {
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
				return e.ValueFactory.NewValue(-1*result.Float(), false)
			case result.IsInteger():
				return e.ValueFactory.NewValue(-1*result.Integer(), false)
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
	return e.ValueFactory.NewValue(values, false)
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
	return e.ValueFactory.NewValue(values, false)
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
	return e.ValueFactory.NewValue(&Dict{pairs}, false)
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
	return e.ValueFactory.NewValue(&Pair{key, value}, false)
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
	return e.ValueFactory.NewValue(rv.Interface(), false)
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
		paramType := param.ReflectValue().Type()
		if wantType != paramType {
			errors.ThrowTemplateRuntimeError(
				"parameter %d of function %s must be of type %s, got %s",
				idx,
				node.String(),
				wantType.String(),
				paramType.String(),
			)
		}
		parameters = append(parameters, param.ReflectValue())
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
