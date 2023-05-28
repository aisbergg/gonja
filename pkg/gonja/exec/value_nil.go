package exec

import (
	"reflect"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

var _ Value = (*NilValue)(nil)

// NilValue represents a value that is nil.
type NilValue struct {
	BaseValue
}

// NewNilValue creates a new [NilValue].
func NewNilValue() *NilValue {
	return &NilValue{}
}

func (*NilValue) IsString() bool {
	return false
}

func (*NilValue) IsBool() bool {
	return false
}

func (*NilValue) IsFloat() bool {
	return false
}

func (*NilValue) IsInteger() bool {
	return false
}

func (*NilValue) IsNumber() bool {
	return false
}

func (*NilValue) IsCallable() bool {
	return false
}

func (*NilValue) IsList() bool {
	return false
}

func (*NilValue) IsDict() bool {
	return false
}

func (*NilValue) IsIterable() bool {
	return false
}

func (*NilValue) IsNil() bool {
	return true
}

func (*NilValue) IsSafe() bool {
	return false
}

func (*NilValue) IsSliceable() bool {
	return false
}

func (*NilValue) String() string {
	return "None"
}

func (*NilValue) Escaped() string {
	return "None"
}

func (*NilValue) Integer() int {
	return 0
}

func (*NilValue) Float() float64 {
	return 0.0
}

func (*NilValue) Bool() bool {
	return false
}

func (*NilValue) Len() int {
	errors.ThrowTemplateRuntimeError("cannot get length of nil value")
	return 0
}

func (*NilValue) Slice(i, j int) Value {
	errors.ThrowTemplateRuntimeError("cannot slice nil value")
	return nil
}

func (*NilValue) Index(i int) Value {
	errors.ThrowTemplateRuntimeError("cannot index nil value")
	return nil
}

func (*NilValue) Contains(other Value) bool {
	errors.ThrowTemplateRuntimeError("cannot check if value contains another nil value")
	return false
}

func (*NilValue) Iterate(fn func(idx, count int, key, value Value) bool, empty func()) {
	errors.ThrowTemplateRuntimeError("cannot iterate over nil value")
}

func (*NilValue) IterateOrder(fn func(idx, count int, key, value Value) bool, empty func(), reverse, sorted, caseSensitive bool) {
	errors.ThrowTemplateRuntimeError("cannot iterate over nil value")
}

func (*NilValue) Interface() any {
	errors.ThrowTemplateRuntimeError("cannot convert nil value to interface")
	return nil
}

func (*NilValue) ReflectValue() reflect.Value {
	errors.ThrowTemplateRuntimeError("cannot get reflect nil value")
	return reflect.Value{}
}

func (*NilValue) EqualValueTo(other Value) bool {
	errors.ThrowTemplateRuntimeError("cannot compare values")
	return false
}

func (*NilValue) Keys() ValuesList {
	errors.ThrowTemplateRuntimeError("cannot get keys of nil value")
	return nil
}

func (*NilValue) Items() []*Pair {
	errors.ThrowTemplateRuntimeError("cannot get items of nil value")
	return nil
}

func (*NilValue) GetItem(key any) Value {
	errors.ThrowTemplateRuntimeError("cannot set nil value")
	return nil
}

func (*NilValue) SetItem(key string, value interface{}) {
	errors.ThrowTemplateRuntimeError("cannot set nil value")
}
