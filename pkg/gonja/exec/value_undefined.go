package exec

import (
	"fmt"
	"reflect"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// Undefined is an interface that represents an Undefined value.
type Undefined interface {
	Value

	// Undefind is a marker method to identify undefined values.
	Undefined()

	// VariableName returns the name of the variable that is undefined.
	VariableName() string

	// Hint returns a hint for the undefined variable.
	Hint() string

	// Get returns the value for the given key.
	// Get(name any) any
}

// UndefinedFunc is a function that creates a new Undefined value.
type UndefinedFunc func(name string, hintFormat string, args ...any) Undefined

// undefinedType represents the reflect.Type of Undefined.
var undefinedType = reflect.TypeOf((*Undefined)(nil)).Elem()

// IsDefined returns true if the given value is not undefined.
func IsDefined(val Value) bool {
	_, ok := val.(Undefined)
	return !ok
}

// -----------------------------------------------------------------------------
//
// UndefinedValue
//
// -----------------------------------------------------------------------------

var _ Undefined = (*UndefinedValue)(nil)

// UndefinedValue represents an undefined value that renders to an empty string.
// Most other access methods will throw an error.
type UndefinedValue struct {
	BaseValue

	name string
	hint string
}

// NewUndefinedValue creates a new UndefinedValue.
func NewUndefinedValue(name, format string, args ...any) Undefined {
	hint := ""
	if format != "" {
		hint = fmt.Sprintf(format, args...)
	}
	return &UndefinedValue{
		name: name,
		hint: hint,
	}
}

// Undefined is a marker method to identify undefined values.
func (*UndefinedValue) Undefined() {}

// VariableName returns the name of the variable that is undefined.
func (u UndefinedValue) VariableName() string {
	return u.name
}

// Hint returns a hint for the undefined variable.
func (u UndefinedValue) Hint() string {
	return u.hint
}

func (*UndefinedValue) String() string {
	return ""
}
func (*UndefinedValue) Escaped() string {
	return ""
}
func (u UndefinedValue) Integer() int {
	errors.ThrowUndefinedError(u.name, u.hint)
	return 0
}
func (u UndefinedValue) Float() float64 {
	errors.ThrowUndefinedError(u.name, u.hint)
	return 0.0
}
func (u UndefinedValue) Bool() bool {
	errors.ThrowUndefinedError(u.name, u.hint)
	return false
}
func (*UndefinedValue) Len() int {
	return 0
}
func (u UndefinedValue) Slice(i, j int) Value {
	return u.valueFactory.NewValue("", false)
}
func (u UndefinedValue) Index(i int) Value {
	return u.valueFactory.NewValue("", false)
}
func (u *UndefinedValue) EqualValueTo(other Value) bool {
	return false
}
func (u *UndefinedValue) Keys() ValuesList {
	errors.ThrowUndefinedError(u.name, u.hint)
	return ValuesList{}
}
func (u *UndefinedValue) Values() ValuesList {
	errors.ThrowUndefinedError(u.name, u.hint)
	return ValuesList{}
}
func (u *UndefinedValue) Items() []*Pair {
	errors.ThrowUndefinedError(u.name, u.hint)
	return []*Pair{}
}
func (u UndefinedValue) GetItem(key any) Value {
	errors.ThrowUndefinedError(u.name, u.hint)
	return nil
}
func (u UndefinedValue) Set(key string, value interface{}) {
	errors.ThrowUndefinedError(u.name, u.hint)
}
func (*UndefinedValue) Iterate(fn func(idx, count int, key, value Value) bool, empty func()) {}
func (*UndefinedValue) IterateOrder(fn func(idx, count int, key, value Value) bool, empty func(), reverse bool, sorted bool, caseSensitive bool) {
}

// -----------------------------------------------------------------------------
//
// StrictUndefinedValue
//
// -----------------------------------------------------------------------------

var _ Undefined = (*StrictUndefinedValue)(nil)

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
	return &StrictUndefinedValue{
		UndefinedValue: UndefinedValue{
			name: varName,
			hint: hint,
		},
	}
}

func (u *StrictUndefinedValue) String() string {
	errors.ThrowUndefinedError(u.name, u.hint)
	return ""
}
func (u *StrictUndefinedValue) IsTrue() bool {
	errors.ThrowUndefinedError(u.name, u.hint)
	return false
}
func (u *StrictUndefinedValue) Len() int {
	errors.ThrowUndefinedError(u.name, u.hint)
	return 0
}
func (u *StrictUndefinedValue) Contains(other Value) bool {
	errors.ThrowUndefinedError(u.name, u.hint)
	return false
}
func (u *StrictUndefinedValue) Iterate(fn func(idx, count int, key, value Value) bool, empty func()) {
	errors.ThrowUndefinedError(u.name, u.hint)
}
func (u *StrictUndefinedValue) IterateOrder(fn func(idx, count int, key, value Value) bool, empty func(), reverse bool, sorted bool, caseSensitive bool) {
	errors.ThrowUndefinedError(u.name, u.hint)
}
func (u *StrictUndefinedValue) EqualValueTo(other Value) bool {
	errors.ThrowUndefinedError(u.name, u.hint)
	return false
}

// -----------------------------------------------------------------------------
//
// ChainedUndefinedValue
//
// -----------------------------------------------------------------------------

var _ Undefined = (*ChainedUndefinedValue)(nil)

// ChainedUndefinedValue represents an undefined value that renders to an empty
// string. Most other access methods will throw an error. It is different from
// UndefinedValue in that it allows for chaining for `Get` calls.
type ChainedUndefinedValue struct {
	UndefinedValue
}

// NewChainedUndefinedValue creates a new ChainedUndefinedValue.
func NewChainedUndefinedValue(varName string, format string, args ...any) Undefined {
	hint := ""
	if format != "" {
		hint = fmt.Sprintf(format, args...)
	}
	return &ChainedUndefinedValue{
		UndefinedValue: UndefinedValue{
			name: varName,
			hint: hint,
		},
	}
}

// Get returns the value for the given key.
func (u *ChainedUndefinedValue) GetItem(key any) Value {
	return NewChainedUndefinedValue(fmt.Sprintf("%s.%s", u.name, key), u.hint)
}

// -----------------------------------------------------------------------------
//
// ChainedStrictUndefinedValue
//
// -----------------------------------------------------------------------------

var _ Undefined = (*ChainedStrictUndefinedValue)(nil)

// ChainedStrictUndefinedValue is like ChainedUndefinedValue but throws an error
// on any other type of access.
type ChainedStrictUndefinedValue struct {
	StrictUndefinedValue
}

// NewChainedStrictUndefinedValue creates a new ChainedStrictUndefinedValue.
func NewChainedStrictUndefinedValue(varName string, format string, args ...any) Undefined {
	hint := ""
	if format != "" {
		hint = fmt.Sprintf(format, args...)
	}
	return &ChainedStrictUndefinedValue{
		StrictUndefinedValue: StrictUndefinedValue{
			UndefinedValue: UndefinedValue{
				name: varName,
				hint: hint,
			},
		},
	}
}

// Get returns the value for the given key.
func (u *ChainedStrictUndefinedValue) GetItem(key any) Value {
	return NewChainedStrictUndefinedValue(fmt.Sprintf("%s.%s", u.name, key), u.hint)
}
