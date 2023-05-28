package exec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

var (
	rtValue      = reflect.TypeOf((*Value)(nil)).Elem()
	rtValuesList = reflect.TypeOf((ValuesList)(nil))
	rtDict       = reflect.TypeOf((*Dict)(nil))
)

func indirectReflectValue(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		return indirectReflectValue(val.Elem())
	}
	return val
}

// -----------------------------------------------------------------------------
//
// Value Interface
//
// -----------------------------------------------------------------------------

// Value is the interface that all value containers must implement. You can use
// [BaseValue] as a base for your own value container implementations.
type Value interface {
	// IsString returns true if the value is a string, false otherwise.
	IsString() bool

	// IsBool returns true if the value is a boolean, false otherwise.
	IsBool() bool

	// IsFloat returns true if the value is a float, false otherwise.
	IsFloat() bool

	// IsInteger returns true if the value is an integer, false otherwise.
	IsInteger() bool

	// IsNumber returns true if the value is a number (integer or float), false
	// otherwise.
	IsNumber() bool

	// IsList returns true if the value is a list or array, false otherwise.
	IsList() bool

	// IsDict returns true if the value is a dictionary or map, false otherwise.
	IsDict() bool

	// IsNil returns true if the value is nil or null, false otherwise.
	IsNil() bool

	// IsSafe returns true if the value is safe for concurrent use, false
	// otherwise.
	IsSafe() bool

	// IsCallable returns true if the value is a callable function or method,
	// false otherwise.
	IsCallable() bool

	// IsIterable returns true if the value is iterable, false otherwise.
	IsIterable() bool

	// IsSliceable returns true if the value is sliceable, false otherwise.
	IsSliceable() bool

	// Interface returns the underlying value as an interface{}.
	Interface() any

	// ReflectValue returns the underlying reflect value.
	ReflectValue() reflect.Value

	// String returns the string representation of the value.
	String() string

	// Escaped returns the escaped string representation of the value.
	Escaped() string

	// Integer returns the integer representation of the value.
	Integer() int

	// Float returns the float representation of the value.
	Float() float64

	// Bool returns the boolean representation of the value.
	Bool() bool

	// Len returns the length of the value, if it's a list, dictionary, or
	// string.
	Len() int

	// Slice returns a slice of the value, if it's a list or string, from index
	// i to j.
	Slice(i, j int) Value

	// Index returns the value at the given index, if it's a list or string.
	Index(i int) Value

	// Contains returns true if the value contains the given value.
	Contains(other Value) bool

	// Keys returns the keys of the underlying map.
	Keys() ValuesList

	// Values returns the values of the underlying map.
	Values() ValuesList

	// Items returns the key-value pairs of the underlying map.
	Items() []*Pair

	// GetItem returns the value associated with the given key.
	GetItem(key any) Value

	// SetItem sets the value associated with the given key to the given value.
	SetItem(key string, value any)

	// Iterate iterates over the value's items, if it's a list or dictionary,
	// and calls the provided function for each item. If the value is empty, the
	// empty function is called instead.
	Iterate(fn func(idx, count int, key, value Value) (cont bool), empty func())

	// IterateOrder iterates over the value's items, if it's a dictionary, in a
	// specified order, and calls the provided function for each item. If the
	// value is empty, the empty function is called instead. If reverse is true,
	// the items are iterated in reverse order. If sorted is true, the items are
	// sorted by key. If caseSensitive is true, the keys are compared
	// case-sensitively.
	IterateOrder(fn func(idx, count int, key, value Value) (cont bool), empty func(), reverse, sorted, caseSensitive bool)

	// EqualValueTo returns true if the value is equal to the other value, false
	// otherwise.
	EqualValueTo(other Value) bool
}

// ValueFunc is a function that creates a new value container.
type ValueFunc func(value any, safe bool, valueFactory *ValueFactory) Value

// -----------------------------------------------------------------------------
//
// BaseValue
//
// -----------------------------------------------------------------------------

var _ Value = (*BaseValue)(nil)

// BaseValue serves as a base for value containers.
type BaseValue struct {
	// valueFactory is used to create new [Value] containers.
	valueFactory *ValueFactory

	// isSafe indicates whether the value needs explicit escaping in the template
	// or not.
	isSafe bool
}

// NewBaseValue creates a new [BaseValue] container.
func NewBaseValue(valueFactory *ValueFactory, isSafe bool) *BaseValue {
	return &BaseValue{
		valueFactory: valueFactory,
		isSafe:       isSafe,
	}
}

func (*BaseValue) IsString() bool {
	return false
}

func (*BaseValue) IsBool() bool {
	return false
}

func (*BaseValue) IsFloat() bool {
	return false
}

func (*BaseValue) IsInteger() bool {
	return false
}

func (*BaseValue) IsNumber() bool {
	return false
}

func (*BaseValue) IsList() bool {
	return false
}

func (*BaseValue) IsDict() bool {
	return false
}

func (*BaseValue) IsNil() bool {
	return false
}

func (v *BaseValue) IsSafe() bool {
	return v.isSafe
}

func (*BaseValue) IsCallable() bool {
	return false
}

func (*BaseValue) IsIterable() bool {
	return false
}

func (*BaseValue) IsSliceable() bool {
	return false
}

func (*BaseValue) Interface() any {
	errors.ThrowTemplateRuntimeError("cannot convert value to interface")
	return nil
}

func (*BaseValue) ReflectValue() reflect.Value {
	errors.ThrowTemplateRuntimeError("cannot get reflect value")
	return reflect.Value{}
}

func (*BaseValue) String() string {
	errors.ThrowTemplateRuntimeError("cannot convert value to string")
	return ""
}

func (*BaseValue) Escaped() string {
	errors.ThrowTemplateRuntimeError("cannot convert value to string")
	return ""
}

func (*BaseValue) Integer() int {
	errors.ThrowTemplateRuntimeError("cannot convert value to integer")
	return 0
}

func (*BaseValue) Float() float64 {
	errors.ThrowTemplateRuntimeError("cannot convert value to float")
	return 0
}

func (*BaseValue) Bool() bool {
	errors.ThrowTemplateRuntimeError("cannot convert value to bool")
	return false
}

func (*BaseValue) Len() int {
	errors.ThrowTemplateRuntimeError("cannot get length of value")
	return 0
}

func (*BaseValue) Slice(i, j int) Value {
	errors.ThrowTemplateRuntimeError("cannot slice value")
	return nil
}

func (*BaseValue) Index(i int) Value {
	errors.ThrowTemplateRuntimeError("cannot index value")
	return nil
}

func (*BaseValue) Contains(other Value) bool {
	errors.ThrowTemplateRuntimeError("cannot check if value contains another value")
	return false
}

func (*BaseValue) Keys() ValuesList {
	errors.ThrowTemplateRuntimeError("cannot get keys of value")
	return nil
}

func (*BaseValue) Values() ValuesList {
	errors.ThrowTemplateRuntimeError("cannot get values of value")
	return nil
}

func (*BaseValue) Items() []*Pair {
	errors.ThrowTemplateRuntimeError("cannot get items of value")
	return nil
}

func (*BaseValue) GetItem(key any) Value {
	errors.ThrowTemplateRuntimeError("cannot set value")
	return nil
}

func (*BaseValue) SetItem(key string, value any) {
	errors.ThrowTemplateRuntimeError("cannot set value")
}

func (*BaseValue) Iterate(fn func(idx, count int, key, value Value) bool, empty func()) {
	errors.ThrowTemplateRuntimeError("cannot iterate over value")
}

func (*BaseValue) IterateOrder(fn func(idx, count int, key, value Value) bool, empty func(), reverse, sorted, caseSensitive bool) {
	errors.ThrowTemplateRuntimeError("cannot iterate over value")
}

func (*BaseValue) EqualValueTo(other Value) bool {
	errors.ThrowTemplateRuntimeError("cannot compare values")
	return false
}

// -----------------------------------------------------------------------------
//
// ValuesList
//
// -----------------------------------------------------------------------------

// ValuesList represents a list of [Value]s.
type ValuesList []Value

// Len is the number of elements in the collection.
func (vl ValuesList) Len() int {
	return len(vl)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (vl ValuesList) Less(i, j int) bool {
	vi := vl[i]
	vj := vl[j]
	switch {
	case vi.IsInteger() && vj.IsInteger():
		return vi.Integer() < vj.Integer()
	case vi.IsFloat() && vj.IsFloat():
		return vi.Float() < vj.Float()
	default:
		return vi.String() < vj.String()
	}
}

// Swap swaps the elements with indexes i and j.
func (vl ValuesList) Swap(i, j int) {
	vl[i], vl[j] = vl[j], vl[i]
}

// String returns a string representation of vl in the form "['value1',
// 'value2']".
func (vl ValuesList) String() string {
	var out strings.Builder
	out.WriteByte('[')
	for idx, key := range vl {
		if idx > 0 {
			out.WriteString(", ")
		}
		if key.IsString() {
			out.WriteString("'")
		}
		out.WriteString(key.String())
		if key.IsString() {
			out.WriteString("'")
		}
	}
	out.WriteByte(']')
	return out.String()
}

// Contains reports whether a value is within vl.
func (vl ValuesList) Contains(value Value) bool {
	for _, val := range vl {
		if value.EqualValueTo(val) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
//
// Dict and Pair Values
//
// -----------------------------------------------------------------------------

// Pair represents a pair of key and value.
type Pair struct {
	Key   Value
	Value Value
}

// String returns a string representation of p in the form "'key': 'value'".
func (p *Pair) String() string {
	var key, value string
	if p.Key.IsString() {
		key = fmt.Sprintf("'%s'", p.Key.String())
	} else {
		key = p.Key.String()
	}
	if p.Value.IsString() {
		value = fmt.Sprintf("'%s'", p.Value.String())
	} else {
		value = p.Value.String()
	}
	return fmt.Sprintf("%s: %s", key, value)
}

// Dict represents a mapping of key-value [Pair]s.
type Dict struct {
	Pairs []*Pair
}

// NewDict creates a new [Dict].
func NewDict() *Dict {
	return &Dict{Pairs: []*Pair{}}
}

// String returns a string representation of d in the form "'key1': 'value',
// "'key2': 'value'".
func (d *Dict) String() string {
	pairs := []string{}
	for _, pair := range d.Pairs {
		pairs = append(pairs, pair.String())
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// Keys returns a [ValueList] of keys contained in the dict.
func (d *Dict) Keys() ValuesList {
	keys := ValuesList{}
	for _, pair := range d.Pairs {
		keys = append(keys, pair.Key)
	}
	return keys
}

// Get returns the [Value] for the given key from d.
func (d *Dict) Get(key Value) (value Value, ok bool) {
	for _, pair := range d.Pairs {
		if pair.Key.EqualValueTo(key) {
			return pair.Value, true
		}
	}
	return nil, false
}
