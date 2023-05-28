package exec

import (
	"fmt"
	"html"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	debug "github.com/aisbergg/gonja/internal/debug/exec"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

var _ Value = (*GenericValue)(nil)

// GenericValue is a container for values of various types.
type GenericValue struct {
	BaseValue

	// Value holds the actual value in form of a reflection value.
	Value reflect.Value

	// IndirectValue holds the indirect (resolved) value . This is used to avoid
	// resolving the pointer values more than once.
	IndirectValue reflect.Value

	// precomputed to improve performance
	valueType reflect.Type
}

// Type returns the type of the value.
func (v *GenericValue) Type() reflect.Type {
	return v.valueType
}

// IsString reports whether the underlying value is a string.
func (v *GenericValue) IsString() bool {
	return v.IndirectValue.IsValid() && v.IndirectValue.Kind() == reflect.String
}

// IsBool reports whether the underlying value is a bool.
func (v *GenericValue) IsBool() bool {
	return v.IndirectValue.IsValid() && v.IndirectValue.Kind() == reflect.Bool
}

// IsFloat reports whether the underlying value is a float.
func (v *GenericValue) IsFloat() bool {
	return v.IndirectValue.IsValid() &&
		(v.IndirectValue.Kind() == reflect.Float32 ||
			v.IndirectValue.Kind() == reflect.Float64)
}

// IsInteger reports whether the underlying value is an integer.
func (v *GenericValue) IsInteger() bool {
	if !v.IndirectValue.IsValid() {
		return false
	}
	kind := v.IndirectValue.Kind()
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
func (v *GenericValue) IsNumber() bool {
	return v.IndirectValue.IsValid() && (v.IsInteger() || v.IsFloat())
}

// IsList reports whether the underlying value is a list.
func (v *GenericValue) IsList() bool {
	return v.IndirectValue.IsValid() &&
		(v.IndirectValue.Kind() == reflect.Array ||
			v.IndirectValue.Kind() == reflect.Slice)
}

// IsDict reports whether the underlying value is a dictionary.
func (v *GenericValue) IsDict() bool {
	return v.IndirectValue.IsValid() &&
		(v.IndirectValue.Kind() == reflect.Map ||
			(v.IndirectValue.Kind() == reflect.Struct && v.valueType == rtDict))
}

// IsNil reports whether the underlying value is NIL.
func (v *GenericValue) IsNil() bool {
	return !v.IndirectValue.IsValid()
}

// IsCallable reports whether the underlying value is a callable function.
func (v *GenericValue) IsCallable() bool {
	return v.IndirectValue.IsValid() && v.IndirectValue.Kind() == reflect.Func
}

// IsIterable reports whether the underlying value is an iterable type. Iterable
// types are strings, lists and dictionaries.
func (v *GenericValue) IsIterable() bool {
	return v.IndirectValue.IsValid() && (v.IsString() || v.IsList() || v.IsDict())
}

// IsSliceable reports whether the underlying value is of type array, slice or
// string. You normally would use IsSliceable() before using the Slice() operation.
func (v *GenericValue) IsSliceable() bool {
	if v.IsNil() {
		return false
	}
	switch v.IndirectValue.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return true
	}

	// unreachable
	return false
}

// Interface returns the underlying value as an interface{}.
func (v *GenericValue) Interface() any {
	if v.Value.IsValid() {
		return v.Value.Interface()
	}
	return nil
}

// ReflectValue returns the underlying reflect value.
func (v *GenericValue) ReflectValue() reflect.Value {
	return v.Value
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
func (v *GenericValue) String() string {
	if v.IsNil() {
		return "None"
	}
	resolved := v.IndirectValue

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
		val := v.Interface()
		// standard time format used by Python's datetime module
		if t, ok := val.(time.Time); ok {
			if t.Nanosecond() == 0 {
				// omit nanoseconds if not set
				return t.Format("2006-01-02 15:04:05-07:00")
			}
			return t.Format("2006-01-02 15:04:05.000000-07:00")
		}
		// try with String() method
		if t, ok := val.(fmt.Stringer); ok {
			return t.String()
		}
		return fmt.Sprintf("<%s>", v.valueType)

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

	case reflect.Func:
		// format function like this:
		//   <function(arg1, arg2, argN)>
		//   <function(arg1, arg2, argN) ret>
		//   <function(arg1, arg2, argN) (ret1, re2)>
		typ := resolved.Type()
		if typ.NumIn() == 0 {
			return "<function()>"
		}
		// get args
		args := []string{}
		for i := 0; i < typ.NumIn(); i++ {
			if typ.IsVariadic() && i == typ.NumIn()-1 {
				args = append(args, fmt.Sprintf("...%s", typ.In(i).String()))
				break
			}
			args = append(args, typ.In(i).String())
		}
		// get return values
		rets := []string{}
		if typ.NumOut() > 0 {
			for i := 0; i < typ.NumOut(); i++ {
				rets = append(rets, typ.Out(i).String())
			}
		}
		// format
		if len(rets) > 1 {
			return fmt.Sprintf("<function(%s) (%s)>", strings.Join(args, ", "), strings.Join(rets, ", "))
		} else if len(rets) == 1 {
			return fmt.Sprintf("<function(%s) %s>", strings.Join(args, ", "), rets[0])
		}
		return fmt.Sprintf("<function(%s)>", strings.Join(args, ", "))

	default:
		errors.ThrowTemplateRuntimeError("Value.String() not implemented for type: %s", resolved.Kind().String())
	}

	return ""
}

// Escaped returns the HTML escaped version of String()
func (v *GenericValue) Escaped() string {
	return html.EscapeString(v.String())
	// return u.HTMLEscape(v.String())
}

// Integer returns the underlying value as an integer (converts the underlying
// value, if necessary). If it's not possible to convert the underlying value,
// it will return 0.
func (v *GenericValue) Integer() int {
	if v.IsNil() {
		return 0
	}

	switch v.IndirectValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(v.IndirectValue.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int(v.IndirectValue.Uint())

	case reflect.Float32, reflect.Float64:
		return int(v.IndirectValue.Float())

	case reflect.String:
		// Try to convert from string to int (base 10)
		f, err := strconv.ParseFloat(v.IndirectValue.String(), 64)
		if err != nil {
			return 0
		}
		return int(f)

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be converted to integer", v.IndirectValue.Kind().String())
	}

	// unreachable
	return 0
}

// Float returns the underlying value as a float (converts the underlying
// value, if necessary). If it's not possible to convert the underlying value,
// it will return 0.0.
func (v *GenericValue) Float() float64 {
	if v.IsNil() {
		return 0.0
	}

	switch v.IndirectValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.IndirectValue.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.IndirectValue.Uint())

	case reflect.Float32, reflect.Float64:
		return v.IndirectValue.Float()

	case reflect.String:
		// Try to convert from string to float64 (base 10)
		f, err := strconv.ParseFloat(v.IndirectValue.String(), 64)
		if err != nil {
			return 0.0
		}
		return f

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be converted to float", v.IndirectValue.Kind().String())
	}

	// unreachable
	return 0.0
}

// Bool returns the underlying value as bool. Non bool values will be evaluated
// to true in the following cases:
//
//   - int != 0
//   - uint != 0
//   - float != 0.0
//   - len(array/chan/map/slice/string) > 0
//   - bool == true
//   - underlying value is a struct
//
// Otherwise returns always FALSE.
func (v *GenericValue) Bool() bool {
	if v.IsNil() {
		return false
	}

	switch v.IndirectValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.IndirectValue.Int() != 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.IndirectValue.Uint() != 0

	case reflect.Float32, reflect.Float64:
		return v.IndirectValue.Float() != 0

	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.IndirectValue.Len() > 0

	case reflect.Bool:
		return v.IndirectValue.Bool()

	case reflect.Struct:
		return true // struct instance is always true

	case reflect.Func:
		return true // function is always true

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be evaluated to boolean", v.IndirectValue.Kind().String())
	}

	// unreachable
	return false
}

// Len returns the length for an array, chan, map, slice or string. Otherwise it
// will return 0.
func (v *GenericValue) Len() int {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil has no length")
	}

	switch v.IndirectValue.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return v.IndirectValue.Len()

	case reflect.String:
		runes := []rune(v.IndirectValue.String())
		return len(runes)

	default:
		errors.ThrowTemplateRuntimeError("type %s has no length", v.IndirectValue.Kind().String())
	}

	return 0
}

// Slice slices an array, slice or string. Otherwise it will return an empty
// []int.
func (v *GenericValue) Slice(i, j int) Value {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be sliced")
	}

	switch v.IndirectValue.Kind() {
	case reflect.Array, reflect.Slice:
		return v.valueFactory.Value(v.IndirectValue.Slice(i, j))

	case reflect.String:
		runes := []rune(v.IndirectValue.String())
		return v.valueFactory.Value(string(runes[i:j]))

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be sliced", v.IndirectValue.Kind().String())
	}

	// unreachable
	return &GenericValue{}
}

func (v *GenericValue) Index(i int) Value {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be indexed")
	}

	switch v.IndirectValue.Kind() {
	case reflect.Array, reflect.Slice:
		if i < 0 {
			i = v.Len() + i
		}
		if i >= v.Len() || i < 0 {
			return v.valueFactory.Value(reflect.Zero(v.IndirectValue.Type()).Interface())
		}
		return v.valueFactory.Value(v.IndirectValue.Index(i).Interface())

	case reflect.String:
		s := v.IndirectValue.String()
		if i >= len(s) {
			return v.valueFactory.Value("")
		}
		runes := []rune(s)
		if i >= 0 {
			if i >= len(runes) {
				return v.valueFactory.Value("")
			}
			return v.valueFactory.Value(string(runes[i]))
		}
		i = len(runes) + i
		if i < 0 {
			return v.valueFactory.Value("")
		}
		return v.valueFactory.Value(string(runes[i]))

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be indexed", v.IndirectValue.Kind().String())
	}

	// unreachable
	return nil
}

// Contains reports whether the underlying value (which must be of type struct,
// map, string, array or slice) contains of another Value (e. g. used to check
// whether a struct contains of a specific field or a map contains a specific
// key).
func (v *GenericValue) Contains(other Value) bool {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be checked for containment")
	}

	resolved := v.IndirectValue
	switch resolved.Kind() {
	case reflect.Struct:
		if dict, ok := v.Value.Interface().(*Dict); ok {
			return dict.Keys().Contains(other)
		}
		fldVal := resolved.FieldByName(other.String())
		return fldVal.IsValid()

	case reflect.Map:
		wantType := resolved.Type().Key()
		otherVal := indirectReflectValue(other.ReflectValue())
		otherType := otherVal.Type()
		if !otherType.AssignableTo(wantType) {
			errors.ThrowTemplateRuntimeError("type %s cannot be used as map key of type %s", otherType.String(), wantType.String())
		}
		mapVal := resolved.MapIndex(otherVal)
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

	// unreachable
	return false
}

// Keys returns a list of keys contained in v.
func (v *GenericValue) Keys() ValuesList {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("cannot get keys from nil value")
	}

	keys := ValuesList{}
	if v.valueType == rtDict {
		for _, pair := range v.Value.Interface().(*Dict).Pairs {
			keys = append(keys, pair.Key)
		}
		return keys

	} else if v.IndirectValue.Kind() == reflect.Map {
		for _, key := range v.IndirectValue.MapKeys() {
			keys = append(keys, v.valueFactory.Value(key))
		}
		return keys
	}

	errors.ThrowTemplateRuntimeError("cannot get keys from value of type %s", v.valueType.String())
	return nil
}

// Values returns a list of values contained in v.
func (v *GenericValue) Values() ValuesList {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("cannot get values from nil value")
	}

	values := ValuesList{}
	if v.valueType == rtDict {
		for _, pair := range v.Value.Interface().(*Dict).Pairs {
			values = append(values, pair.Value)
		}
		return values

	} else if v.IndirectValue.Kind() == reflect.Map {
		iter := v.IndirectValue.MapRange()
		for iter.Next() {
			values = append(values, v.valueFactory.Value(iter.Value()))
		}
		return values
	}

	errors.ThrowTemplateRuntimeError("cannot get values from value of type %s", v.valueType.String())
	return nil
}

// Items returns a list items contained in v.
func (v *GenericValue) Items() []*Pair {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("cannot get items from nil value")
	}

	items := []*Pair{}
	if v.valueType == rtDict {
		return v.Value.Interface().(Dict).Pairs
	} else if v.IndirectValue.Kind() == reflect.Map {
		iter := v.IndirectValue.MapRange()
		for iter.Next() {
			items = append(items, &Pair{
				Key:   v.valueFactory.Value(iter.Key()),
				Value: v.valueFactory.Value(iter.Value()),
			})
		}
		return items
	}

	errors.ThrowTemplateRuntimeError("cannot get items from value of type %s", v.valueType.String())
	return nil
}

// Get returns the value for the given key. If 'value' has no such key, the
// undefined value is returned.
func (v *GenericValue) GetItem(key any) Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("try to get item '%s' from %s", key, v.Value.Kind().String())

	if v.IsNil() {
		debug.Print("get item '%s' from invalid or nil value -> return undefined", key)
		return v.valueFactory.NewUndefined(fmt.Sprintf("%s", key), "")
	}

	var resVal reflect.Value
	if index, ok := key.(int); ok {
		val := v.IndirectValue
		switch val.Kind() {
		case reflect.String, reflect.Array, reflect.Slice:
			if index >= val.Len() {
				debug.Print("index '%v' out of range -> return undefined", index)
				return v.valueFactory.NewUndefined(strconv.Itoa(index), "%s has no element %d", val.Kind().String(), index)
			}
			if index < 0 {
				index = val.Len() + index
			}
			if index < 0 {
				debug.Print("index '%v' out of range -> return undefined", index)
				return v.valueFactory.NewUndefined(strconv.Itoa(index), "%s has no element %d", val.Kind().String(), index)
			}
			resVal = val.Index(index)

		case reflect.Map:
			resVal = val.MapIndex(reflect.ValueOf(index))
			if !resVal.IsValid() {
				debug.Print("map has no key '%v' -> return undefined", index)
				return v.valueFactory.NewUndefined(fmt.Sprintf("%s", key), "")
			}

		default:
			errors.ThrowTemplateRuntimeError("can't access an index on type %s", val.Kind().String())
		}

	} else if name, ok := key.(string); ok {
		// check if value has a method with the given name
		val := v.Value.MethodByName(name)
		if val.IsValid() {
			return v.valueFactory.Value(val)
		}

		val = v.IndirectValue
		switch val.Kind() {
		case reflect.Map:
			resVal = val.MapIndex(reflect.ValueOf(name))
			if !resVal.IsValid() {
				debug.Print("map has no key '%s' -> return undefined", name)
				return v.valueFactory.NewUndefined(name, "map has no key '%s'", name)
			}

		case reflect.Struct:
			if debug.Enabled {
				if v.valueType == rtValue {
					panic(fmt.Errorf("[BUG] Value was wrapped in a Value"))
				} else if v.valueType == rtValue {
					panic(fmt.Errorf("[BUG] reflect.Value was wrapped in a reflect.Value"))
				}
			}

			// special handling for dict type
			if v.valueType == rtDict {
				dict := v.Value.Interface().(*Dict)
				resVal, ok := dict.Get(v.valueFactory.Value(name))
				if !ok {
					debug.Print("dict has no key '%s' -> return undefined", name)
					return v.valueFactory.NewUndefined(name, "dict has no key '%s'", name)
				}
				return resVal
			}

			structFlds := getStructFields(val)
			fld, ok := structFlds[name]
			if !ok {
				debug.Print("struct has no field '%s' -> return undefined", name)
				return v.valueFactory.NewUndefined(name, "struct has no field '%s'", name)
			}
			resVal = val.Field(fld.Index)

		default:
			debug.Print("cannot get item '%s' from '%s' value -> return undefined", name, val.Kind().String())
			return v.valueFactory.NewUndefined(name, "")
		}

	} else {
		val := v.IndirectValue
		switch val.Kind() {
		case reflect.Map:
			resVal = val.MapIndex(reflect.ValueOf(key))
			if !resVal.IsValid() {
				debug.Print("map has no key '%v' -> return undefined", key)
				return v.valueFactory.NewUndefined(fmt.Sprintf("%s", key), "")
			}

		default:
			debug.Print("get item '%v' from '%s' value -> return undefined", key, val.Kind().String())
			return v.valueFactory.NewUndefined(fmt.Sprintf("%s", key), "")
		}
	}

	if !resVal.CanInterface() {
		errors.ThrowTemplateRuntimeError("cannot get value for key '%s'", key)
	}
	debug.Print("return value")
	if resVal.Type() == rtValue {
		return resVal.Interface().(Value)
	}
	return v.valueFactory.Value(resVal)
}

// XXX: need to work on that
func (v *GenericValue) SetItem(key string, value interface{}) {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("can't set attribute or item on nil value")
	}
	val := v.Value
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
func (v *GenericValue) Iterate(fn func(idx, count int, key, value Value) (cont bool), empty func()) {
	v.IterateOrder(fn, empty, false, false, false)
}

// IterateOrder behaves like [Value.Iterate], but can iterate through an
// array/slice/string in reverse. Does not affect the iteration through a map
// because maps don't have any particular order. However, you can force an order
// using the `sorted` keyword (and even use `reversed sorted`).
func (v *GenericValue) IterateOrder(
	fn func(idx, count int, key, value Value) (cont bool),
	empty func(),
	reverse bool,
	sorted bool,
	caseSensitive bool,
) {
	if v.IsNil() {
		errors.ThrowTemplateRuntimeError("nil cannot be iterated")
	}

	rflVal := v.IndirectValue
	switch rflVal.Kind() {
	case reflect.Map:
		keys := rflVal.MapKeys()
		keysCount := len(keys)
		if keysCount == 0 {
			empty()
			return
		}

		if sorted {
			sortKeys := sortRefelctValues(keys, caseSensitive)
			if reverse {
				sort.Sort(sort.Reverse(sortKeys))
			} else {
				sort.Sort(sortKeys)
			}

		} else if reverse {
			// reverse order without sorting
			for i := keysCount - 1; i >= 0; i-- {
				key := keys[i]
				value := rflVal.MapIndex(key)
				if !fn(keysCount-i-1, keysCount, v.valueFactory.Value(key), v.valueFactory.Value(value)) {
					return
				}
			}
			return
		}

		for idx, key := range keys {
			value := rflVal.MapIndex(key)
			if !fn(idx, keysCount, v.valueFactory.Value(key), v.valueFactory.Value(value)) {
				return
			}
		}

		return // done

	case reflect.Array, reflect.Slice:
		var items ValuesList
		var itemCount int
		if v.valueType == rtValuesList {
			items = v.Value.Interface().(ValuesList)
			itemCount = len(items)
		} else {
			itemCount = rflVal.Len()
			items = make(ValuesList, 0, itemCount)
			for i := 0; i < itemCount; i++ {
				items = append(items, v.valueFactory.Value(rflVal.Index(i).Interface()))
			}
		}
		if itemCount == 0 {
			empty()
			return
		}

		if sorted {
			sortItems := sortValuesList(items, caseSensitive)
			if reverse {
				sort.Sort(sort.Reverse(sortItems))
			} else {
				sort.Sort(sortItems)
			}

		} else if reverse {
			// reverse order without sorting
			for i := itemCount - 1; i >= 0; i-- {
				if !fn(itemCount-i-1, itemCount, items[i], nil) {
					return
				}
			}
			return
		}

		for i, item := range items {
			if !fn(i, itemCount, item, nil) {
				return
			}
		}

		return // done

	case reflect.String:
		runes := []rune(rflVal.String())
		runesCount := len(runes)
		if runesCount == 0 {
			empty()
			return
		}

		if sorted {
			sortRunes := sortRunes(runes, caseSensitive)
			if reverse {
				sort.Sort(sort.Reverse(sortRunes))
			} else {
				sort.Sort(sortRunes)
			}

		} else if reverse {
			// reverse order without sorting
			for i := runesCount - 1; i >= 0; i-- {
				if !fn(runesCount-i-1, runesCount, v.valueFactory.Value(string(runes[i])), nil) {
					return
				}
			}
			return
		}

		for i, r := range runes {
			if !fn(i, runesCount, v.valueFactory.Value(string(r)), nil) {
				return
			}
		}

		return // done

	case reflect.Chan:
		items := []reflect.Value{}
		for {
			value, ok := rflVal.Recv()
			if !ok {
				break
			}
			items = append(items, value)
		}
		count := len(items)
		if count == 0 {
			empty()
			return
		}
		for idx, value := range items {
			fn(idx, count, v.valueFactory.Value(value), nil)
		}

		return // done

	case reflect.Struct:
		if v.valueType != rtDict {
			errors.ThrowTemplateRuntimeError("Value.Iterate() not available for type: %s", rflVal.Kind().String())
		}
		dict := v.Value.Interface().(*Dict)
		keys := dict.Keys()
		keysCount := len(dict.Pairs)
		if keysCount == 0 {
			empty()
			return
		}

		if sorted {
			sortKeys := sortValuesList(keys, caseSensitive)
			if reverse {
				sort.Sort(sort.Reverse(sortKeys))
			} else {
				sort.Sort(sortKeys)
			}

		} else if reverse {
			// reverse order without sorting
			for i := keysCount - 1; i >= 0; i-- {
				key := keys[i]
				item, _ := dict.Get(key)
				if !fn(keysCount-i-1, keysCount, key, item) {
					return
				}
			}
			return
		}

		for i, key := range keys {
			item, _ := dict.Get(key)
			if !fn(i, keysCount, key, item) {
				return
			}
		}

		return // done

	default:
		errors.ThrowTemplateRuntimeError("type %s cannot be iterated", rflVal.Kind().String())
	}
}

// EqualValueTo reports whether two values are containing the same value or
// object.
func (v *GenericValue) EqualValueTo(other Value) bool {
	// comparison of uint with int fails using .Interface()-comparison
	if v.IsInteger() && other.IsInteger() {
		return v.Integer() == other.Integer()
	}
	return v.Interface() == other.Interface()
}

// -----------------------------------------------------------------------------
//
// Helpers
//
// -----------------------------------------------------------------------------

type structField struct {
	Index int
	Name  string
}

var structFieldsCache = make(map[reflect.Type]map[string]structField)

// getStructFields returns the (filtered) fields of the given struct. It caches
// the result in order to speed up subsequent calls.
func getStructFields(structVal reflect.Value) map[string]structField {
	typ := structVal.Type()
	if cached, ok := structFieldsCache[typ]; ok {
		return cached
	}

	structFlds := make(map[string]structField, structVal.NumField())
	for i := 0; i < structVal.NumField(); i++ {
		// the struct field `PkgPath` is empty for exported fields
		fld := structVal.Type().Field(i)
		if fld.PkgPath != "" {
			continue
		}
		sf := structField{
			Index: i,
			Name:  fld.Name,
		}
		// add field name as is
		structFlds[fld.Name] = sf
		// add field name with json tag as key
		if gonjaTag := fld.Tag.Get("gonja"); gonjaTag != "" && gonjaTag != "-" {
			var commaIdx int
			if commaIdx = strings.Index(gonjaTag, ","); commaIdx < 0 {
				commaIdx = len(gonjaTag)
			}
			fieldName := gonjaTag[:commaIdx]
			if fieldName != "" {
				structFlds[fieldName] = sf
			}
		}
	}

	// cache fields info
	structFieldsCache[typ] = structFlds

	return structFlds
}

type sortable interface {
	int64 | uint64 | float64 | string
}

// sortRefelctValues returns a sort.Interface that can be used to sort a
// list of [reflect.Value]s.
func sortRefelctValues(values []reflect.Value, caseSensitive bool) sort.Interface {
	if len(values) == 0 {
		return nil
	}
	switch values[0].Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return sortrefelctValuesImpl[int64]{Values: values, GetValueFn: func(v reflect.Value) any { return v.Int() }}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return sortrefelctValuesImpl[uint64]{Values: values, GetValueFn: func(v reflect.Value) any { return v.Uint() }}
	case reflect.Float32, reflect.Float64:
		return sortrefelctValuesImpl[float64]{Values: values, GetValueFn: func(v reflect.Value) any { return v.Float() }}
	case reflect.String:
		if caseSensitive {
			return sortrefelctValuesImpl[string]{Values: values, GetValueFn: func(v reflect.Value) any { return v.String() }}
		}
		return sortStringValuesCaseInsensitive(values)
	}
	// check if value type implements sort.Interface
	if sortInterface, ok := values[0].Interface().(sort.Interface); ok {
		return sortInterface
	}
	// can't sort it
	return nil
}

type sortrefelctValuesImpl[T sortable] struct {
	Values     []reflect.Value
	GetValueFn func(reflect.Value) any
}

// Len is the number of elements in the collection.
func (vl sortrefelctValuesImpl[T]) Len() int {
	return len(vl.Values)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (vl sortrefelctValuesImpl[T]) Less(i, j int) bool {
	vi := vl.GetValueFn(vl.Values[i]).(T)
	vj := vl.GetValueFn(vl.Values[j]).(T)
	return vi < vj
}

// Swap swaps the elements with indexes i and j.
func (vl sortrefelctValuesImpl[T]) Swap(i, j int) {
	vl.Values[i], vl.Values[j] = vl.Values[j], vl.Values[i]
}

// sortStringValuesCaseInsensitive returns a sort.Interface that can be used to
// sort a list of strings wrapped in reflect.Value in a case insensitive way.
func sortStringValuesCaseInsensitive(values []reflect.Value) sort.Interface {
	return sortStringValuesCaseInsensitiveImpl{Values: values}
}

// sortStringValuesCaseInsensitiveImpl is a sort.Interface implementation for
// sorting a list of strings wrapped in reflect.Value in a case insensitive way.
type sortStringValuesCaseInsensitiveImpl struct {
	Values []reflect.Value
}

// Len is the number of elements in the collection.
func (sl sortStringValuesCaseInsensitiveImpl) Len() int {
	return len(sl.Values)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (sl sortStringValuesCaseInsensitiveImpl) Less(i, j int) bool {
	a := sl.Values[i].String()
	b := sl.Values[j].String()
	return stringLessCaseInsensitive(a, b)
}

func stringLessCaseInsensitive(a, b string) bool {
	for {
		if len(b) == 0 {
			return false
		}
		if len(a) == 0 {
			return true
		}
		ad, aSize := utf8.DecodeRuneInString(a)
		bd, bSize := utf8.DecodeRuneInString(b)

		aLower := unicode.ToLower(ad)
		bLower := unicode.ToLower(bd)

		if aLower < bLower {
			return true
		}
		if aLower > bLower {
			return false
		}

		a = a[aSize:]
		b = b[bSize:]
	}
}

// Swap swaps the elements with indexes i and j.
func (sl sortStringValuesCaseInsensitiveImpl) Swap(i, j int) {
	sl.Values[i], sl.Values[j] = sl.Values[j], sl.Values[i]
}

// sortRunes returns a sort.Interface that can be used to sort a
// list of runes.
func sortRunes(runes []rune, caseSensitive bool) sort.Interface {
	if len(runes) == 0 {
		return nil
	}
	if caseSensitive {
		return sortRunesCaseSensitiveImpl(runes)
	}
	return sortRunesCaseInsensitive(runes)
}

// sortRunesCaseSensitiveImpl implements [sort.Interface] for sorting a slice of runes.
type sortRunesCaseSensitiveImpl []rune

// Len is the number of elements in the collection.
func (s sortRunesCaseSensitiveImpl) Len() int {
	return len(s)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (s sortRunesCaseSensitiveImpl) Less(i, j int) bool {
	return s[i] < s[j]
}

// Swap swaps the elements with indexes i and j.
func (s sortRunesCaseSensitiveImpl) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// sortRunesCaseInsensitive returns a [sort.Interface] for sorting a slice of
// runes.
func sortRunesCaseInsensitive(runes []rune) sortRunesCaseInsensitiveImpl {
	lowercased := make([]rune, len(runes))
	for i, r := range runes {
		lowercased[i] = unicode.ToLower(r)
	}
	return sortRunesCaseInsensitiveImpl{runes: runes, lowercased: lowercased}
}

// sortRunes implements [sort.Interface] for sorting a slice of runes.
type sortRunesCaseInsensitiveImpl struct {
	runes      []rune
	lowercased []rune
}

// Len is the number of elements in the collection.
func (s sortRunesCaseInsensitiveImpl) Len() int {
	return len(s.runes)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (s sortRunesCaseInsensitiveImpl) Less(i, j int) bool {
	return s.lowercased[i] < s.lowercased[j]
}

// Swap swaps the elements with indexes i and j.
func (s sortRunesCaseInsensitiveImpl) Swap(i, j int) {
	s.runes[i], s.runes[j] = s.runes[j], s.runes[i]
	s.lowercased[i], s.lowercased[j] = s.lowercased[j], s.lowercased[i]
}

// sortValuesList  returns a [sort.Interface] for sorting a [ValuesList].
func sortValuesList(data ValuesList, caseSensitive bool) sort.Interface {
	return sortValuesListImpl{valuesList: data, caseSensitive: caseSensitive}
}

// sortValuesListImpl implements [sort.Interface] for sorting a [ValuesList].
type sortValuesListImpl struct {
	valuesList    ValuesList
	caseSensitive bool
}

// Len is the number of elements in the collection.
func (ci sortValuesListImpl) Len() int {
	return len(ci.valuesList)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (ci sortValuesListImpl) Less(i, j int) bool {
	vi := ci.valuesList[i]
	vj := ci.valuesList[j]
	switch {
	case vi.IsInteger() && vj.IsInteger():
		return vi.Integer() < vj.Integer()
	case vi.IsFloat() && vj.IsFloat():
		return vi.Float() < vj.Float()
	default:
		a := vi.String()
		b := vj.String()
		if ci.caseSensitive {
			return a < b
		}
		return stringLessCaseInsensitive(a, b)
	}
}

// Swap swaps the elements with indexes i and j.
func (ci sortValuesListImpl) Swap(i, j int) {
	ci.valuesList[i], ci.valuesList[j] = ci.valuesList[j], ci.valuesList[i]
}
