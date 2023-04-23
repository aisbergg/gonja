package exec

import (
	"reflect"
)

// ValueFactory is a factory for creating values.
type ValueFactory struct {
	// undefinedFn is the function that is called when a value is not found.
	undefinedFn UndefinedFunc

	// customTypes allows to add custom getters for types that are not
	// supported by default. For example, if you want to resolve value from a
	// custom ordered map type, you can add a custom getter for that.
	customTypes map[reflect.Type]ValueFunc

	// customTypesEnabled is true if at least one custom getter is registered.
	customTypesEnabled bool
}

// NewValueFactory creates a new value factory.
func NewValueFactory(undefined UndefinedFunc, customTypes map[reflect.Type]ValueFunc) *ValueFactory {
	customTypesEnabled := (customTypes != nil && len(customTypes) > 0)
	return &ValueFactory{
		undefinedFn:        undefined,
		customTypes:        customTypes,
		customTypesEnabled: customTypesEnabled,
	}
}

// NewValue creates a new value from the given interface.
func (vf *ValueFactory) NewValue(value interface{}, isSafe bool) Value {
	if value == nil {
		return NewNilValue()
	}

	// already a [Value] container -> return as-is
	if v, ok := value.(Value); ok {
		return v
	}

	rflVal := reflect.Value{}
	indVal := reflect.Value{}
	if rv, ok := value.(reflect.Value); ok {
		rflVal = rv
		indVal = indirectReflectValue(rflVal)
	} else {
		rflVal = reflect.ValueOf(value)
		indVal = indirectReflectValue(rflVal)
	}
	typ := rflVal.Type()

	// check for user defined types
	if vf.customTypesEnabled {
		for cstTyp, fn := range vf.customTypes {
			if typ == cstTyp {
				return fn(value, isSafe, vf)
			}
		}
	}

	if !indVal.IsValid() {
		return NewNilValue()
	}

	// fallback to generic value implementation
	return &GenericValue{
		BaseValue: BaseValue{
			valueFactory: vf,
			isSafe:       isSafe,
		},
		Value:         rflVal,
		IndirectValue: indVal,
		valueType:     typ,
	}
}

// NewUndefined creates a new undefined value.
func (vf *ValueFactory) NewUndefined(name string, hintFormat string, args ...any) Undefined {
	return vf.undefinedFn(name, hintFormat, args...)
}

// func (vf *ValueFactory) Slice(i, j int) Value {}
// func (vf *ValueFactory) Index(i int) Value    {}
// func (vf *ValueFactory) IterateOrder(fn func(idx, count int, key, value Value) bool, empty func(), reverse bool, sorted bool, caseSensitive bool) {
// }
// func (vf *ValueFactory) Interface() any                                    {}
// func (vf *ValueFactory) GetItem(key any, valueFactory *ValueFactory) Value {}

// ToValueFunc converts a reflect.Value to a Value container.
// type ToValueFunc func(val interface{}, safe bool) Value

// // getWithCustom uses the provided custom converters to copy the value.
// func (r *ValueFactory) getWithCustom(val reflect.Value, typ reflect.Type, key any) (ret reflect.Value, ok, usedGetter bool) {
// 	if getter, ok := r.customGetters[typ]; ok {
// 		ret, ok = getter(val, key)
// 		return ret, ok, true
// 	}
// 	return
// }
