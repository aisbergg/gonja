package errors

import (
	"fmt"
	"strings"
)

// -----------------------------------------------------------------------------
// TemplateLoadError
// -----------------------------------------------------------------------------

// TemplateLoadError is thrown when a template cannot be loaded.
type TemplateLoadError interface {
	TemplateError
	TemplateLoadError()
	Name() string
}

var _ TemplateLoadError = (*templateLoadError)(nil)

type templateLoadError struct {
	msg  string
	name string
}

// TemplateError is a marker interface for template errors.
func (e *templateLoadError) TemplateError() {}

// TemplateLoadError is a marker interface for template syntax errors.
func (e *templateLoadError) TemplateLoadError() {}

func (e *templateLoadError) Error() string {
	return e.msg
}

func (e *templateLoadError) Name() string {
	return e.name
}

// NewTemplateLoadError creates a new TemplateLoadError.
func NewTemplateLoadError(name, format string, args ...any) TemplateLoadError {
	return &templateLoadError{
		msg:  fmt.Sprintf(format, args...),
		name: name,
	}
}

// -----------------------------------------------------------------------------
// TemplateNotFoundError
// -----------------------------------------------------------------------------

// TemplateNotFoundError is thrown when a filter is called with invalid arguments.
type TemplateNotFoundError interface {
	TemplateError
	TemplateNotFoundError()
	Name() string
}

var _ TemplateNotFoundError = (*templateNotFoundError)(nil)

type templateNotFoundError struct {
	name string
}

// TemplateError is a marker interface for template errors.
func (e *templateNotFoundError) TemplateError() {}

// TemplateNotFoundError is a marker interface for template syntax errors.
func (e *templateNotFoundError) TemplateNotFoundError() {}

func (e *templateNotFoundError) Error() string {
	return fmt.Sprintf("template '%s' not found", e.name)
}

func (e *templateNotFoundError) Name() string {
	return e.name
}

// NewTemplateNotFoundError creates a new TemplateNotFoundError.
func NewTemplateNotFoundError(name string) TemplateNotFoundError {
	return &templateNotFoundError{
		name: name,
	}
}

// -----------------------------------------------------------------------------
// TemplatesNotFoundError
// -----------------------------------------------------------------------------

// TemplatesNotFoundError is thrown when a filter is called with invalid arguments.
type TemplatesNotFoundError interface {
	TemplateError
	TemplatesNotFoundError()
	Names() []string
}

var _ TemplatesNotFoundError = (*templatesNotFoundError)(nil)

type templatesNotFoundError struct {
	names []string
}

// TemplateError is a marker interface for template errors.
func (e *templatesNotFoundError) TemplateError() {}

// TemplatesNotFoundError is a marker interface for template syntax errors.
func (e *templatesNotFoundError) TemplatesNotFoundError() {}

func (e *templatesNotFoundError) Error() string {
	return fmt.Sprintf("non of the given templates could be found: %s", strings.Join(e.names, ", "))
}

// Names returns the names of the templates that were not found.
func (e *templatesNotFoundError) Names() []string {
	return e.names
}
