package gonja

import (
	"sync"

	"github.com/aisbergg/gonja/pkg/gonja/builtins"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/loaders"
)

// Environment is the core component of the Gonja template engine. It contains
// important shared variables like configuration, filters, tests, globals and
// others.
type Environment struct {
	*exec.EvalConfig
	loader loaders.Loader

	cache      map[string]*exec.Template
	cacheMutex sync.Mutex
}

// NewEnvironment creates a new [Environment].
func NewEnvironment(options ...Option) *Environment {
	env := &Environment{
		EvalConfig: exec.NewEvalConfig(),
		loader:     loaders.NewNullLoader(),
		cache:      map[string]*exec.Template{},
	}
	env.EvalConfig.TemplateLoadFn = func(name string) (*exec.Template, error) {
		return env.loader.Load(name, env.EvalConfig)
	}
	env.Filters.Update(builtins.Filters)
	env.Statements.Update(builtins.Statements)
	env.Tests.Update(builtins.Tests)
	for k, v := range builtins.Globals {
		env.Globals[k] = v
	}

	// apply user options
	for _, option := range options {
		option(env)
	}
	return env
}

// FromString loads a template from string and returns a Template instance.
func (env *Environment) FromString(tpl string) (*exec.Template, error) {
	return exec.NewTemplate("string", tpl, env.EvalConfig)
}

// FromBytes loads a template from bytes and returns a Template instance.
func (env *Environment) FromBytes(tpl []byte) (*exec.Template, error) {
	return exec.NewTemplate("bytes", string(tpl), env.EvalConfig)
}

// FromFile loads a template from a path and returns a Template instance. It
// uses the configured loader, so make sure you provided a loader that will find
// and load the template.
func (env *Environment) FromFile(path string) (*exec.Template, error) {
	return env.loader.Load(path, env.EvalConfig)
}
