package exec

import (
	"fmt"

	debug "github.com/aisbergg/gonja/internal/debug/exec"
)

type Context struct {
	data   map[string]any
	parent *Context

	// user provided data that can take on any type (only used by root context)
	userData     Value
	valueFactory *ValueFactory
}

func NewContext(data map[string]any, userData any, valueFactory *ValueFactory) *Context {
	return &Context{
		data:         data,
		userData:     valueFactory.NewValue(userData, false),
		valueFactory: valueFactory,
	}
}

func EmptyContext() *Context {
	return &Context{data: map[string]any{}}
}

// // setResolver sets the resolver for this context.
// func (ctx *Context) setResolver(valueFactory *ValueFactory) {
// 	ctx.valueFactory = valueFactory
// }

// // setUserData sets the user data for this context.
// func (ctx *Context) setUserData(userData any) {
// 	ctx.userData = ctx.valueFactory.NewValue(userData, false)
// }

func (ctx *Context) Get(name string) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("try to get value for key '%s' from context", name)

	value, exists := ctx.data[name]
	if exists {
		return ctx.valueFactory.NewValue(value, false)
	} else if ctx.parent != nil {
		return ctx.parent.Get(name)
	} else if ctx.userData != nil {
		item := ctx.userData.GetItem(name)
		if _, ok := item.(Undefined); ok {
			item = ctx.valueFactory.NewUndefined(name, fmt.Sprintf("'%s' not found in context", name))
		}
		// save the item in the context so that we do not have to resolve it
		// again
		ctx.data[name] = item
		return item
	}

	return nil
}

func (ctx *Context) Set(name string, value any) {
	ctx.data[name] = value
}

func (ctx *Context) Inherit() *Context {
	return &Context{
		data:         map[string]any{},
		parent:       ctx,
		valueFactory: ctx.valueFactory,
	}
}

// Update updates this context with the key/value pairs from a map.
func (ctx *Context) Update(other map[string]any) *Context {
	for k, v := range other {
		ctx.data[k] = v
	}
	return ctx
}
