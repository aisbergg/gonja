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

// Value creates a new [Value] container from the given value.
func (vf *ValueFactory) Value(value any) Value {
	return vf.asValue(value, false)
}

// SafeValue creates a new [Value] container from the given value and marks
// it as safe.
func (vf *ValueFactory) SafeValue(value any) Value {
	return vf.asValue(value, true)
}

// asValue converts the given value to a [Value] container.
func (vf *ValueFactory) asValue(value any, isSafe bool) Value {
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
func (vf *ValueFactory) NewUndefined(name, hintFormat string, args ...any) Undefined {
	return vf.undefinedFn(name, hintFormat, args...)
}
