package exec

import log "github.com/aisbergg/gonja/internal/log/exec"

type Context struct {
	data   map[string]any
	parent *Context

	// user provided data that can take on any type (only used by root context)
	userData *Value
	resolver *Resolver
}

func NewContext(data map[string]any, userData any, resolver *Resolver) *Context {
	return &Context{
		data:     data,
		userData: ToValue(userData),
		resolver: resolver,
	}
}

func EmptyContext() *Context {
	return &Context{data: map[string]any{}}
}

// setResolver sets the resolver for this context.
func (ctx *Context) setResolver(resolver *Resolver) {
	ctx.resolver = resolver
}

// setUserData sets the user data for this context.
func (ctx *Context) setUserData(userData any) {
	ctx.userData = ToValue(userData)
}

func (ctx *Context) Get(name string) *Value {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("try to get value for key '%s' from context", name)

	value, exists := ctx.data[name]
	if exists {
		return ToValue(value)
	} else if ctx.parent != nil {
		return ctx.parent.Get(name)
	} else if ctx.resolver != nil {
		item := ctx.resolver.Get(ctx.userData, name)
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
		data:   map[string]any{},
		parent: ctx,
	}
}

// Update updates this context with the key/value pairs from a map.
func (ctx *Context) Update(other map[string]any) *Context {
	for k, v := range other {
		ctx.data[k] = v
	}
	return ctx
}
