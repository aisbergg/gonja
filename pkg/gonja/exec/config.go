package exec

import (
	"github.com/pkg/errors"

	"github.com/aisbergg/gonja/pkg/gonja/ext"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// EvalConfig is the configuration for the execution of a template.
type EvalConfig struct {
	*parse.Config

	// TrimBlocks will remove the first newline after a block (block, not
	// variable tag!), if set to true. Defaults to false.
	TrimBlocks bool
	// LstripBlocks will strip leading spaces and tabs from the start of a line
	// to a block, if set to true. Defaults to false.
	LstripBlocks bool
	// NewlineSequence defines the sequence that starts a newline. Must be one
	// of '\r', '\n' or '\r\n'. The default is '\n' which is a useful default
	// for Linux and OS X systems as well as web applications.
	NewlineSequence string
	// KeepTrailingNewline will preserve the trailing newline when rendering
	// templates, if set to true. The default is false, which causes a single
	// newline, if present, to be stripped from the end of the template.
	KeepTrailingNewline bool
	// Autoescape will escape XML/HTML automatically, if set to true. Defaults
	// to false.
	Autoescape bool

	// Resolver allows to customize the way variables are resolved.
	// Resolver Resolver

	// Undefined is the type of undefined values that the resolver returns when
	// a value is not found.
	Undefined UndefinedFunc
	// ExtensionConfig stores configuration for extensions.
	ExtensionConfig map[string]ext.Inheritable

	Filters    *FilterSet
	Globals    map[string]any
	Statements *StatementSet
	Tests      *TestSet
	Loader     TemplateLoader
}

// NewEvalConfig creates a new evaluator configuration.
func NewEvalConfig() *EvalConfig {
	return &EvalConfig{
		Config: parse.NewConfig(),

		TrimBlocks:          false,
		LstripBlocks:        false,
		NewlineSequence:     "\n",
		KeepTrailingNewline: false,
		Autoescape:          false,
		Undefined:           NewUndefinedValue,
		ExtensionConfig:     map[string]ext.Inheritable{},

		Globals:    map[string]any{},
		Filters:    &FilterSet{},
		Statements: &StatementSet{},
		Tests:      &TestSet{},
	}
}

// Inherit copies the configuration and returns a new configuration.
func (cfg EvalConfig) Inherit() *EvalConfig {
	extCfg := map[string]ext.Inheritable{}
	for key, cfg := range cfg.ExtensionConfig {
		extCfg[key] = cfg.Inherit()
	}
	return &EvalConfig{
		Config: cfg.Config.Inherit(),

		TrimBlocks:          cfg.TrimBlocks,
		LstripBlocks:        cfg.LstripBlocks,
		NewlineSequence:     cfg.NewlineSequence,
		KeepTrailingNewline: cfg.KeepTrailingNewline,
		Autoescape:          cfg.Autoescape,
		ExtensionConfig:     extCfg,

		// rendererCfg: cfg.rendererCfg.Inherit(),
		Globals:    cfg.Globals,
		Filters:    cfg.Filters,
		Statements: cfg.Statements,
		Tests:      cfg.Tests,
		Loader:     cfg.Loader,
	}
}

// GetTemplate returns the template for the given filename.
func (cfg *EvalConfig) GetTemplate(filename string) (*parse.TemplateNode, error) {
	tpl, err := cfg.Loader.GetTemplate(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse template '%s'", filename)
	}
	return tpl.Root, nil
}
