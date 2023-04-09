package gonja

import (
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/ext"
)

// Option is a function that can be used to configure the gonja template engine.
type Option func(*Environment)

// BlockStartString marks the beginning of a block. Defaults to '{%'
func BlockStartString(s string) Option {
	return func(cfg *Environment) {
		cfg.BlockStartString = s
	}
}

// BlockEndString marks the end of a block. Defaults to '%}'.
func BlockEndString(s string) Option {
	return func(cfg *Environment) {
		cfg.BlockEndString = s
	}
}

// VariableStartString marks the the beginning of a print statement. Defaults to '{{'.
func VariableStartString(s string) Option {
	return func(cfg *Environment) {
		cfg.VariableStartString = s
	}
}

// VariableEndString marks the end of a print statement. Defaults to '}}'.
func VariableEndString(s string) Option {
	return func(cfg *Environment) {
		cfg.VariableEndString = s
	}
}

// CommentStartString marks the beginning of a comment. Defaults to '{#'.
func CommentStartString(s string) Option {
	return func(cfg *Environment) {
		cfg.CommentStartString = s
	}
}

// CommentEndString marks the end of a comment. Defaults to '#}'.
func CommentEndString(s string) Option {
	return func(cfg *Environment) {
		cfg.CommentEndString = s
	}
}

// LineStatementPrefix will be used as prefix for line based statements, if
// given and a string.
func LineStatementPrefix(s string) Option {
	return func(cfg *Environment) {
		cfg.LineStatementPrefix = s
	}
}

// LineCommentPrefix will be used as prefix for line based comments, if given
// and a string.
func LineCommentPrefix(s string) Option {
	return func(cfg *Environment) {
		cfg.LineCommentPrefix = s
	}
}

// NewlineSequence defines the sequence that starts a newline. Must be one of
// '\r', '\n' or '\r\n'. The default is '\n' which is a useful default for Linux
// and OS X systems as well as web applications.
func NewlineSequence(s string) Option {
	return func(cfg *Environment) {
		cfg.NewlineSequence = s
	}
}

// TrimBlocks enables the removal of the first newline after a block (block, not
// variable tag!). Disabled by default.
func TrimBlocks() Option {
	return func(cfg *Environment) {
		cfg.TrimBlocks = true
	}
}

// NoTrimBlocks disables the TrimBlocks feature.
func NoTrimBlocks() Option {
	return func(cfg *Environment) {
		cfg.TrimBlocks = false
	}
}

// LstripBlocks enables the strip of leading spaces and tabs from the start of a
// line to a block. Disabled by default.
func LstripBlocks() Option {
	return func(cfg *Environment) {
		cfg.LstripBlocks = true
	}
}

// NoLstripBlocks disables the LstripBlocks feature.
func NoLstripBlocks() Option {
	return func(cfg *Environment) {
		cfg.LstripBlocks = false
	}
}

// KeepTrailingNewline enables the preservation of trailing newline when
// rendering templates. It is disabled by default, which causes a single
// newline, if present, to be stripped from the end of the template.
func KeepTrailingNewline() Option {
	return func(cfg *Environment) {
		cfg.KeepTrailingNewline = true
	}
}

// NoKeepTrailingNewline disables the KeepTrailingNewline feature.
func NoKeepTrailingNewline() Option {
	return func(cfg *Environment) {
		cfg.KeepTrailingNewline = false
	}
}

// Autoescape enables the XML/HTML autoescaping feature. It is disabled by
// default.
func Autoescape() Option {
	return func(cfg *Environment) {
		cfg.Autoescape = true
	}
}

// NoAutoescape disables the Autoescape feature.
func NoAutoescape() Option {
	return func(cfg *Environment) {
		cfg.Autoescape = false
	}
}

// UndefinedOpt sets the behavior for undefined variables.
func UndefinedOpt(undefined exec.UndefinedFunc) Option {
	return func(cfg *Environment) {
		cfg.Undefined = undefined
	}
}

// SetExtensionConfig sets a configuration for an extension.
func SetExtensionConfig(name string, config ext.Inheritable) Option {
	return func(cfg *Environment) {
		cfg.ExtensionConfig[name] = config
	}
}

// SetGlobal sets a global variable in the environment.
func SetGlobal(name string, value any) Option {
	return func(cfg *Environment) {
		cfg.Globals[name] = value
	}
}
