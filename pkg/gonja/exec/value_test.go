package exec_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

type flags struct {
	IsString   bool
	IsCallable bool
	IsBool     bool
	IsFloat    bool
	IsInteger  bool
	IsNumber   bool
	IsList     bool
	IsDict     bool
	IsIterable bool
	IsNil      bool
	Bool       bool
}

func (f *flags) assert(t *testing.T, value exec.Value) {
	assert := testutils.NewAssert(t)

	val := reflect.ValueOf(value)
	fval := reflect.ValueOf(f).Elem()

	for i := 0; i < fval.NumField(); i++ {
		name := fval.Type().Field(i).Name
		method := val.MethodByName(name)
		bVal := fval.Field(i).Interface().(bool)
		result := method.Call([]reflect.Value{})
		bResult := result[0].Interface().(bool)
		if bVal {
			assert.True(bResult, `%s() should be true`, name)
		} else {
			assert.False(bResult, `%s() should be false`, name)
		}
	}
}

var valueCases = []struct {
	name     string
	value    any
	asString string
	flags    flags
}{
	{"nil", nil, "None", flags{IsNil: true}},
	{"string", "Hello World", "Hello World", flags{IsString: true, Bool: true, IsIterable: true}},
	{"int", 42, "42", flags{IsInteger: true, IsNumber: true, Bool: true}},
	{"int 0", 0, "0", flags{IsInteger: true, IsNumber: true}},
	{"float", 42., "42.0", flags{IsFloat: true, IsNumber: true, Bool: true}},
	{"float with trailing zeros", 42.04200, "42.042", flags{IsFloat: true, IsNumber: true, Bool: true}},
	{"float max precision", 42.5556700089099, "42.55567000891", flags{IsFloat: true, IsNumber: true, Bool: true}},
	{"float max precision rounded up", 42.555670008999999, "42.555670009", flags{IsFloat: true, IsNumber: true, Bool: true}},
	{"float 0.0", 0., "0.0", flags{IsFloat: true, IsNumber: true}},
	{"true", true, "True", flags{IsBool: true, Bool: true}},
	{"false", false, "False", flags{IsBool: true}},
	{"slice", []int{1, 2, 3}, "[1, 2, 3]", flags{Bool: true, IsIterable: true, IsList: true}},
	{"strings slice", []string{"a", "b", "c"}, "['a', 'b', 'c']", flags{Bool: true, IsIterable: true, IsList: true}},
	{
		"values slice",
		[]exec.Value{testutils.NewValue(1), testutils.NewValue(2), testutils.NewValue(3)},
		"[1, 2, 3]",
		flags{Bool: true, IsIterable: true, IsList: true},
	},
	{
		"string values slice",
		[]exec.Value{testutils.NewValue("a"), testutils.NewValue("b"), testutils.NewValue("c")},
		"['a', 'b', 'c']",
		flags{Bool: true, IsIterable: true, IsList: true},
	},
	{"array", [3]int{1, 2, 3}, "[1, 2, 3]", flags{Bool: true, IsIterable: true, IsList: true}},
	{"strings array", [3]string{"a", "b", "c"}, "['a', 'b', 'c']", flags{Bool: true, IsIterable: true, IsList: true}},
	{
		"dict as map",
		map[string]string{"a": "a", "b": "b"},
		"{'a': 'a', 'b': 'b'}",
		flags{Bool: true, IsIterable: true, IsDict: true},
	},
	{
		"dict as Dict/Pairs",
		&exec.Dict{[]*exec.Pair{
			{testutils.NewValue("a"), testutils.NewValue("a")},
			{testutils.NewValue("b"), testutils.NewValue("b")},
		}},
		"{'a': 'a', 'b': 'b'}",
		flags{Bool: true, IsIterable: true, IsDict: true},
	},
	{"func", func() {}, "<function()>", flags{IsCallable: true, Bool: true}},
	{"func with args", func(a, b int) {}, "<function(int, int)>", flags{IsCallable: true, Bool: true}},
	{"func with args and return", func(a, b int) int { return 0 }, "<function(int, int) int>", flags{IsCallable: true, Bool: true}},
}

func TestValue(t *testing.T) {
	for _, lc := range valueCases {
		test := lc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			assert := testutils.NewAssert(t)

			value := testutils.NewValue(test.value)

			assert.Equal(test.asString, value.String())
			test.flags.assert(t, value)
		})
	}
}

func TestValueFromMap(t *testing.T) {
	for _, lc := range valueCases {
		test := lc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			assert := testutils.NewAssert(t)

			data := map[string]any{"value": test.value}
			value := testutils.NewValue(data["value"])

			assert.Equal(test.asString, value.String())
			test.flags.assert(t, value)
		})
	}
}

type testStruct struct {
	Attr string
}

func (t testStruct) String() string {
	return t.Attr
}

// var setCases = []struct {
// 	name     string
// 	value    any
// 	attr     string
// 	set      any
// 	error    bool
// 	asString string
// }{
// 	{"nil", nil, "missing", "whatever", true, ""},
// 	{"existing attr on struct by ref", &testStruct{"test"}, "Attr", "value", false, "value"},
// 	{"existing attr on struct by value", testStruct{"test"}, "Attr", "value", true, `Can't write field "Attr"`},
// 	{"missing attr on struct by ref", &testStruct{"test"}, "Missing", "value", true, "test"},
// 	{"missing attr on struct by value", testStruct{"test"}, "Missing", "value", true, "test"},
// 	{
// 		"existing key on map",
// 		map[string]any{"Attr": "test"},
// 		"Attr",
// 		"value",
// 		false,
// 		"{'Attr': 'value'}",
// 	},
// 	{
// 		"new key on map",
// 		map[string]any{"Attr": "test"},
// 		"New",
// 		"value",
// 		false,
// 		"{'Attr': 'test', 'New': 'value'}",
// 	},
// }

// func TestValueSet(t *testing.T) {
// 	for _, lc := range setCases {
// 		test := lc
// 		t.Run(test.name, func(t *testing.T) {
// 			defer func() {
// 				if err := recover(); err != nil {
// 					t.Error(err)
// 				}
// 			}()
// 			assert := testutils.NewAssert(t)

// 			value := testutils.NewValue(test.value)
// 			err := value.Set(test.attr, test.set)

// 			if test.error {
// 				assert.NotNil(err)
// 			} else {
// 				assert.Nil(err)
// 				assert.Equal(test.asString, value.String())
// 			}
// 		})
// 	}
// }

var valueKeysCases = []struct {
	name     string
	value    any
	asString string
	isError  bool
}{
	{"nil", nil, "", true},
	{"string", "Hello World", "", true},
	{"int", 42, "", true},
	{"float", 42., "", true},
	{"true", true, "", true},
	{"false", false, "", true},
	{"slice", []int{1, 2, 3}, "", true},
	// Map keys are sorted alphabetically, case insensitive
	{"dict as map", map[string]string{"c": "c", "a": "a", "b": "b"}, "['a', 'b', 'c']", false},
	// Dict as Pairs keys are kept in order
	{
		"dict as Dict/Pairs",
		&exec.Dict{[]*exec.Pair{
			{testutils.NewValue("c"), testutils.NewValue("c")},
			{testutils.NewValue("a"), testutils.NewValue("a")},
			{testutils.NewValue("b"), testutils.NewValue("b")},
		}},
		"['a', 'b', 'c']",
		false,
	},
	{"func", func() {}, "", true},
}

func TestValueKeys(t *testing.T) {
	for _, lc := range valueKeysCases {
		test := lc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			assert := testutils.NewAssert(t)

			value := testutils.NewValue(test.value)
			if test.isError {
				assert.Panic(func() { value.Keys() })
			} else {
				keys := value.Keys()
				if !value.IsNil() {
					sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
				}
				assert.Equal(test.asString, keys.String())
			}
		})
	}
}
