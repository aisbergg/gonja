package exec

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	log "github.com/aisbergg/gonja/internal/log/exec"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// Resolver allows to resolve values from different types of variables.
type Resolver struct {
	// undefinedFunc is the function that is called when a value is not found.
	undefinedFunc UndefinedFunc
}

// NewResolver creates a new resolver.
func NewResolver(undefined UndefinedFunc) *Resolver {
	return &Resolver{
		undefinedFunc: undefined,
	}
}

// GetItem returns the value for the given key. If 'value' has no such key, the
// undefined value is returned.
func (r *Resolver) GetItem(value *Value, key any) *Value {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("try to get item '%s' from %s", key, value.Val.Kind().String())

	if value.IsNil() {
		log.Print("get item '%s' from invalid or nil value -> return undefined", key)
		return AsValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
	}

	val := value.Val
	typ := value.Val.Type()
	if typ.Implements(undefinedType) {
		return AsValue(value.Val.Interface().(Undefined).GetItem(key))
	}

	var resVal reflect.Value
	if index, ok := key.(int); ok {
		val = value.IndVal
		switch val.Kind() {
		case reflect.String, reflect.Array, reflect.Slice:
			if index >= val.Len() {
				log.Print("index '%v' out of range -> return undefined", index)
				return AsValue(r.undefinedFunc(strconv.Itoa(index), "%s has no element %d", val.Kind().String(), index))
			}
			if index < 0 {
				index = val.Len() + index
			}
			if index < 0 {
				log.Print("index '%v' out of range -> return undefined", index)
				return AsValue(r.undefinedFunc(strconv.Itoa(index), "%s has no element %d", val.Kind().String(), index))
			}
			resVal = val.Index(index)

		case reflect.Map:
			resVal = val.MapIndex(reflect.ValueOf(index))
			if !resVal.IsValid() {
				log.Print("map has no key '%v' -> return undefined", index)
				return AsValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
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
				return AsValue(r.undefinedFunc(name, ""))
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
				return AsValue(r.undefinedFunc(name, "struct has no field '%s'", name))
			}
			resVal = val.Field(fld.Index)

		default:
			log.Print("cannot get item '%s' from '%s' value -> return undefined", name, val.Kind().String())
			return AsValue(r.undefinedFunc(name, ""))
		}

	} else {
		val = value.IndVal
		switch val.Kind() {
		case reflect.Map:
			resVal = val.MapIndex(reflect.ValueOf(key))
			if !resVal.IsValid() {
				log.Print("map has no key '%v' -> return undefined", key)
				return AsValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
			}

		default:
			log.Print("get item '%v' from '%s' value -> return undefined", key, val.Kind().String())
			return AsValue(r.undefinedFunc(fmt.Sprintf("%s", key), ""))
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
