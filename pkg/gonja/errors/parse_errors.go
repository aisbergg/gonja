package errors

import (
	"fmt"
)

// -----------------------------------------------------------------------------
// TemplateSyntaxError
// -----------------------------------------------------------------------------

// TemplateSyntaxError is thrown when a filter is called with invalid arguments.
type TemplateSyntaxError interface {
	TemplateError
	TemplateSyntaxError()
	Pos() int
}

var _ TemplateSyntaxError = (*templateSyntaxError)(nil)

type templateSyntaxError struct {
	msg   string
	token *Token
}

// TemplateError is a marker interface for template errors.
func (e *templateSyntaxError) TemplateError() {}

// TemplateSyntaxError is a marker interface for template syntax errors.
func (e *templateSyntaxError) TemplateSyntaxError() {}

func (e *templateSyntaxError) Error() string {
	return fmt.Sprintf("%s (pos: %d, line: %d, column: %d, near: '%s')", e.msg, e.token.Pos, e.token.Line, e.token.Col, e.token.Val)
}

// Pos returns the position of the error.
func (e *templateSyntaxError) Pos() int {
	return e.token.Pos
}

// Enrich enriches the error with a token.
func (e *templateSyntaxError) Enrich(tk *Token) {
	if e.token == nil {
		e.token = tk
	}
}

// ThrowSyntaxError throws a syntax error.
func ThrowSyntaxError(token *Token, format string, args ...any) {
	panic(&templateSyntaxError{
		msg:   fmt.Sprintf(format, args...),
		token: token,
	})
}

// -----------------------------------------------------------------------------
// TemplateAssertionError
// -----------------------------------------------------------------------------

// TemplateAssertionError is thrown when a filter is called with invalid arguments.
type TemplateAssertionError interface {
	TemplateSyntaxError
	TemplateAssertionError()
}

var _ TemplateAssertionError = (*templateAssertionError)(nil)

type templateAssertionError struct {
	templateSyntaxError
}

// TemplateSyntaxError is a marker interface for template syntax errors.
func (e *templateSyntaxError) TemplateAssertionError() {}

// ThrowTemplateAssertionError throws a filter argument error.
func ThrowTemplateAssertionError(fname, format string, args ...any) {
	panic(&templateAssertionError{
		templateSyntaxError: templateSyntaxError{
			msg: fmt.Sprintf(format, args...),
		},
	})
}
