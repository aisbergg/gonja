package errors

import (
	"fmt"
)

// -----------------------------------------------------------------------------
// TemplateRuntimeError
// -----------------------------------------------------------------------------

// TemplateRuntimeError is a generic runtime error. It is used as a base for
// other more specific runtime errors.
type TemplateRuntimeError interface {
	TemplateError
	TemplateRuntimeError()
	Enrich(*Token)
	Token() *Token
}

var _ TemplateRuntimeError = (*templateRuntimeError)(nil)

type templateRuntimeError struct {
	msg   string
	token *Token
}

// TemplateError is a marker interface for template errors.
func (e *templateRuntimeError) TemplateError() {}

// TemplateRuntimeError is a marker interface for template runtime errors.
func (e *templateRuntimeError) TemplateRuntimeError() {}

func (e *templateRuntimeError) Error() string {
	return fmt.Sprintf("%s (pos: %d, line: %d, column: %d, near: '%s')", e.msg, e.token.Pos, e.token.Line, e.token.Col, e.token.Val)
}

func (e *templateRuntimeError) Token() *Token {
	return e.token
}

func (e *templateRuntimeError) Enrich(tk *Token) {
	if e.token == nil {
		e.token = tk
	}
}

// NewTemplateRuntimeError creates a new TemplateRuntimeError.
func NewTemplateRuntimeError(format string, args ...any) TemplateRuntimeError {
	return &templateRuntimeError{
		msg: fmt.Sprintf(format, args...),
	}
}

// ThrowTemplateRuntimeError throws a generic template runtime error.
func ThrowTemplateRuntimeError(format string, args ...any) {
	panic(&templateRuntimeError{
		msg: fmt.Sprintf(format, args...),
	})
}

// -----------------------------------------------------------------------------
// FilterArgumentError
// -----------------------------------------------------------------------------

// FilterArgumentError is thrown when a filter is called with invalid arguments.
type FilterArgumentError interface {
	TemplateRuntimeError
	FilterName() string
}

var _ FilterArgumentError = (*filterArgumentError)(nil)

type filterArgumentError struct {
	templateRuntimeError
	filterName string
}

// ThrowFilterArgumentError throws a filter argument error.
func ThrowFilterArgumentError(fname, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	panic(&filterArgumentError{
		templateRuntimeError: templateRuntimeError{
			msg: fmt.Sprintf("%s: %s", fname, msg),
		},
		filterName: fname,
	})
}

// FilterName returns the name of the filter that caused the error.
func (e *filterArgumentError) FilterName() string {
	return e.filterName
}

// -----------------------------------------------------------------------------
// UndefinedError
// -----------------------------------------------------------------------------

// UndefinedError is thrown when an action is performed on an undefined variable.
type UndefinedError interface {
	TemplateRuntimeError
	VariableName() string
}

var _ UndefinedError = (*undefinedError)(nil)

type undefinedError struct {
	templateRuntimeError
	variableName string
}

// ThrowUndefinedError throws an undefined error.
func ThrowUndefinedError(varName string, hint string) {
	msg := ""
	if hint == "" {
		msg = fmt.Sprintf("undefined variable: %s", varName)
	} else {
		msg = fmt.Sprintf("undefined: %s", hint)
	}
	panic(&undefinedError{
		templateRuntimeError: templateRuntimeError{
			msg: msg,
		},
		variableName: varName,
	})
}

// VariableName returns the name of the undefined variable.
func (e *undefinedError) VariableName() string {
	return e.variableName
}
