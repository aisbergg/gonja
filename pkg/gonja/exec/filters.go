package exec

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// FilterFunction is the type filter functions must fulfil
type FilterFunction func(e *Evaluator, in Value, params *VarArgs) Value

type FilterSet map[string]FilterFunction

// Exists returns true if the given filter is already registered
func (fs FilterSet) Exists(name string) bool {
	_, existing := fs[name]
	return existing
}

// Register registers a new filter. If there's already a filter with the same
// name, Register will panic. You usually want to call this
// function in the filter's init() function:
// http://golang.org/doc/effective_go.html#init
//
// See http://www.florian-schlachter.de/post/gonja/ for more about
// writing filters and tags.
func (fs *FilterSet) Register(name string, fn FilterFunction) error {
	if fs.Exists(name) {
		return fmt.Errorf("filter with name '%s' is already registered", name)
	}
	(*fs)[name] = fn
	return nil
}

// Replace replaces an already registered filter with a new implementation. Use this
// function with caution since it allows you to change existing filter behavior.
func (fs *FilterSet) Replace(name string, fn FilterFunction) error {
	if !fs.Exists(name) {
		return fmt.Errorf("filter with name '%s' does not exist (therefore cannot be overridden)", name)
	}
	(*fs)[name] = fn
	return nil
}

func (fs *FilterSet) Update(other FilterSet) FilterSet {
	for name, filter := range other {
		(*fs)[name] = filter
	}
	return *fs
}

// evalFiltered evaluates a filtered expression.
func (e *Evaluator) evalFiltered(expr *parse.FilteredExpression) Value {
	value := e.Eval(expr.Expression)
	for _, filter := range expr.Filters {
		value = e.ExecuteFilter(filter, value)
	}
	return value
}

// ExecuteFilter execute a filter node
func (e *Evaluator) ExecuteFilter(fc *parse.FilterCall, v Value) Value {
	params := NewVarArgs(e.ValueFactory)

	for _, param := range fc.Args {
		value := e.Eval(param)
		params.Args = append(params.Args, value)
	}

	for key, param := range fc.Kwargs {
		value := e.Eval(param)
		params.SetKwarg(key, value)
	}
	return e.ExecuteFilterByName(fc.Name, v, params)
}

// ExecuteFilterByName execute a filter given its name
func (e *Evaluator) ExecuteFilterByName(name string, in Value, params *VarArgs) Value {
	if !e.Filters.Exists(name) {
		return e.ValueFactory.Value(fmt.Errorf("Filter '%s' not found", name))
	}
	fn := (*e.Filters)[name]

	return fn(e, in, params)
}
