package gonja

import (
	"reflect"

	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/ext"
	"github.com/aisbergg/gonja/pkg/gonja/loaders"
)

// Option is a function that can be used to configure the gonja template engine.
type Option func(*Environment)

// -----------------------------------------------------------------------------
//
// Loader Options
//
// -----------------------------------------------------------------------------

// OptLoader sets the loader that will be used to load templates by name. To set
// a cached filesystem loader for example, use the following code:
//
//	env := gonja.NewEnvironment(
//	    gonja.OptLoader(gonja.CachedLoader(gonja.MustFilesystemLoader("path/to/templates"))),
//	)
func OptLoader(loader loaders.Loader) Option {
	return func(cfg *Environment) {
		cfg.loader = loader
	}
}

// -----------------------------------------------------------------------------
//
// Parser Options
//
// -----------------------------------------------------------------------------

// OptBlockStartString marks the beginning of a block. Defaults to '{%'
func OptBlockStartString(s string) Option {
	return func(cfg *Environment) {
		cfg.BlockStartString = s
	}
}

// OptBlockEndString marks the end of a block. Defaults to '%}'.
func OptBlockEndString(s string) Option {
	return func(cfg *Environment) {
		cfg.BlockEndString = s
	}
}

// OptVariableStartString marks the the beginning of a print statement. Defaults
// to '{{'.
func OptVariableStartString(s string) Option {
	return func(cfg *Environment) {
		cfg.VariableStartString = s
	}
}

// OptVariableEndString marks the end of a print statement. Defaults to '}}'.
func OptVariableEndString(s string) Option {
	return func(cfg *Environment) {
		cfg.VariableEndString = s
	}
}

// OptCommentStartString marks the beginning of a comment. Defaults to '{#'.
func OptCommentStartString(s string) Option {
	return func(cfg *Environment) {
		cfg.CommentStartString = s
	}
}

// OptCommentEndString marks the end of a comment. Defaults to '#}'.
func OptCommentEndString(s string) Option {
	return func(cfg *Environment) {
		cfg.CommentEndString = s
	}
}

// OptLineStatementPrefix will be used as prefix for line based statements, if
// given and a string.
func OptLineStatementPrefix(s string) Option {
	return func(cfg *Environment) {
		cfg.LineStatementPrefix = s
	}
}

// OptLineCommentPrefix will be used as prefix for line based comments, if given
// and a string.
func OptLineCommentPrefix(s string) Option {
	return func(cfg *Environment) {
		cfg.LineCommentPrefix = s
	}
}

// -----------------------------------------------------------------------------
//
// Exec Options
//
// -----------------------------------------------------------------------------

// OptNewlineSequence defines the sequence that starts a newline. Must be one of
// '\r', '\n' or '\r\n'. The default is '\n' which is a useful default for Linux
// and OS X systems as well as web applications.
func OptNewlineSequence(s string) Option {
	return func(cfg *Environment) {
		cfg.NewlineSequence = s
	}
}

// OptTrimBlocks enables the removal of the first newline after a block (block,
// not variable tag!). Disabled by default.
func OptTrimBlocks() Option {
	return func(cfg *Environment) {
		cfg.TrimBlocks = true
	}
}

// OptNoTrimBlocks disables the TrimBlocks feature. It is disabled by default.
func OptNoTrimBlocks() Option {
	return func(cfg *Environment) {
		cfg.TrimBlocks = false
	}
}

// OptLstripBlocks enables the strip of leading spaces and tabs from the start
// of a line to a block. Disabled by default.
func OptLstripBlocks() Option {
	return func(cfg *Environment) {
		cfg.LstripBlocks = true
	}
}

// OptNoLstripBlocks disables the LstripBlocks feature.
func OptNoLstripBlocks() Option {
	return func(cfg *Environment) {
		cfg.LstripBlocks = false
	}
}

// OptKeepTrailingNewline enables the preservation of trailing newline when
// rendering templates. It is disabled by default, which causes a single
// newline, if present, to be stripped from the end of the template.
func OptKeepTrailingNewline() Option {
	return func(cfg *Environment) {
		cfg.KeepTrailingNewline = true
	}
}

// OptNoKeepTrailingNewline disables the KeepTrailingNewline feature.
func OptNoKeepTrailingNewline() Option {
	return func(cfg *Environment) {
		cfg.KeepTrailingNewline = false
	}
}

// OptAutoescape enables the XML/HTML autoescaping feature. It is disabled by
// default.
func OptAutoescape() Option {
	return func(cfg *Environment) {
		cfg.Autoescape = true
	}
}

// OptNoAutoescape disables the XML/HTML autoescaping feature. It is disabled by
// default.
func OptNoAutoescape() Option {
	return func(cfg *Environment) {
		cfg.Autoescape = false
	}
}

// OptUndefined sets the behavior for undefined variables.
func OptUndefined(undefined exec.UndefinedFunc) Option {
	return func(cfg *Environment) {
		cfg.Undefined = undefined
	}
}

// OptSetExtensionConfig sets a configuration for an extension. If the given
// config is nil, the named configuration will be removed from the environment.
func OptSetExtensionConfig(name string, config ext.Inheritable) Option {
	return func(cfg *Environment) {
		if config == nil {
			delete(cfg.ExtensionConfig, name)
			return
		}
		cfg.ExtensionConfig[name] = config
	}
}

// OptSetGlobal sets a global variable in the environment. If the value is nil,
// the variable will be removed from the environment.
func OptSetGlobal(name string, value any) Option {
	return func(cfg *Environment) {
		if value == nil {
			delete(cfg.Globals, name)
			return
		}
		cfg.Globals[name] = value
	}
}

// OptRegisterCustomType registers a custom value type that is not supported by
// default, e.g. an ordered map. If the value is nil, the type will be
// unregistered from the environment.
func OptRegisterCustomType(typ reflect.Type, getter exec.ValueFunc) Option {
	return func(cfg *Environment) {
		if getter == nil {
			delete(cfg.CustomTypes, typ)
			return
		}
		cfg.CustomTypes[typ] = getter
	}
}
