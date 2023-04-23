package loaders

import (
	"path/filepath"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

// MergedLoader represents a merged loader for templates. It wraps multiple
// loaders and returns the first found template.
type MergedLoader struct {
	loaders []Loader
}

// NewMergedLoader creates a new [MergedLoader].
func NewMergedLoader(loaders ...Loader) *MergedLoader {
	return &MergedLoader{
		loaders: loaders,
	}
}

// Load returns a template by name.
func (fs *MergedLoader) Load(name string, cfg *exec.EvalConfig) (*exec.Template, error) {
	name = filepath.Clean(name)

	for _, loader := range fs.loaders {
		tpl, err := loader.Load(name, cfg)
		if err != nil {
			continue
		}
		return tpl, nil
	}

	return nil, errors.NewTemplateNotFoundError(name)
}
