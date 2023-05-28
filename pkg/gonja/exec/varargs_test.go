package exec_test

import (
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

func TestVarArgs(t *testing.T) {
	t.Run("first", testVAFirst)
	t.Run("GetKwarg", testVAGetKwarg)
	t.Run("expect", testVAExpect)
}

func testVAFirst(t *testing.T) {
	t.Run("nil if empty", func(t *testing.T) {
		assert := testutils.NewAssert(t)

		va := exec.VarArgs{}
		first := va.First()
		assert.True(first.IsNil())
	})
	t.Run("first value", func(t *testing.T) {
		assert := testutils.NewAssert(t)

		va := exec.VarArgs{Args: []exec.Value{testutils.NewValue(42)}}
		first := va.First()
		assert.Equal(42, first.Integer())
	})
}

func testVAGetKwarg(t *testing.T) {
	t.Run("value if found", func(t *testing.T) {
		assert := testutils.NewAssert(t)

		va := exec.VarArgs{Kwargs: []exec.KVPair{
			{"key", testutils.NewValue(42)},
		}}
		kwarg := va.GetKwarg("key")
		assert.Equal(42, kwarg.Integer())
	})
	t.Run("defaut if missing", func(t *testing.T) {
		assert := testutils.NewAssert(t)

		va := exec.VarArgs{}
		assert.Panic(func() {
			va.GetKwarg("missing")
		})
	})
}

var nothingCases = []struct {
	name  string
	va    *exec.VarArgs
	error string
}{
	{
		"got nothing",
		testutils.NewVarArgs(nil, nil),
		"",
	}, {
		"got an argument",
		testutils.NewVarArgs([]exec.Value{testutils.NewValue(42)}, nil),
		"expected no arguments, got 1",
	}, {
		"got multiples arguments",
		testutils.NewVarArgs([]exec.Value{testutils.NewValue(42), testutils.NewValue(7)}, nil),
		"expected no arguments, got 2",
	}, {
		"got a keyword argument",
		testutils.NewVarArgs(nil, []exec.KVPair{{"key", testutils.NewValue(42)}}),
		"expected no arguments, got 1",
	}, {
		"got multiple keyword arguments",
		testutils.NewVarArgs(nil, []exec.KVPair{{"key", testutils.NewValue(42)}, {"other", testutils.NewValue(7)}}),
		"expected no arguments, got 2",
	}, {
		"got one of each",
		testutils.NewVarArgs([]exec.Value{testutils.NewValue(42)}, []exec.KVPair{{"key", testutils.NewValue(42)}}),
		"expected no arguments, got 2",
	},
}

var argsCases = []struct {
	name  string
	va    *exec.VarArgs
	args  int
	error string
}{
	{
		"got expected",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42), testutils.NewValue(7)},
			nil,
		),
		2, "",
	}, {
		"got less arguments",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42)},
			nil,
		),
		2, "expected 2 arguments, got 1",
	}, {
		"got less arguments (singular)",
		testutils.NewVarArgs(nil, nil),
		1, "expected an argument, got 0",
	}, {
		"got more arguments",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42), testutils.NewValue(7)},
			nil,
		),
		1, "unexpected argument '7'",
	}, {
		"got a keyword argument",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42)},
			[]exec.KVPair{{"key", testutils.NewValue(42)}},
		),
		1, "unexpected keyword argument 'key=42'",
	},
}

var kwargsCases = []struct {
	name   string
	va     *exec.VarArgs
	kwargs []*exec.Kwarg
	error  string
}{
	{
		"got expected",
		testutils.NewVarArgs(
			nil,
			[]exec.KVPair{{"key", testutils.NewValue(42)}, {"other", testutils.NewValue(7)}},
		),
		[]*exec.Kwarg{
			{"key", "default key"},
			{"other", "default other"},
		},
		"",
	}, {
		"got unexpected arguments",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42), testutils.NewValue(7), testutils.NewValue("unexpected")},
			nil,
		),
		[]*exec.Kwarg{
			{"key", "default key"},
			{"other", "default other"},
		},
		"unexpected argument 'unexpected'",
	}, {
		"got an unexpected keyword argument",
		testutils.NewVarArgs(
			nil,
			[]exec.KVPair{{"unknown", testutils.NewValue(42)}},
		),
		[]*exec.Kwarg{
			{"key", "default key"},
			{"other", "default other"},
		},
		"unexpected keyword argument 'unknown=42'",
	}, {
		"got multiple keyword arguments",
		testutils.NewVarArgs(
			nil,
			[]exec.KVPair{
				{"unknown", testutils.NewValue(42)},
				{"seven", testutils.NewValue(7)},
			},
		),
		[]*exec.Kwarg{
			{"key", "default key"},
			{"other", "default other"},
		},
		"unexpected keyword arguments 'seven=7, unknown=42'",
	},
}

var mixedArgsKwargsCases = []struct {
	name     string
	va       *exec.VarArgs
	args     int
	kwargs   []*exec.Kwarg
	expected *exec.VarArgs
	error    string
}{
	{
		"got expected",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42)},
			[]exec.KVPair{
				{"key", testutils.NewValue(42)},
				{"other", testutils.NewValue(7)},
			},
		),
		1,
		[]*exec.Kwarg{
			{"key", "default key"},
			{"other", "default other"},
		},
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42)},
			[]exec.KVPair{
				{"key", testutils.NewValue(42)},
				{"other", testutils.NewValue(7)},
			},
		),
		"",
	},
	{
		"fill with default",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42)},
			nil,
		),
		1,
		[]*exec.Kwarg{
			{"key", "default key"},
			{"other", "default other"},
		},
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42)},
			[]exec.KVPair{
				{"key", testutils.NewValue("default key")},
				{"other", testutils.NewValue("default other")},
			},
		),
		"",
	},
	{
		"keyword as argument",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42), testutils.NewValue(42)},
			[]exec.KVPair{
				{"other", testutils.NewValue(7)},
			},
		),
		1,
		[]*exec.Kwarg{
			{"key", "default key"},
			{"other", "default other"},
		},
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42)},
			[]exec.KVPair{
				{"key", testutils.NewValue(42)},
				{"other", testutils.NewValue(7)},
			},
		),
		"",
	},
	{
		"keyword submitted twice",
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42), testutils.NewValue(5)},
			[]exec.KVPair{
				{"key", testutils.NewValue(42)},
				{"other", testutils.NewValue(7)},
			},
		),
		1,
		[]*exec.Kwarg{
			{"key", "default key"},
			{"other", "default other"},
		},
		testutils.NewVarArgs(
			[]exec.Value{testutils.NewValue(42), testutils.NewValue(5)},
			[]exec.KVPair{
				{"key", testutils.NewValue(42)},
				{"other", testutils.NewValue(7)},
			},
		),
		"got multiple values for argument 'key'",
	},
}

func assertError(t *testing.T, rva *exec.ReducedVarArgs, expected string) {
	assert := testutils.NewAssert(t)
	if len(expected) > 0 {
		if assert.True(rva.IsError(), "Should have returned an error") {
			assert.Equal(expected, rva.Error())
		}
	} else {
		assert.False(rva.IsError(), "unexpected error: %s", rva.Error())
	}
}

func testVAExpect(t *testing.T) {
	t.Run("nothing", func(t *testing.T) {
		for _, tc := range nothingCases {
			test := tc
			t.Run(test.name, func(t *testing.T) {
				rva := test.va.ExpectNothing()
				assertError(t, rva, test.error)
			})
		}
	})
	t.Run("arguments", func(t *testing.T) {
		for _, tc := range argsCases {
			test := tc
			t.Run(test.name, func(t *testing.T) {
				rva := test.va.ExpectArgs(test.args)
				assertError(t, rva, test.error)
			})
		}
	})
	t.Run("keyword arguments", func(t *testing.T) {
		for _, tc := range kwargsCases {
			test := tc
			t.Run(test.name, func(t *testing.T) {
				rva := test.va.Expect(0, test.kwargs)
				assertError(t, rva, test.error)
			})
		}
	})
	t.Run("mixed arguments", func(t *testing.T) {
		for _, tc := range mixedArgsKwargsCases {
			test := tc
			t.Run(test.name, func(t *testing.T) {
				assert := testutils.NewAssert(t)
				rva := test.va.Expect(test.args, test.kwargs)
				assertError(t, rva, test.error)
				if assert.Equal(len(test.expected.Args), len(rva.Args)) {
					for idx, expected := range test.expected.Args {
						arg := rva.Args[idx]
						assert.Equal(expected.Interface(), arg.Interface(),
							"Argument %d mismatch: expected '%s' got '%s'",
							idx, expected.String(), arg.String(),
						)
					}
				}
				if assert.Equal(len(test.expected.Kwargs), len(rva.Kwargs)) {
					for _, expectedKV := range test.expected.Kwargs {
						expKey, expValue := expectedKV.Key, expectedKV.Value
						if assert.True(rva.HasKwarg(expKey)) {
							value := rva.GetKwarg(expKey)
							assert.Equal(expValue.Interface(), value.Interface(),
								"Keyword argument %s mismatch: expected '%s' got '%s'",
								expKey, expValue.String(), value.String(),
							)
						}
					}
				}
			})
		}
	})
}
