package gonja

import (
	"fmt"
	"io/ioutil"
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

// NewEnvironment creates a new Environment instance.
func NewEnvironment(loader loaders.Loader, options ...Option) *Environment {
	env := &Environment{
		EvalConfig: exec.NewEvalConfig(),
		loader:     loader,
		cache:      map[string]*exec.Template{},
	}
	env.EvalConfig.Loader = env
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

// CleanCache cleans the template cache. If filenames is not empty,
// it will remove the template caches of those filenames.
// Or it will empty the whole template cache. It is thread-safe.
func (env *Environment) CleanCache(filenames ...string) {
	env.cacheMutex.Lock()
	defer env.cacheMutex.Unlock()

	if len(filenames) == 0 {
		env.cache = map[string]*exec.Template{}
	}

	for _, filename := range filenames {
		delete(env.cache, filename)
	}
}

// FromCache is a convenient method to cache templates. It is thread-safe
// and will only compile the template associated with a filename once.
func (env *Environment) FromCache(filename string) (*exec.Template, error) {
	env.cacheMutex.Lock()
	defer env.cacheMutex.Unlock()

	tpl, has := env.cache[filename]

	// Cache miss
	if !has {
		tpl, err := env.FromFile(filename)
		if err != nil {
			return nil, err
		}
		env.cache[filename] = tpl
		return tpl, nil
	}

	// Cache hit
	return tpl, nil
}

// FromString loads a template from string and returns a Template instance.
func (env *Environment) FromString(tpl string) (*exec.Template, error) {
	return exec.NewTemplate("string", tpl, env.EvalConfig)
}

// FromBytes loads a template from bytes and returns a Template instance.
func (env *Environment) FromBytes(tpl []byte) (*exec.Template, error) {
	return exec.NewTemplate("bytes", string(tpl), env.EvalConfig)
}

// FromFile loads a template from a filename and returns a Template instance.
func (env *Environment) FromFile(filename string) (*exec.Template, error) {
	fd, err := env.loader.Get(filename)
	if err != nil {
		// TODO: return loader error
		return nil, fmt.Errorf("error loading template %s: %w", filename, err)
	}
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, fmt.Errorf("error loading template %s: %w", filename, err)
	}

	return exec.NewTemplate(filename, string(buf), env.EvalConfig)
}

// GetTemplate returns a template for the given filename.
func (env *Environment) GetTemplate(filename string) (*exec.Template, error) {
	return env.FromFile(filename)
}
