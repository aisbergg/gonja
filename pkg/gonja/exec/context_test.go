package exec_test

import (
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

var ctxCases = []struct {
	name     string
	value    any
	asString string
	flags    flags
}{
	{"nil", nil, "", flags{IsNil: true}},
	{"string", "Hello World", "Hello World", flags{IsString: true, Bool: true, IsIterable: true}},
	{"int", 42, "42", flags{IsInteger: true, IsNumber: true, Bool: true}},
	{"int 0", 0, "0", flags{IsInteger: true, IsNumber: true}},
	{"float", 42., "42.000000", flags{IsFloat: true, IsNumber: true, Bool: true}},
	{"float 0.0", 0., "0.000000", flags{IsFloat: true, IsNumber: true}},
	{"true", true, "True", flags{IsBool: true, Bool: true}},
	{"false", false, "False", flags{IsBool: true}},
}

func TestContext(t *testing.T) {
	for _, cc := range ctxCases {
		test := cc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			assert := testutils.NewAssert(t)

			ctx := exec.NewEmptyContext(testutils.ValueVactory)
			ctx.Set(test.name, test.value)
			value := ctx.Get(test.name)

			if test.value != nil {
				assert.Equal(test.value, value.Interface())
			}
			test.flags.assert(t, value)
		})
	}
}

func TestSubContext(t *testing.T) {
	for _, cc := range ctxCases {
		test := cc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			assert := testutils.NewAssert(t)

			ctx := exec.NewEmptyContext(testutils.ValueVactory)
			ctx.Set(test.name, test.value)
			sub := ctx.Inherit()
			value := sub.Get(test.name)

			if test.value != nil {
				assert.Equal(test.value, value.Interface())
			}
			test.flags.assert(t, value)
		})
	}
}

func TestFuncContext(t *testing.T) {
	ctx := exec.NewEmptyContext(testutils.ValueVactory)
	ctx.Set("func", func() {})

	cases := []struct {
		name string
		ctx  *exec.Context
	}{
		{"top context", ctx},
		{"sub context", ctx.Inherit()},
	}

	for _, c := range cases {
		test := c
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			assert := testutils.NewAssert(t)

			value := test.ctx.Get("func")
			if assert.True(value.IsCallable()) {
				_, ok := value.Interface().(func())
				assert.True(ok)
			}
		})
	}
}

// func TestValueFromMap(t *testing.T) {
// 	for _, lc := range valueCases {
// 		test := lc
// 		t.Run(test.name, func(t *testing.T) {
// 			defer func() {
// 				if err := recover(); err != nil {
// 					t.Error(err)
// 				}
// 			}()
// 			assert := testutils.NewAssert(t)

// 			data := map[string]any{"value": test.value}
// 			value := exec.AsValue(data["value"])

// 			assert.Equal(test.asString, value.String())
// 			test.flags.assert(t, value)
// 		})
// 	}
// }

// type testStruct struct {
// 	Attr string
// }

// var getattrCases = []struct {
// 	name     string
// 	value    any
// 	attr     string
// 	found    bool
// 	asString string
// 	flags    flags
// }{
// 	{"nil", nil, "missing", false, "", flags{IsError: true}},
// 	{"attr found", testStruct{"test"}, "Attr", true, "test", flags{IsString: true, IsTrue: true}},
// 	{"item", map[string]any{"Attr": "test"}, "Attr", false, "", flags{IsError: true}},
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

// 			if !test.flags.IsError && out.IsError() {
// 				t.Fatalf(`Unexpected error: %s`, out.Error())
// 			}

// 			if test.found {
// 				assert.Truef(found, `Attribute '%s' should be found on %s`, test.attr, value)
// 				assert.Equal(test.asString, out.String())
// 			} else {
// 				assert.Falsef(found, `Attribute '%s' should not be found on %s`, test.attr, value)
// 			}

// 			test.flags.assert(t, out)
// 		})
// 	}
// }

// var getitemCases = []struct {
// 	name     string
// 	value    any
// 	key      any
// 	found    bool
// 	asString string
// 	flags    flags
// }{
// 	{"nil", nil, "missing", false, "", flags{IsError: true}},
// 	{"item found", map[string]any{"Attr": "test"}, "Attr", true, "test", flags{IsString: true, IsTrue: true}},
// 	{"attr", testStruct{"test"}, "Attr", false, "", flags{IsError: true}},
// }

// func TestValueGetItem(t *testing.T) {
// 	for _, lc := range getitemCases {
// 		test := lc
// 		t.Run(test.name, func(t *testing.T) {
// 			defer func() {
// 				if err := recover(); err != nil {
// 					t.Error(err)
// 				}
// 			}()
// 			assert := testutils.NewAssert(t)

// 			value := exec.AsValue(test.value)
// 			out, found := value.GetItem(test.key)

// 			if !test.flags.IsError && out.IsError() {
// 				t.Fatalf(`Unexpected error: %s`, out.Error())
// 			}

// 			if test.found {
// 				assert.Truef(found, `Key '%s' should be found on %s`, test.key, value)
// 				assert.Equal(test.asString, out.String())
// 			} else {
// 				assert.Falsef(found, `Key '%s' should not be found on %s`, test.key, value)
// 			}

// 			test.flags.assert(t, out)
// 		})
// 	}
// }
