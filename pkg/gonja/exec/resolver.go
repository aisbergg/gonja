package exec

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	log "github.com/aisbergg/gonja/internal/log/exec"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

type CustomGetter func(value reflect.Value, key any) (ret reflect.Value, ok bool)

// Resolver allows to resolve values from different types of variables.
type Resolver struct {
	// undefinedFunc is the function that is called when a value is not found.
	undefinedFunc UndefinedFunc

	// customGetters allows to add custom getters for types that are not
	// supported by default. For example, if you want to resolve value from a
	// custom ordered map type, you can add a custom getter for that.
	customGetters map[reflect.Type]CustomGetter

	// customGettersEnabled is true if at least one custom getter is registered.
	customGettersEnabled bool
}

// NewResolver creates a new resolver.
func NewResolver(undefined UndefinedFunc, customGetters map[reflect.Type]CustomGetter) *Resolver {
	customGettersEnabled := (customGetters != nil && len(customGetters) > 0)
	return &Resolver{
		undefinedFunc:        undefined,
		customGetters:        customGetters,
		customGettersEnabled: customGettersEnabled,
	}
}

// Get returns the value for the given key. If 'value' has no such key, the
// undefined value is returned.
func (r *Resolver) Get(value *Value, key any) *Value {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("try to get item '%s' from %s", key, value.Val.Kind().String())

	if value.IsNil() {
		log.Print("get item '%s' from invalid or nil value -> return undefined", key)
		return toUndefinedValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
	}

	val := value.Val
	typ := value.Val.Type()
	if typ.Implements(undefinedType) {
		return AsValue(value.Val.Interface().(Undefined).Get(key))
	}

	// try to use user defined custom getters
	if r.customGettersEnabled {
		if value, ok, usedGetter := r.getWithCustom(val, typ, key); usedGetter {
			if ok {
				return ToValue(value)
			} else {
				return toUndefinedValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
			}
		}
	}

	var resVal reflect.Value
	if index, ok := key.(int); ok {
		val = value.IndVal
		switch val.Kind() {
		case reflect.String, reflect.Array, reflect.Slice:
			if index >= val.Len() {
				log.Print("index '%v' out of range -> return undefined", index)
				return toUndefinedValue(r.undefinedFunc(strconv.Itoa(index), "%s has no element %d", val.Kind().String(), index))
			}
			if index < 0 {
				index = val.Len() + index
			}
			if index < 0 {
				log.Print("index '%v' out of range -> return undefined", index)
				return toUndefinedValue(r.undefinedFunc(strconv.Itoa(index), "%s has no element %d", val.Kind().String(), index))
			}
			resVal = val.Index(index)

		case reflect.Map:
			resVal = val.MapIndex(reflect.ValueOf(index))
			if !resVal.IsValid() {
				log.Print("map has no key '%v' -> return undefined", index)
				return toUndefinedValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
			}

		default:
			errors.ThrowTemplateRuntimeError("can't access an index on type %s", val.Kind().String())
		}

	} else if name, ok := key.(string); ok {
		// check if value has a method with the given name
		val = value.Val.MethodByName(name)
		if val.IsValid() {
			return ToValue(val)
		}

		val = value.IndVal
		switch val.Kind() {
		case reflect.Map:
			resVal = val.MapIndex(reflect.ValueOf(name))
			if !resVal.IsValid() {
				log.Print("map has no key '%s' -> return undefined", name)
				return toUndefinedValue(r.undefinedFunc(name, ""))
			}

		case reflect.Struct:
			if log.Enabled {
				if typ == rtValue {
					panic(fmt.Errorf("BUG: *Value was wrapped in a *Value"))
				} else if typ == rtValue {
					panic(fmt.Errorf("BUG: reflect.Value was wrapped in a reflect.Value"))
				}
			}

			structFlds := getStructFields(val)
			fld, ok := structFlds[name]
			if !ok {
				log.Print("struct has no field '%s' -> return undefined", name)
				return toUndefinedValue(r.undefinedFunc(name, "struct has no field '%s'", name))
			}
			resVal = val.Field(fld.Index)

		default:
			log.Print("cannot get item '%s' from '%s' value -> return undefined", name, val.Kind().String())
			return toUndefinedValue(r.undefinedFunc(name, ""))
		}

	} else {
		val = value.IndVal
		switch val.Kind() {
		case reflect.Map:
			resVal = val.MapIndex(reflect.ValueOf(key))
			if !resVal.IsValid() {
				log.Print("map has no key '%v' -> return undefined", key)
				return toUndefinedValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
			}

		default:
			log.Print("get item '%v' from '%s' value -> return undefined", key, val.Kind().String())
			return toUndefinedValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
		}
	}

	if !resVal.CanInterface() {
		errors.ThrowTemplateRuntimeError("cannot get value for key '%s'", key)

	}
	log.Print("return value")
	if resVal.Type() == rtValue {
		return resVal.Interface().(*Value)
	}
	return ToValue(resVal)
}

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

// getWithCustom uses the provided custom converters to copy the value.
func (r *Resolver) getWithCustom(val reflect.Value, typ reflect.Type, key any) (ret reflect.Value, ok, usedGetter bool) {
	if getter, ok := r.customGetters[typ]; ok {
		ret, ok = getter(val, key)
		return ret, ok, true
	}
	return
}

func (r *Resolver) IterateOrder(value *Value, fn func(idx, count int, key, value *Value) bool, empty func(), reverse bool, sorted bool, caseSensitive bool) {

}
