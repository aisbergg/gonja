package exec

import (
	"fmt"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// FilterFunction is the type filter functions must fulfil
type Macro func(params *VarArgs) Value

type MacroSet map[string]Macro

// Exists returns true if the given filter is already registered
func (ms MacroSet) Exists(name string) bool {
	_, existing := ms[name]
	return existing
}

// Register registers a new filter. If there's already a filter with the same
// name, Register will panic. You usually want to call this
// function in the filter's init() function:
// http://golang.org/doc/effective_go.html#init
func (ms *MacroSet) Register(name string, fn Macro) error {
	if ms.Exists(name) {
		return fmt.Errorf("filter with name '%s' is already registered", name)
	}
	(*ms)[name] = fn
	return nil
}

// Replace replaces an already registered filter with a new implementation. Use
// this function with caution since it allows you to change existing filter
// behavior.
func (ms *MacroSet) Replace(name string, fn Macro) error {
	if !ms.Exists(name) {
		return fmt.Errorf("filter with name '%s' does not exist (therefore cannot be overridden)", name)
	}
	(*ms)[name] = fn
	return nil
}

func MacroNodeToFunc(node *parse.MacroNode, r *Renderer) Macro {
	// Compute default values once
	defaultKwargs := []*Kwarg{}
	for _, pair := range node.Kwargs {
		key := r.Eval(pair.Key).String()
		value := r.Eval(pair.Value)
		defaultKwargs = append(defaultKwargs, &Kwarg{key, value.Interface()})
	}

	return func(params *VarArgs) Value {
		var out strings.Builder
		sub := r.Inherit()
		sub.Out = &out
		p := params.Expect(len(node.Args), defaultKwargs)
		if p.IsError() {
			errors.ThrowTemplateAssertionError("wrong '%s' macro signature: %s", node.Name, p.Error())
		}
		for idx, arg := range p.Args {
			sub.Ctx.Set(node.Args[idx], arg)
		}
		for _, kv := range p.Kwargs {
			sub.Ctx.Set(kv.Key, kv.Value)
		}
		err := sub.ExecuteWrapper(node.Wrapper)
		if err != nil {
			// pass error up the call stack
			panic(err)
		}
		return r.ValueFactory.SafeValue(out.String())
	}
}
