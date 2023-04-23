package exec

import (
	"reflect"

	"github.com/aisbergg/gonja/pkg/gonja/ext"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// EvalConfig is the configuration for the execution of a template.
type EvalConfig struct {
	*parse.Config

	Globals        map[string]any
	Filters        *FilterSet
	Statements     *StatementSet
	Tests          *TestSet
	TemplateLoadFn TemplateLoadFn

	// ExtensionConfig stores configuration for extensions.
	ExtensionConfig map[string]ext.Inheritable

	// CustomTypes allows to add custom value representations for types that are
	// not supported by default. For example, if you want to resolve value from
	// a custom ordered map type, you can add a custom value representation for
	// that and implement the [GetItem] method.
	CustomTypes map[reflect.Type]ValueFunc

	// Undefined is the type of undefined values that the resolver returns when
	// a value is not found.
	Undefined UndefinedFunc

	// NewlineSequence defines the sequence that starts a newline. Must be one
	// of '\r', '\n' or '\r\n'. The default is '\n' which is a useful default
	// for Linux and OS X systems as well as web applications.
	NewlineSequence string

	// TrimBlocks will remove the first newline after a block (block, not
	// variable tag!), if set to true. Defaults to false.
	TrimBlocks bool

	// LstripBlocks will strip leading spaces and tabs from the start of a line
	// to a block, if set to true. Defaults to false.
	LstripBlocks bool

	// KeepTrailingNewline will preserve the trailing newline when rendering
	// templates, if set to true. The default is false, which causes a single
	// newline, if present, to be stripped from the end of the template.
	KeepTrailingNewline bool

	// Autoescape will escape XML/HTML automatically, if set to true. Defaults
	// to false.
	Autoescape bool
}

// NewEvalConfig creates a new evaluator configuration.
func NewEvalConfig() *EvalConfig {
	return &EvalConfig{
		Config: parse.NewConfig(),

		Globals:    map[string]any{},
		Filters:    &FilterSet{},
		Statements: &StatementSet{},
		Tests:      &TestSet{},

		ExtensionConfig:     map[string]ext.Inheritable{},
		CustomTypes:         map[reflect.Type]ValueFunc{},
		Undefined:           NewUndefinedValue,
		NewlineSequence:     "\n",
		TrimBlocks:          false,
		LstripBlocks:        false,
		KeepTrailingNewline: false,
		Autoescape:          false,
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

		Globals:        cfg.Globals,
		Filters:        cfg.Filters,
		Statements:     cfg.Statements,
		Tests:          cfg.Tests,
		TemplateLoadFn: cfg.TemplateLoadFn,

		ExtensionConfig:     extCfg,
		CustomTypes:         cfg.CustomTypes,
		Undefined:           cfg.Undefined,
		NewlineSequence:     cfg.NewlineSequence,
		TrimBlocks:          cfg.TrimBlocks,
		LstripBlocks:        cfg.LstripBlocks,
		KeepTrailingNewline: cfg.KeepTrailingNewline,
		Autoescape:          cfg.Autoescape,
	}
}

// templateParseFn returns the parsed template for the given filename.
func (cfg *EvalConfig) templateParseFn(filename string) (*parse.TemplateNode, error) {
	tpl, err := cfg.TemplateLoadFn(filename)
	if err != nil {
		return nil, err
	}
	return tpl.Root, nil
}
