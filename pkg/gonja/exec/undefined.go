package exec

import (
	"fmt"
	"reflect"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// Undefined is an interface that represents an Undefined value.
type Undefined interface {
	// Undefind is a marker method to identify undefined values.
	Undefined()

	// String returns the string representation of the value.
	String() string

	// GetItem returns the value for the given key.
	GetItem(name any) any
}

// UndefinedFunc is a function that creates a new Undefined value.
type UndefinedFunc func(name string, hintFormat string, args ...any) Undefined

// undefinedType represents the reflect.Type of Undefined.
var undefinedType = reflect.TypeOf((*Undefined)(nil)).Elem()

// -----------------------------------------------------------------------------
// UndefinedValue
// -----------------------------------------------------------------------------

// UndefinedValue represents an undefined value that renders to an empty string.
type UndefinedValue struct {
	name string
	hint string
}

// NewUndefinedValue creates a new UndefinedValue.
func NewUndefinedValue(name, format string, args ...any) Undefined {
	hint := ""
	if format != "" {
		hint = fmt.Sprintf(format, args...)
	}
	return UndefinedValue{
		name: name,
		hint: hint,
	}
}

// Undefined is a marker method to identify undefined values.
func (UndefinedValue) Undefined() {}

func (UndefinedValue) String() string {
	return ""
}

// GetItem returns the value for the given key.
func (u UndefinedValue) GetItem(key any) any {
	errors.ThrowUndefinedError(fmt.Sprintf("%s.%s", u.name, key), u.hint)
	return nil
}

// -----------------------------------------------------------------------------
// StrictUndefinedValue
// -----------------------------------------------------------------------------

// StrictUndefinedValue represents an undefined value that throws an error when
// accessed.
type StrictUndefinedValue struct {
	UndefinedValue
}

// NewStrictUndefinedValue creates a new StrictUndefinedValue.
func NewStrictUndefinedValue(varName string, format string, args ...any) Undefined {
	hint := ""
	if format != "" {
		hint = fmt.Sprintf(format, args...)
	}
	return StrictUndefinedValue{
		UndefinedValue: UndefinedValue{
			name: varName,
			hint: hint,
		},
	}
}

func (u StrictUndefinedValue) String() string {
	errors.ThrowUndefinedError(u.name, u.hint)
	return ""
}
