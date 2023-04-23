package loaders

import (
	"path/filepath"
	"sync"

	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

// CachedLoader represents a cached loader for templates. It wraps another
// loader and caches the loaded templates.
type CachedLoader struct {
	loader     Loader
	cache      map[string]*exec.Template
	cacheMutex sync.Mutex
}

// NewCachedLoader creates a new [CachedLoader].
func NewCachedLoader(loader Loader) *CachedLoader {
	return &CachedLoader{
		loader: loader,
		cache:  make(map[string]*exec.Template),
	}
}

// Load returns a template by name.
func (fs *CachedLoader) Load(name string, cfg *exec.EvalConfig) (*exec.Template, error) {
	name = filepath.Clean(name)
	fs.cacheMutex.Lock()
	defer fs.cacheMutex.Unlock()

	// check if template is cached
	if t, ok := fs.cache[name]; ok {
		return t, nil
	}

	// load template from wrapped loader
	tpl, err := fs.loader.Load(name, cfg)
	if err != nil {
		return nil, err
	}

	// cache template for later use
	fs.cache[name] = tpl
	return tpl, nil
}

// Clear clears the cache of the loader.
func (fs *CachedLoader) Clear() {
	fs.cacheMutex.Lock()
	defer fs.cacheMutex.Unlock()
	fs.cache = make(map[string]*exec.Template)
}
