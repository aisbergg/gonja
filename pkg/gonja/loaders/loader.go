package loaders

import (
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

// Loader is an interface for loading templates by name.
type Loader interface {
	// Load returns a template by name.
	Load(name string, cfg *exec.EvalConfig) (*exec.Template, error)
}
