package loaders

import (
	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

// NullLoader represents a loader that refuses to load anything.
type NullLoader struct{}

// NewNullLoader creates a new [NullLoader].
func NewNullLoader() *NullLoader {
	return &NullLoader{}
}

// Load returns a template by name.
func (fs *NullLoader) Load(name string, _ *exec.EvalConfig) (*exec.Template, error) {
	return nil, errors.NewTemplateLoadError(name, "no loader for this environment specified")
}
