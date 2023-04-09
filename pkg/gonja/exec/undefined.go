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

	// Get returns the value for the given key.
	Get(name any) any

	// Value interfaces
	IsString() bool
	IsBool() bool
	IsFloat() bool
	IsInteger() bool
	IsNumber() bool
	IsCallable() bool
	IsList() bool
	IsDict() bool
	IsIterable() bool
	IsNil() bool
	String() string
	Integer() int
	Float() float64
	Bool() bool
	IsTrue() bool
	Len() int
	Slice(i, j int) *Value
	Index(i int) *Value
	Contains(other *Value) bool
	CanSlice() bool
	Iterate(fn func(idx, count int, key, value *Value) bool, empty func())
	IterateOrder(fn func(idx, count int, key, value *Value) bool, empty func(), reverse bool, sorted bool, caseSensitive bool)
	EqualValueTo(other *Value) bool
	Keys() ValuesList
	Items() []*Pair
	Set(key string, value interface{})
}

// UndefinedFunc is a function that creates a new Undefined value.
type UndefinedFunc func(name string, hintFormat string, args ...any) Undefined

// undefinedType represents the reflect.Type of Undefined.
var undefinedType = reflect.TypeOf((*Undefined)(nil)).Elem()

// -----------------------------------------------------------------------------
// UndefinedValue
// -----------------------------------------------------------------------------

// UndefinedValue represents an undefined value that renders to an empty string.
// Most other access methods will throw an error.
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

// Get returns the value for the given key.
func (u UndefinedValue) Get(key any) any {
	errors.ThrowUndefinedError(u.name, u.hint)
	return nil
}

func (UndefinedValue) IsString() bool {
	return false
}
func (UndefinedValue) IsBool() bool {
	return false
}
func (UndefinedValue) IsFloat() bool {
	return false
}
func (UndefinedValue) IsInteger() bool {
	return false
}
func (UndefinedValue) IsNumber() bool {
	return false
}
func (UndefinedValue) IsCallable() bool {
	return false
}
func (UndefinedValue) IsList() bool {
	return false
}
func (UndefinedValue) IsDict() bool {
	return false
}
func (UndefinedValue) IsIterable() bool {
	return false
}
func (UndefinedValue) IsNil() bool {
	return false
}
func (UndefinedValue) String() string {
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
func (UndefinedValue) IsTrue() bool {
	return false
}
func (UndefinedValue) Len() int {
	return 0
}
func (u UndefinedValue) Slice(i, j int) *Value {
	errors.ThrowUndefinedError(u.name, u.hint)
	return nil
}
func (u UndefinedValue) Index(i int) *Value {
	errors.ThrowUndefinedError(u.name, u.hint)
	return nil
}
func (UndefinedValue) Contains(other *Value) bool {
	return false
}
func (UndefinedValue) CanSlice() bool {
	return false
}
func (UndefinedValue) Iterate(fn func(idx, count int, key, value *Value) bool, empty func()) {}
func (UndefinedValue) IterateOrder(fn func(idx, count int, key, value *Value) bool, empty func(), reverse bool, sorted bool, caseSensitive bool) {
}
func (UndefinedValue) EqualValueTo(other *Value) bool {
	return false
}
func (UndefinedValue) Keys() ValuesList {
	return ValuesList{}
}
func (UndefinedValue) Items() []*Pair {
	return []*Pair{}
}
func (u UndefinedValue) Set(key string, value interface{}) {
	errors.ThrowUndefinedError(u.name, u.hint)
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
func (u StrictUndefinedValue) IsTrue() bool {
	errors.ThrowUndefinedError(u.name, u.hint)
	return false
}
func (u StrictUndefinedValue) Len() int {
	errors.ThrowUndefinedError(u.name, u.hint)
	return 0
}
func (u StrictUndefinedValue) Contains(other *Value) bool {
	errors.ThrowUndefinedError(u.name, u.hint)
	return false
}
func (u StrictUndefinedValue) Iterate(fn func(idx, count int, key, value *Value) bool, empty func()) {
	errors.ThrowUndefinedError(u.name, u.hint)
}
func (u StrictUndefinedValue) IterateOrder(fn func(idx, count int, key, value *Value) bool, empty func(), reverse bool, sorted bool, caseSensitive bool) {
	errors.ThrowUndefinedError(u.name, u.hint)
}
func (u StrictUndefinedValue) EqualValueTo(other *Value) bool {
	errors.ThrowUndefinedError(u.name, u.hint)
	return false
}
func (u StrictUndefinedValue) Keys() ValuesList {
	errors.ThrowUndefinedError(u.name, u.hint)
	return ValuesList{}
}
func (u StrictUndefinedValue) Items() []*Pair {
	errors.ThrowUndefinedError(u.name, u.hint)
	return nil
}

// -----------------------------------------------------------------------------
// ChainedUndefinedValue
// -----------------------------------------------------------------------------

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
	return ChainedUndefinedValue{
		UndefinedValue: UndefinedValue{
			name: varName,
			hint: hint,
		},
	}
}

// Get returns the value for the given key.
func (u ChainedUndefinedValue) Get(key any) any {
	return NewChainedUndefinedValue(fmt.Sprintf("%s.%s", u.name, key), u.hint)
}

// -----------------------------------------------------------------------------
// ChainedStrictUndefinedValue
// -----------------------------------------------------------------------------

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
	return ChainedStrictUndefinedValue{
		StrictUndefinedValue: StrictUndefinedValue{
			UndefinedValue: UndefinedValue{
				name: varName,
				hint: hint,
			},
		},
	}
}

// Get returns the value for the given key.
func (u ChainedStrictUndefinedValue) Get(key any) any {
	return NewChainedStrictUndefinedValue(fmt.Sprintf("%s.%s", u.name, key), u.hint)
}
