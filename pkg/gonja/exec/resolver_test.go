package exec_test

import (
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
	"github.com/aisbergg/gonja/pkg/gonja"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

// var getattrCases = []struct {
// 	name     string
// 	value    any
// 	attr     string
// 	found    bool
// 	asString string
// 	flags    flags
// }{
// 	{"nil", nil, "missing", false, "", flags{IsError: true}},
// 	{"attr found", testStruct{"test"}, "Attr", true, "test", flags{IsString: true, IsTrue: true, IsIterable: true}},
// 	{"attr not found", testStruct{"test"}, "Missing", false, "", flags{IsNil: true}},
// 	{"item", map[string]any{"Attr": "test"}, "Attr", false, "", flags{IsNil: true}},
// }

// func TestValueGetAttr(t *testing.T) {
// 	for _, lc := range getattrCases {
// 		test := lc
// 		t.Run(test.name, func(t *testing.T) {
// 			defer func() {
// 				if err := recover(); err != nil {
// 					t.Error(err)
// 				}
// 			}()
// 			assert := testutils.NewAssert(t)

// 			value := exec.AsValue(test.value)
// 			out, found := value.Getattr(test.attr)

// 			if !test.flags.IsError && err {
// 				t.Fatalf("Unexpected error: %s", err)
// 			}

// 			if test.found {
// 				assert.True(found, "Attribute '%s' should be found on %s", test.attr, value)
// 				assert.Equal(test.asString, out.String())
// 			} else {
// 				assert.False(found, "Attribute '%s' should not be found on %s", test.attr, value)
// 			}

// 			test.flags.assert(t, out)
// 		})
// 	}
// }

var getitemCases = []struct {
	name     string
	value    any
	key      any
	found    bool
	asString string
	flags    flags
}{
	// {"nil", nil, "missing", false, "", flags{IsError: true}},
	// {"item found", map[string]any{"Attr": "test"}, "Attr", true, "test", flags{IsString: true, IsTrue: true, IsIterable: true}},
	// {"item not found", map[string]any{"Attr": "test"}, "Missing", false, "test", flags{IsNil: true}},
	// {"attr", testStruct{"test"}, "Attr", false, "", flags{IsNil: true}},
	// {"dict found", &exec.Dict{[]*exec.Pair{
	// 	{exec.AsValue("key"), exec.AsValue("value")},
	// 	{exec.AsValue("otherKey"), exec.AsValue("otherValue")},
	// }}, "key", true, "value", flags{IsTrue: true, IsString: true, IsIterable: true}},
}

func TestValueGetItem(t *testing.T) {
	assert := testutils.NewAssert(t)
	for _, lc := range getitemCases {
		test := lc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()

			value := exec.AsValue(test.value)
			out, found := getValue(value, test.key)

			if test.found {
				assert.True(found, "Key '%s' should be found on %s", test.key, value)
				assert.Equal(test.asString, out.String())
			} else {
				assert.False(found, "Key '%s' should not be found on %s", test.key, value)
			}

			// XXX: panics
			test.flags.assert(t, out)
		})
	}
}

func getValue(val *exec.Value, key any) (*exec.Value, bool) {
	resolver := exec.NewResolver(gonja.Undefined, nil)
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(errors.TemplateRuntimeError); ok {
				return
			}
			panic(r)
		}
	}()

	resolved := resolver.Get(val, key)
	if _, ok := resolved.Interface().(exec.Undefined); ok {
		return nil, false
	}
	return resolved, true
}
