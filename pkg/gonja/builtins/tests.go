package builtins

import (
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

var Tests = exec.TestSet{
	"callable":    testCallable,
	"defined":     testDefined,
	"divisibleby": testDivisibleby,
	"eq":          testEqual,
	"equalto":     testEqual,
	"==":          testEqual,
	// TODO: "escaped": testEscaped,
	"even":        testEven,
	"ge":          testGreaterEqual,
	">=":          testGreaterEqual,
	"gt":          testGreaterThan,
	"greaterthan": testGreaterThan,
	">":           testGreaterThan,
	"in":          testIn,
	"iterable":    testIterable,
	"le":          testLessEqual,
	"<=":          testLessEqual,
	"lower":       testLower,
	"lt":          testLessThan,
	"lessthan":    testLessThan,
	"<":           testLessThan,
	"mapping":     testMapping,
	"ne":          testNotEqual,
	"!=":          testNotEqual,
	"none":        testNone,
	"number":      testNumber,
	"odd":         testOdd,
	"sameas":      testSameas,
	"sequence":    testIterable,
	"string":      testString,
	"undefined":   testUndefined,
	"upper":       testUpper,
}

// testCallable returns true if the input is a callable value.
func testCallable(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	return in.IsCallable()
}

// testDefined returns true if the input is a defined value.
func testDefined(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	return exec.IsDefined(in)
}

// testDivisibleby returns true if the input is divisible by the given number.
func testDivisibleby(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("divisibleby(value, num)", p.Error())
	}
	param := params.Args[0]
	if param.Integer() == 0 {
		return false
	}
	return in.Integer()%param.Integer() == 0
}

func testEqual(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("eq(value, other)", p.Error())
	}
	param := params.Args[0]
	return in.Interface() == param.Interface()
}

func testEven(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	if !in.IsInteger() {
		return false
	}
	return in.Integer()%2 == 0
}

func testGreaterEqual(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("ge(value, other)", p.Error())
	}
	param := params.Args[0]
	if !in.IsNumber() || !param.IsNumber() {
		return false
	}
	return in.Float() >= param.Float()
}

func testGreaterThan(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("gt(value, other)", p.Error())
	}
	param := params.Args[0]
	if !in.IsNumber() || !param.IsNumber() {
		return false
	}
	return in.Float() > param.Float()
}

func testIn(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("in(value, seq)", p.Error())
	}
	seq := params.Args[0]
	return seq.Contains(in)
}

func testIterable(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	return in.IsSliceable()
}

func testLessEqual(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("le(value, other)", p.Error())
	}
	param := params.Args[0]
	if !in.IsNumber() || !param.IsNumber() {
		return false
	}
	return in.Float() <= param.Float()
}

func testLower(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	if !in.IsString() {
		return false
	}
	return strings.ToLower(in.String()) == in.String()
}

func testLessThan(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("lt(value, other)", p.Error())
	}
	param := params.Args[0]
	if !in.IsNumber() || !param.IsNumber() {
		return false
	}
	return in.Float() < param.Float()
}

func testMapping(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	return in.IsDict()
}

func testNotEqual(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("ne(value, other)", p.Error())
	}
	param := params.Args[0]
	return in.Interface() != param.Interface()
}

func testNone(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	return in.IsNil()
}

func testNumber(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	return in.IsNumber()
}

func testOdd(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	if !in.IsInteger() {
		return false
	}
	return in.Integer()%2 == 1
}

// testSameas returns true if the input points to the same memory address as the
// other value.
func testSameas(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("sameas(other)", p.Error())
	}
	param := params.Args[0]
	if in.IsNil() && param.IsNil() {
		return true
	} else if param.ReflectValue().CanAddr() && in.ReflectValue().CanAddr() {
		return param.ReflectValue().Addr() == in.ReflectValue().Addr()
	}
	return false
	// return reflect.Indirect(param.ReflectValue()) == reflect.Indirect(in.ReflectValue())
}

func testString(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	return in.IsString()
}

func testUndefined(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	return !exec.IsDefined(in)
}

func testUpper(ctx *exec.Context, in exec.Value, params *exec.VarArgs) bool {
	if !in.IsString() {
		return false
	}
	return strings.ToUpper(in.String()) == in.String()
}
