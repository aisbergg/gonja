package gonja

import (
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/loaders"
)

var (
	// DefaultEnv is an environment created for quick/standalone template
	// rendering. It uses the NullLoader, which means that you can't use
	// `include` or `extends` to load other templates.
	DefaultEnv       = NewEnvironment()
	fileSystemLoader = loaders.MustNewFileSystemLoader("")

	// FromString is a quick way to parse a template from a string. The template
	// doesn't allow any includes. If you want to use includes, create a custom
	// environment with an appropriate loader.
	FromString = DefaultEnv.FromString
	// FromBytes is a quick way to parse a template from a byte slice. The
	// template doesn't allow any includes. If you want to use includes, create
	// a custom environment with an appropriate loader.
	FromBytes = DefaultEnv.FromBytes
)

// FromFile is a quick way to parse a template from a file. The template doesn't
// allow any includes. If you want to use includes, create a custom environment
// with an appropriate loader.
func FromFile(path string) (*exec.Template, error) {
	return fileSystemLoader.Load(path, DefaultEnv.EvalConfig)
}

// Must panics, if a Template couldn't successfully parsed. This is how you
// would use it:
//
//	var tpl = gonja.Must(gonja.FromFile("templates/base.html"))
func Must(tpl *exec.Template, err error) *exec.Template {
	if err != nil {
		panic(err)
	}
	return tpl
}

// convenient interface to create a new UndefinedFunc
var (
	Undefined              exec.UndefinedFunc = exec.NewUndefinedValue
	StrictUndefined        exec.UndefinedFunc = exec.NewStrictUndefinedValue
	ChainedUndefined       exec.UndefinedFunc = exec.NewChainedUndefinedValue
	ChainedStrictUndefined exec.UndefinedFunc = exec.NewChainedStrictUndefinedValue
)

// convenient interface to create a new Loaders
var (
	NullLoader                      = loaders.NewNullLoader
	FileSystemLoader                = loaders.NewFileSystemLoader
	MustFileSystemLoader            = loaders.MustNewFileSystemLoader
	FileSystemLoaderWithOptions     = loaders.NewFileSystemLoaderWithOptions
	MustFileSystemLoaderWithOptions = loaders.MustNewFileSystemLoaderWithOptions
	CachedLoader                    = loaders.NewCachedLoader
)
