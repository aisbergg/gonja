package exec

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	u "github.com/aisbergg/gonja/pkg/gonja/utils"
)

var (
	rtUndefined = reflect.TypeOf((*Undefined)(nil)).Elem()
	rtValue     = reflect.TypeOf((*Value)(nil))
)

func indirectReflectValue(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		return indirectReflectValue(val.Elem())
	}
	return val
}

// -----------------------------------------------------------------------------
//
// Value Container
//
// -----------------------------------------------------------------------------

// Value is a container for values of various types.
type Value struct {
	// Val holds the actual value in form of a reflection value.
	Val reflect.Value

	// IndVal holds the indirect (resolved) value . This is used to avoid
	// resolving the pointer values more than once.
	IndVal reflect.Value

	// Safe indicates whether the value needs explicit escaping in the template
	// or not.
	Safe bool
}

// AsValue wraps a given value in a `Value` container. Usually being used within
// functions passed to a template through a Context or within filter functions.
func AsValue(val any) *Value {
	if val == nil {
		return &Value{
			Val:    reflect.Value{},
			IndVal: reflect.Value{},
		}
	}
	refVal := reflect.ValueOf(val)
	indVal := indirectReflectValue(refVal)
	return &Value{
		Val:    refVal,
		IndVal: indVal,
	}
}

// AsSafeValue works like `AsValue`, but does not apply the `escape` filter.
func AsSafeValue(i any) *Value {
	if i == nil {
		return &Value{
			Val:    reflect.Value{},
			IndVal: reflect.Value{},
		}
	}
	refVal := reflect.ValueOf(i)
	indVal := indirectReflectValue(refVal)
	return &Value{
		Val:    refVal,
		IndVal: indVal,
		Safe:   true,
	}
}

// ToValue returns a `Value` container. If the given value is already of type `Value`, then it is returned as is. If it is a `reflect.Value`, then it is wrapped in a `Value` container. All other types are wrapped directly in a `Value` container.
func ToValue(val any) *Value {
	// return an empty value if the given value is nil
	if val == nil {
		return &Value{
			Val:    reflect.Value{},
			IndVal: reflect.Value{},
		}
	}

	// return the value as is, without wrapping it in another `Value` container
	if v, ok := val.(*Value); ok {
		return v
	}

	// all values are turned into a `reflect.Value` first, but if the given
	// value is already a `reflect.Value`, then it is used directly
	refVal := reflect.Value{}
	if rv, ok := val.(reflect.Value); ok {
		refVal = rv
	} else {
		refVal = reflect.ValueOf(val)
	}
	indVal := indirectReflectValue(refVal)
	return &Value{
		Val:    refVal,
		IndVal: indVal,
	}
}

// IsDefined reports whether the underlying value is defined.
func (v *Value) IsDefined() bool {
	return v.IndVal.IsValid() && !v.IndVal.Type().Implements(rtUndefined)
}

// IsString reports whether the underlying value is a string.
func (v *Value) IsString() bool {
	fmt.Println(v.IndVal)
	fmt.Println(v.IndVal.IsValid())
	return v.IndVal.IsValid() && v.IndVal.Kind() == reflect.String
}

// IsBool reports whether the underlying value is a bool.
func (v *Value) IsBool() bool {
	return v.IndVal.IsValid() && v.IndVal.Kind() == reflect.Bool
}

// IsFloat reports whether the underlying value is a float.
func (v *Value) IsFloat() bool {
	return v.IndVal.IsValid() &&
		(v.IndVal.Kind() == reflect.Float32 ||
			v.IndVal.Kind() == reflect.Float64)
}

// IsInteger reports whether the underlying value is an integer.
func (v *Value) IsInteger() bool {
	if !v.IndVal.IsValid() {
		return false
	}
	kind := v.IndVal.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

// IsNumber reports whether the underlying value is either an integer or a
// float.
func (v *Value) IsNumber() bool {
	return v.IndVal.IsValid() && (v.IsInteger() || v.IsFloat())
}

// IsCallable reports whether the underlying value is a callable function.
func (v *Value) IsCallable() bool {
	return v.IndVal.IsValid() && v.IndVal.Kind() == reflect.Func
}

// IsList reports whether the underlying value is a list.
func (v *Value) IsList() bool {
	return v.IndVal.IsValid() &&
		(v.IndVal.Kind() == reflect.Array ||
			v.IndVal.Kind() == reflect.Slice)
}

// IsDict reports whether the underlying value is a dictionary.
func (v *Value) IsDict() bool {
	return v.IndVal.IsValid() &&
		(v.IndVal.Kind() == reflect.Map ||
			(v.IndVal.Kind() == reflect.Struct && v.IndVal.Type() == TypeDict))
}

// IsIterable reports whether the underlying value is an iterable type. Iterable
// types are strings, lists and dictionaries.
func (v *Value) IsIterable() bool {
	return v.IndVal.IsValid() && (v.IsString() || v.IsList() || v.IsDict())
}

// IsNil reports whether the underlying value is NIL.
func (v *Value) IsNil() bool {
	return !v.IndVal.IsValid()
}

// String returns a string for the underlying value. If this value is not of
// type string, gonja tries to convert it. Currently the following types for
// underlying values are supported:
//
//  1. string
//  2. int/uint (any size)
//  3. float (any precision)
//  4. bool
//  5. array/slice
//  6. map
//  7. String() will be called on the underlying value if provided
//
// NIL values will lead to an empty string. Unsupported types are leading to
// their respective type name.
func (v *Value) String() string {
	if v.IsNil() {
		return ""
	}
	resolved := v.IndVal

	switch resolved.Kind() {
	case reflect.String:
		return resolved.String()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(resolved.Int(), 10)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(resolved.Uint(), 10)

	case reflect.Float32, reflect.Float64:
		formated := strconv.FormatFloat(resolved.Float(), 'f', 11, 64)
		if !strings.Contains(formated, ".") {
			formated = formated + "."
		}
		formated = strings.TrimRight(formated, "0")
		if formated[len(formated)-1] == '.' {
			formated += "0"
		}
		return formated

	case reflect.Bool:
		if v.Bool() {
			return "True"
		}
		return "False"

	case reflect.Struct:
		if t, ok := v.Interface().(fmt.Stringer); ok {
			return t.String()
		}

	case reflect.Slice, reflect.Array:
		var out strings.Builder
		length := v.Len()
		out.WriteByte('[')
		for i := 0; i < length; i++ {
			if i > 0 {
				out.WriteString(", ")
			}
			item := v.Index(i)
			if item.IsString() {
				out.WriteString(fmt.Sprintf(`'%s'`, item.String()))
			} else {
				out.WriteString(item.String())
			}
		}
		out.WriteByte(']')
		return out.String()

	case reflect.Map:
		pairs := []string{}
		for _, key := range resolved.MapKeys() {
			keyLabel := key.String()
			if key.Kind() == reflect.String {
				keyLabel = fmt.Sprintf(`'%s'`, keyLabel)
			}

			value := resolved.MapIndex(key)
			// Check whether this is an interface and resolve it where required
			for value.Kind() == reflect.Interface {
				value = reflect.ValueOf(value.Interface())
			}
			valueLabel := value.String()
			if value.Kind() == reflect.String {
				valueLabel = fmt.Sprintf(`'%s'`, valueLabel)
			}
			pair := fmt.Sprintf(`%s: %s`, keyLabel, valueLabel)
			pairs = append(pairs, pair)
		}
		sort.Strings(pairs)
		return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))

	default:
		errors.ThrowTemplateRuntimeError("Value.String() not implemented for type: %s", resolved.Kind().String())
	}

	return ""
}

// Escaped returns the escaped version of String()
func (v *Value) Escaped() string {
	return u.HTMLEscape(v.String())
}

// Integer returns the underlying value as an integer (converts the underlying
// value, if necessary). If it's not possible to convert the underlying value,
// it will return 0.
func (v *Value) Integer() int {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be converted to integer")
	}

	switch v.IndVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(v.IndVal.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int(v.IndVal.Uint())

	case reflect.Float32, reflect.Float64:
		return int(v.IndVal.Float())

	case reflect.String:
		// Try to convert from string to int (base 10)
		f, err := strconv.ParseFloat(v.IndVal.String(), 64)
		if err != nil {
			return 0
		}
		return int(f)

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be converted to integer", v.IndVal.Kind().String())
	}

	return 0
}

// Float returns the underlying value as a float (converts the underlying
// value, if necessary). If it's not possible to convert the underlying value,
// it will return 0.0.
func (v *Value) Float() float64 {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be converted to float")
	}

	switch v.IndVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.IndVal.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.IndVal.Uint())

	case reflect.Float32, reflect.Float64:
		return v.IndVal.Float()

	case reflect.String:
		// Try to convert from string to float64 (base 10)
		f, err := strconv.ParseFloat(v.IndVal.String(), 64)
		if err != nil {
			return 0.0
		}
		return f

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be converted to float", v.IndVal.Kind().String())
	}

	return 0.0
}

// Bool returns the underlying value as bool. If the value is not bool, false
// will always be returned. If you're looking for true/false-evaluation of the
// underlying value, have a look on the IsTrue()-function.
func (v *Value) Bool() bool {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be converted to bool")
	}

	switch v.IndVal.Kind() {
	case reflect.Bool:
		return v.IndVal.Bool()

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be converted to boolean", v.IndVal.Kind().String())
	}

	return false
}

// IsTrue tries to evaluate the underlying value the Pythonic-way:
//
// Returns TRUE in one the following cases:
//
//   - int != 0
//   - uint != 0
//   - float != 0.0
//   - len(array/chan/map/slice/string) > 0
//   - bool == true
//   - underlying value is a struct
//
// Otherwise returns always FALSE.
func (v *Value) IsTrue() bool {
	if v.IsNil() {
		return false
	}

	switch v.IndVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.IndVal.Int() != 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.IndVal.Uint() != 0

	case reflect.Float32, reflect.Float64:
		return v.IndVal.Float() != 0

	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.IndVal.Len() > 0

	case reflect.Bool:
		return v.IndVal.Bool()

	case reflect.Struct:
		return true // struct instance is always true

	default:
		errors.ThrowTemplateRuntimeError("Value.IsTrue() not available for type: %s", v.IndVal.Kind().String())
	}

	return false
}

// Negate tries to negate the underlying value. It's mainly used for the
// NOT-operator and in conjunction with a call to return_value.IsTrue()
// afterwards.
//
// Example:
//
//	AsValue(1).Negate().IsTrue() == false
func (v *Value) Negate() *Value {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be negated")
	}

	switch v.IndVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Integer() != 0 {
			return AsValue(0)
		}
		return AsValue(1)

	case reflect.Float32, reflect.Float64:
		if v.Float() != 0.0 {
			return AsValue(float64(0.0))
		}
		return AsValue(float64(1.1))

	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return AsValue(v.IndVal.Len() == 0)

	case reflect.Bool:
		return AsValue(!v.IndVal.Bool())

	case reflect.Struct:
		return AsValue(false)

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be negated", v.IndVal.Kind().String())
	}

	return nil
}

// Len returns the length for an array, chan, map, slice or string. Otherwise it
// will return 0.
func (v *Value) Len() int {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil has no length")
	}

	switch v.IndVal.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return v.IndVal.Len()

	case reflect.String:
		runes := []rune(v.IndVal.String())
		return len(runes)

	default:
		errors.ThrowTemplateRuntimeError("type %s has no length", v.IndVal.Kind().String())
	}

	return 0
}

// Slice slices an array, slice or string. Otherwise it will return an empty
// []int.
func (v *Value) Slice(i, j int) *Value {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be sliced")
	}

	switch v.IndVal.Kind() {
	case reflect.Array, reflect.Slice:
		return AsValue(v.IndVal.Slice(i, j).Interface())

	case reflect.String:
		runes := []rune(v.IndVal.String())
		return AsValue(string(runes[i:j]))

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be sliced", v.IndVal.Kind().String())
	}

	return nil
}

// Index gets the i-th item of an array, slice or string. Otherwise it will
// return NIL.
func (v *Value) Index(i int) *Value {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be indexed")
	}

	switch v.IndVal.Kind() {
	case reflect.Array, reflect.Slice:
		if i >= v.Len() {
			return AsValue(nil)
		}
		return AsValue(v.IndVal.Index(i).Interface())

	case reflect.String:
		//return AsValue(v.IndVal.Slice(i, i+1).Interface())
		s := v.IndVal.String()
		runes := []rune(s)
		if i < len(runes) {
			return AsValue(string(runes[i]))
		}
		return AsValue("")

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be indexed", v.IndVal.Kind().String())
	}
	return nil
}

// Contains reports whether the underlying value (which must be of type struct,
// map, string, array or slice) contains of another Value (e. g. used to check
// whether a struct contains of a specific field or a map contains a specific
// key).
//
// Example:
//
//	AsValue("Hello, World!").Contains(AsValue("World")) == true
func (v *Value) Contains(other *Value) bool {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be checked for containment")
	}

	resolved := v.IndVal
	switch resolved.Kind() {
	case reflect.Struct:
		if dict, ok := resolved.Interface().(Dict); ok {
			return dict.Keys().Contains(other)
		}
		fldVal := resolved.FieldByName(other.String())
		return fldVal.IsValid()

	case reflect.Map:
		var mapVal reflect.Value
		// XXX: type checking required, key value must be of same type
		switch other.Interface().(type) {
		case int, string:
			mapVal = resolved.MapIndex(other.IndVal)
		default:
			errors.ThrowTemplateRuntimeError("XXX")
		}
		return mapVal.IsValid()

	case reflect.String:
		return strings.Contains(resolved.String(), other.String())

	case reflect.Slice, reflect.Array:
		if vl, ok := resolved.Interface().(ValuesList); ok {
			return vl.Contains(other)
		}
		for i := 0; i < resolved.Len(); i++ {
			item := resolved.Index(i)
			if other.Interface() == item.Interface() {
				return true
			}
		}
		return false

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be checked for containment", resolved.Kind().String())
	}

	return false
}

// CanSlice reports whether the underlying value is of type array, slice or
// string. You normally would use CanSlice() before using the Slice() operation.
func (v *Value) CanSlice() bool {
	if v.IsNil() {
		return false
	}
	switch v.IndVal.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return true
	}
	return false
}

// Iterate iterates over a map, array, slice or a string. It calls the
// function's first argument for every value with the following arguments:
//
//	idx      current 0-index
//	count    total amount of items
//	key      *Value for the key or item
//	value    *Value (only for maps, the respective value for a specific key)
//
// If the underlying value has no items or is not one of the types above, the
// empty function (function's second argument) will be called.
func (v *Value) Iterate(fn func(idx, count int, key, value *Value) bool, empty func()) {
	v.IterateOrder(fn, empty, false, false, false)
}

// IterateOrder behaves like `Value.Iterate`, but can iterate through an
// array/slice/string in reverse. Does not affect the iteration through a map
// because maps don't have any particular order. However, you can force an order
// using the `sorted` keyword (and even use `reversed sorted`).
func (v *Value) IterateOrder(fn func(idx, count int, key, value *Value) bool, empty func(), reverse bool, sorted bool, caseSensitive bool) {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be iterated")
	}

	resolved := v.IndVal
	switch resolved.Kind() {
	case reflect.Map:
		keys := resolved.MapKeys()
		if sorted {
			if reverse {
				if !caseSensitive {
					// XXX: needs to be implemented
					sort.Sort(sort.Reverse(CaseInsensitive(sortRefelctValuesList(keys))))
				} else {
					sort.Sort(sort.Reverse(sortRefelctValuesList(keys)))
				}
			} else {
				if !caseSensitive {
					sort.Sort(CaseInsensitive(sortRefelctValuesList(keys)))
				} else {
					sort.Sort(sortRefelctValuesList(keys))
				}
			}
		}
		keyLen := len(keys)
		for idx, key := range keys {
			value := v.Val.MapIndex(key)
			if !fn(idx, keyLen, AsValue(key), AsValue(value)) {
				return
			}
		}
		if keyLen == 0 {
			empty()
		}
		return // done

	case reflect.Array, reflect.Slice:
		var items ValuesList

		itemCount := resolved.Len()
		for i := 0; i < itemCount; i++ {
			items = append(items, ToValue(resolved.Index(i)))
		}

		if sorted {
			if reverse {
				if !caseSensitive && items[0].IsString() {
					sort.Slice(items, func(i, j int) bool {
						return strings.ToLower(items[i].String()) > strings.ToLower(items[j].String())
					})
				} else {
					sort.Sort(sort.Reverse(items))
				}
			} else {
				if !caseSensitive && items[0].IsString() {
					sort.Slice(items, func(i, j int) bool {
						return strings.ToLower(items[i].String()) < strings.ToLower(items[j].String())
					})
				} else {
					sort.Sort(items)
				}
			}
		} else {
			if reverse {
				for i := 0; i < itemCount/2; i++ {
					items[i], items[itemCount-1-i] = items[itemCount-1-i], items[i]
				}
			}
		}

		if len(items) > 0 {
			for idx, item := range items {
				if !fn(idx, itemCount, item, nil) {
					return
				}
			}
		} else {
			empty()
		}
		return // done

	case reflect.String:
		if sorted {
			r := []rune(resolved.String())
			if caseSensitive {
				sort.Sort(sortRunes(r))
			} else {
				sort.Sort(CaseInsensitive(sortRunes(r)))
			}
			resolved = reflect.ValueOf(string(r))
		}

		// TODO(flosch): Not utf8-compatible (utf8-decoding necessary)
		charCount := resolved.Len()
		if charCount > 0 {
			if reverse {
				for i := charCount - 1; i >= 0; i-- {
					if !fn(i, charCount, ToValue(resolved.Slice(i, i+1)), nil) {
						return
					}
				}
			} else {
				for i := 0; i < charCount; i++ {
					if !fn(i, charCount, ToValue(resolved.Slice(i, i+1)), nil) {
						return
					}
				}
			}
		} else {
			empty()
		}
		return // done

	case reflect.Chan:
		items := []reflect.Value{}
		for {
			value, ok := resolved.Recv()
			if !ok {
				break
			}
			items = append(items, value)
		}
		count := len(items)
		if count > 0 {
			for idx, value := range items {
				fn(idx, count, ToValue(value), nil)
			}
		} else {
			empty()
		}
		return

	case reflect.Struct:
		if resolved.Type() != TypeDict {
			errors.ThrowTemplateRuntimeError("Value.Iterate() not available for type: %s", resolved.Kind().String())
		}
		dict := resolved.Interface().(Dict)
		keys := dict.Keys()
		length := len(dict.Pairs)
		if sorted {
			if reverse {
				if !caseSensitive {
					sort.Sort(sort.Reverse(CaseInsensitive(keys)))
				} else {
					sort.Sort(sort.Reverse(keys))
				}
			} else {
				if !caseSensitive {
					sort.Sort(CaseInsensitive(keys))
				} else {
					sort.Sort(keys)
				}
			}
		}
		if len(keys) > 0 {
			for idx, key := range keys {
				if !fn(idx, length, key, dict.Get(key)) {
					return
				}
			}
		} else {
			empty()
		}

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be iterated", resolved.Kind().String())
	}

	empty()
}

// Interface returns the underlying value as an interface{}.
func (v *Value) Interface() any {
	if v.Val.IsValid() {
		return v.Val.Interface()
	}
	return nil
}

// EqualValueTo reports whether two values are containing the same value or
// object.
func (v *Value) EqualValueTo(other *Value) bool {
	// comparison of uint with int fails using .Interface()-comparison (see issue #64)
	if v.IsInteger() && other.IsInteger() {
		return v.Integer() == other.Integer()
	}
	return v.Interface() == other.Interface()
}

// Keys returns a list of keys contained in v.
func (v *Value) Keys() ValuesList {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("cannot get keys from nil value")
	}

	keys := ValuesList{}
	resolved := v.IndVal
	if resolved.Type() == TypeDict {
		for _, pair := range resolved.Interface().(Dict).Pairs {
			keys = append(keys, pair.Key)
		}
		return keys
	} else if resolved.Kind() != reflect.Map {
		return keys
	}
	for _, key := range resolved.MapKeys() {
		keys = append(keys, ToValue(key))
	}
	sort.Sort(CaseInsensitive(keys))
	return keys
}

// Items returns a list items contained in v.
func (v *Value) Items() []*Pair {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("cannot get items from nil value")
	}

	out := []*Pair{}
	resolved := v.IndVal
	if resolved.Kind() != reflect.Map {
		return out
	}
	iter := resolved.MapRange()
	for iter.Next() {
		out = append(out, &Pair{
			Key:   ToValue(iter.Key()),
			Value: ToValue(iter.Value()),
		})
	}
	return out
}

// XXX: need to work on that
func (v *Value) Set(key string, value interface{}) {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("can't set attribute or item on nil value")
	}
	val := v.Val
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
		if !val.IsValid() {
			errors.ThrowTemplateRuntimeError("invalid value '%s'", val)
		}
	}

	switch val.Kind() {
	case reflect.Struct:
		field := val.FieldByName(key)
		if !(field.IsValid() && field.CanSet()) {
			errors.ThrowTemplateRuntimeError("can't write field '%s'", key)
		}
		field.Set(reflect.ValueOf(value))

	case reflect.Map:
		val.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))

	default:
		errors.ThrowTemplateRuntimeError("can't set attribute or item on type '%s'", val.Kind())
	}
}

// -----------------------------------------------------------------------------
//
// ValuesList
//
// -----------------------------------------------------------------------------

// ValuesList represents a list of `Value`s.
type ValuesList []*Value

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
func (vl ValuesList) Contains(value *Value) bool {
	for _, val := range vl {
		if value.EqualValueTo(val) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
//
// RefelctValuesList
//
// -----------------------------------------------------------------------------

type sortable interface {
	int64 | uint64 | float64 | string
}

// sortRefelctValuesList returns a sort.Interface that can be used to sort a
// list of `reflect.Value`s.
func sortRefelctValuesList(values []reflect.Value) sort.Interface {
	if len(values) == 0 {
		return nil
	}
	switch values[0].Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return refelctValuesList[int64]{Values: values, GetValueFn: func(v reflect.Value) any { return v.Int() }}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return refelctValuesList[uint64]{Values: values, GetValueFn: func(v reflect.Value) any { return v.Uint() }}
	case reflect.Float32, reflect.Float64:
		return refelctValuesList[float64]{Values: values, GetValueFn: func(v reflect.Value) any { return v.Float() }}
	case reflect.String:
		return refelctValuesList[string]{Values: values, GetValueFn: func(v reflect.Value) any { return v.String() }}
	}
	return nil
}

type refelctValuesList[T sortable] struct {
	Values     []reflect.Value
	GetValueFn func(reflect.Value) any
}

// Len is the number of elements in the collection.
func (vl refelctValuesList[T]) Len() int {
	return len(vl.Values)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (vl refelctValuesList[T]) Less(i, j int) bool {
	vi := vl.GetValueFn(vl.Values[i]).(T)
	vj := vl.GetValueFn(vl.Values[j]).(T)
	return vi < vj
}

// Swap swaps the elements with indexes i and j.
func (vl refelctValuesList[T]) Swap(i, j int) {
	vl.Values[i], vl.Values[j] = vl.Values[j], vl.Values[i]
}

// -----------------------------------------------------------------------------
//
// Dict And Pair Values
//
// -----------------------------------------------------------------------------

// Pair represents a pair of key and value.
type Pair struct {
	Key   *Value
	Value *Value
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

// Dict represents a mapping of key-value `Pair`s.
type Dict struct {
	Pairs []*Pair
}

// NewDict creates a new `Dict`.
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

// Keys returns a `ValueList` of keys contained in the d.
func (d *Dict) Keys() ValuesList {
	keys := ValuesList{}
	for _, pair := range d.Pairs {
		keys = append(keys, pair.Key)
	}
	return keys
}

// Get returns the `Value` for the given key from d.
func (d *Dict) Get(key *Value) *Value {
	for _, pair := range d.Pairs {
		if pair.Key.EqualValueTo(key) {
			return pair.Value
		}
	}
	return AsValue(nil)
}

// TypeDict represents the reflection type of `Dict`.
var TypeDict = reflect.TypeOf(Dict{})

// -----------------------------------------------------------------------------
//
// Utils For Sorting
//
// -----------------------------------------------------------------------------

// sortRunes implements `sort.Interface` for sorting a slice of runes.
type sortRunes []rune

// Len is the number of elements in the collection.
func (s sortRunes) Len() int {
	return len(s)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (s sortRunes) Less(i, j int) bool {
	return s[i] < s[j]
}

// Swap swaps the elements with indexes i and j.
func (s sortRunes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// caseInsensitiveSortedRunes represents a case-insensitive version of
// `sortRunes` for the purpose of sorting.
type caseInsensitiveSortedRunes struct {
	sortRunes
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (ci caseInsensitiveSortedRunes) Less(i, j int) bool {
	return strings.ToLower(string(ci.sortRunes[i])) < strings.ToLower(string(ci.sortRunes[j]))
}

// caseInsensitiveValueList represents a case-insensitive version of
// `ValuesList` for the purpose of sorting.
type caseInsensitiveValueList struct {
	ValuesList
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (ci caseInsensitiveValueList) Less(i, j int) bool {
	vi := ci.ValuesList[i]
	vj := ci.ValuesList[j]
	switch {
	case vi.IsInteger() && vj.IsInteger():
		return vi.Integer() < vj.Integer()
	case vi.IsFloat() && vj.IsFloat():
		return vi.Float() < vj.Float()
	default:
		return strings.ToLower(vi.String()) < strings.ToLower(vj.String())
	}
}

// CaseInsensitive returns the the data sorted in a case insensitive way (if
// string).
func CaseInsensitive(data sort.Interface) sort.Interface {
	if vl, ok := data.(ValuesList); ok {
		return &caseInsensitiveValueList{vl}
	} else if sr, ok := data.(sortRunes); ok {
		return &caseInsensitiveSortedRunes{sr}
	}
	return data
}
