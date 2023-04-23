package parse_test

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var testCases = []struct {
	name     string
	text     string
	expected specs
}{
	{"comment", "{# My comment #}", specs{parse.CommentNode{}, attrs{
		"Text": val{" My comment "},
	}}},
	{"multiline comment", "{# My\nmultiline\ncomment #}", specs{parse.CommentNode{}, attrs{
		"Text": val{" My\nmultiline\ncomment "},
	}}},
	{"empty comment", "{##}", specs{parse.CommentNode{}, attrs{
		"Text": val{""},
	}}},
	{"raw text", "raw text", specs{parse.DataNode{}, attrs{
		"Data": _token("raw text"),
	}}},
	// Literals
	{"single quotes string", "{{ 'test' }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.StringNode{}, "test"),
	}}},
	{"single quotes string with whitespace chars", "{{ '  \n\ttest' }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.StringNode{}, "  \n\ttest"),
	}}},
	{"single quotes string with raw whitespace chars", `{{ '  \n\ttest' }}`, specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.StringNode{}, "  \n\ttest"),
	}}},
	{"double quotes string", `{{ "test" }}`, specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.StringNode{}, "test"),
	}}},
	{"double quotes string with whitespace chars", "{{ \"  \n\ttest\" }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.StringNode{}, "  \n\ttest"),
	}}},
	{"double quotes string with raw whitespace chars", `{{ "  \n\ttest" }}`, specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.StringNode{}, "  \n\ttest"),
	}}},
	{"single quotes inside double quotes string", `{{ "'quoted' test" }}`, specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.StringNode{}, "'quoted' test"),
	}}},
	{"integer", "{{ 42 }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.IntegerNode{}, int64(42)),
	}}},
	{"negative-integer", "{{ -42 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.UnaryExpressionNode{}, attrs{
			"Negative": val{true},
			"Term":     _literal(parse.IntegerNode{}, int64(42)),
		}},
	}}},
	{"float", "{{ 42.0 }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.FloatNode{}, float64(42)),
	}}},
	{"negative-float", "{{ -42.0 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.UnaryExpressionNode{}, attrs{
			"Negative": val{true},
			"Term":     _literal(parse.FloatNode{}, float64(42)),
		}},
	}}},
	{"bool-true", "{{ true }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.BoolNode{}, true),
	}}},
	{"bool-True", "{{ True }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.BoolNode{}, true),
	}}},
	{"bool-false", "{{ false }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.BoolNode{}, false),
	}}},
	{"bool-False", "{{ False }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.BoolNode{}, false),
	}}},
	{"list", "{{ ['list', \"of\", 'objects'] }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.ListNode{}, slice{
			_literal(parse.StringNode{}, "list"),
			_literal(parse.StringNode{}, "of"),
			_literal(parse.StringNode{}, "objects"),
		}),
	}}},
	{"list with trailing coma", "{{ ['list', \"of\", 'objects',] }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.ListNode{}, slice{
			_literal(parse.StringNode{}, "list"),
			_literal(parse.StringNode{}, "of"),
			_literal(parse.StringNode{}, "objects"),
		}),
	}}},
	{"single entry list", "{{ ['list'] }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.ListNode{}, slice{
			_literal(parse.StringNode{}, "list"),
		}),
	}}},
	{"empty list", "{{ [] }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.ListNode{}, slice{}),
	}}},
	{"tuple", "{{ ('tuple', \"of\", 'objects') }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.TupleNode{}, slice{
			_literal(parse.StringNode{}, "tuple"),
			_literal(parse.StringNode{}, "of"),
			_literal(parse.StringNode{}, "objects"),
		}),
	}}},
	{"tuple with trailing coma", "{{ ('tuple', \"of\", 'objects',) }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.TupleNode{}, slice{
			_literal(parse.StringNode{}, "tuple"),
			_literal(parse.StringNode{}, "of"),
			_literal(parse.StringNode{}, "objects"),
		}),
	}}},
	{"single entry tuple", "{{ ('tuple',) }}", specs{parse.OutputNode{}, attrs{
		"Expression": _literal(parse.TupleNode{}, slice{
			_literal(parse.StringNode{}, "tuple"),
		}),
	}}},
	{"empty dict", "{{ {} }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.DictNode{}, attrs{}},
	}}},
	{"dict string", "{{ {'dict': 'of', 'key': 'and', 'value': 'pairs'} }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.DictNode{}, attrs{
			"Pairs": slice{
				specs{parse.PairNode{}, attrs{
					"Key":   _literal(parse.StringNode{}, "dict"),
					"Value": _literal(parse.StringNode{}, "of"),
				}},
				specs{parse.PairNode{}, attrs{
					"Key":   _literal(parse.StringNode{}, "key"),
					"Value": _literal(parse.StringNode{}, "and"),
				}},
				specs{parse.PairNode{}, attrs{
					"Key":   _literal(parse.StringNode{}, "value"),
					"Value": _literal(parse.StringNode{}, "pairs"),
				}},
			},
		}},
	}}},
	{"dict int", "{{ {1: 'one', 2: 'two', 3: 'three'} }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.DictNode{}, attrs{
			"Pairs": slice{
				specs{parse.PairNode{}, attrs{
					"Key":   _literal(parse.IntegerNode{}, int64(1)),
					"Value": _literal(parse.StringNode{}, "one"),
				}},
				specs{parse.PairNode{}, attrs{
					"Key":   _literal(parse.IntegerNode{}, int64(2)),
					"Value": _literal(parse.StringNode{}, "two"),
				}},
				specs{parse.PairNode{}, attrs{
					"Key":   _literal(parse.IntegerNode{}, int64(3)),
					"Value": _literal(parse.StringNode{}, "three"),
				}},
			},
		}},
	}}},
	{"addition", "{{ 40 + 2 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left":     _literal(parse.IntegerNode{}, int64(40)),
			"Right":    _literal(parse.IntegerNode{}, int64(2)),
			"Operator": _binOp("+"),
		}},
	}}},
	{"multiple additions", "{{ 40 + 1 + 1 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left": specs{parse.BinaryExpressionNode{}, attrs{
				"Left":     _literal(parse.IntegerNode{}, int64(40)),
				"Right":    _literal(parse.IntegerNode{}, int64(1)),
				"Operator": _binOp("+"),
			}},
			"Right":    _literal(parse.IntegerNode{}, int64(1)),
			"Operator": _binOp("+"),
		}},
	}}},
	{"multiple additions with power", "{{ 40 + 2 ** 1 + 0 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left": specs{parse.BinaryExpressionNode{}, attrs{
				"Left": _literal(parse.IntegerNode{}, int64(40)),
				"Right": specs{parse.BinaryExpressionNode{}, attrs{
					"Left":     _literal(parse.IntegerNode{}, int64(2)),
					"Right":    _literal(parse.IntegerNode{}, int64(1)),
					"Operator": _binOp("**"),
				}},
				"Operator": _binOp("+"),
			}},
			"Right":    _literal(parse.IntegerNode{}, int64(0)),
			"Operator": _binOp("+"),
		}},
	}}},
	{"substract", "{{ 40 - 2 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left":     _literal(parse.IntegerNode{}, int64(40)),
			"Right":    _literal(parse.IntegerNode{}, int64(2)),
			"Operator": _binOp("-"),
		}},
	}}},
	{"complex math", "{{ -1 * (-(-(10-100)) ** 2) ** 3 + 3 * (5 - 17) + 1 + 2 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left": specs{parse.BinaryExpressionNode{}, attrs{
				"Left": specs{parse.BinaryExpressionNode{}, attrs{
					"Left": specs{parse.BinaryExpressionNode{}, attrs{
						"Left": specs{parse.UnaryExpressionNode{}, attrs{
							"Negative": val{true},
							"Term":     _literal(parse.IntegerNode{}, int64(1)),
						}},
						"Right": specs{parse.BinaryExpressionNode{}, attrs{
							"Left": specs{parse.UnaryExpressionNode{}, attrs{
								"Negative": val{true},
								"Term": specs{parse.BinaryExpressionNode{}, attrs{
									"Left": specs{parse.UnaryExpressionNode{}, attrs{
										"Negative": val{true},
										"Term": specs{parse.BinaryExpressionNode{}, attrs{
											"Left":     _literal(parse.IntegerNode{}, int64(10)),
											"Right":    _literal(parse.IntegerNode{}, int64(100)),
											"Operator": _binOp("-"),
										}},
									}},
									"Right":    _literal(parse.IntegerNode{}, int64(2)),
									"Operator": _binOp("**"),
								}},
							}},
							"Right":    _literal(parse.IntegerNode{}, int64(3)),
							"Operator": _binOp("**"),
						}},
						"Operator": _binOp("*"),
					}},
					"Right": specs{parse.BinaryExpressionNode{}, attrs{
						"Left": _literal(parse.IntegerNode{}, int64(3)),
						"Right": specs{parse.BinaryExpressionNode{}, attrs{
							"Left":     _literal(parse.IntegerNode{}, int64(5)),
							"Right":    _literal(parse.IntegerNode{}, int64(17)),
							"Operator": _binOp("-"),
						}},
						"Operator": _binOp("*"),
					}},
					"Operator": _binOp("+"),
				}},
				"Right":    _literal(parse.IntegerNode{}, int64(1)),
				"Operator": _binOp("+"),
			}},
			"Right":    _literal(parse.IntegerNode{}, int64(2)),
			"Operator": _binOp("+"),
		}},
	}}},
	{"negative-expression", "{{ -(40 + 2) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.UnaryExpressionNode{}, attrs{
			"Negative": val{true},
			"Term": specs{parse.BinaryExpressionNode{}, attrs{
				"Left":     _literal(parse.IntegerNode{}, int64(40)),
				"Right":    _literal(parse.IntegerNode{}, int64(2)),
				"Operator": _binOp("+"),
			}},
		}},
	}}},
	{"Operators precedence", "{{ 2 * 3 + 4 % 2 + 1 - 2 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left": specs{parse.BinaryExpressionNode{}, attrs{
				"Left": specs{parse.BinaryExpressionNode{}, attrs{
					"Left": specs{parse.BinaryExpressionNode{}, attrs{
						"Left":     _literal(parse.IntegerNode{}, int64(2)),
						"Right":    _literal(parse.IntegerNode{}, int64(3)),
						"Operator": _binOp("*"),
					}},
					"Right": specs{parse.BinaryExpressionNode{}, attrs{
						"Left":     _literal(parse.IntegerNode{}, int64(4)),
						"Right":    _literal(parse.IntegerNode{}, int64(2)),
						"Operator": _binOp("%"),
					}},
					"Operator": _binOp("+"),
				}},
				"Right":    _literal(parse.IntegerNode{}, int64(1)),
				"Operator": _binOp("+"),
			}},
			"Right":    _literal(parse.IntegerNode{}, int64(2)),
			"Operator": _binOp("-"),
		}},
	}}},
	{"Operators precedence with parenthesis", "{{ 2 * (3 + 4) % 2 + (1 - 2) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left": specs{parse.BinaryExpressionNode{}, attrs{
				"Left": specs{parse.BinaryExpressionNode{}, attrs{
					"Left": _literal(parse.IntegerNode{}, int64(2)),
					"Right": specs{parse.BinaryExpressionNode{}, attrs{
						"Left":     _literal(parse.IntegerNode{}, int64(3)),
						"Right":    _literal(parse.IntegerNode{}, int64(4)),
						"Operator": _binOp("+"),
					}},
					"Operator": _binOp("*"),
				}},
				"Right":    _literal(parse.IntegerNode{}, int64(2)),
				"Operator": _binOp("%"),
			}},
			"Right": specs{parse.BinaryExpressionNode{}, attrs{
				"Left":     _literal(parse.IntegerNode{}, int64(1)),
				"Right":    _literal(parse.IntegerNode{}, int64(2)),
				"Operator": _binOp("-"),
			}},
			"Operator": _binOp("+"),
		}},
	}}},
	{"variable", "{{ a_var }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.NameNode{}, attrs{
			"Name": _token("a_var"),
		}},
	}}},
	{"variable attribute", "{{ a_var.attr }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.GetItemNode{}, attrs{
			"Node": specs{parse.NameNode{}, attrs{
				"Name": _token("a_var"),
			}},
			"Arg": val{"attr"},
		}},
	}}},
	{"variable and filter", "{{ a_var|safe }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.FilteredExpression{}, attrs{
			"Expression": specs{parse.NameNode{}, attrs{
				"Name": _token("a_var"),
			}},
			"Filters": slice{
				filter{"safe", slice{}, attrs{}},
			},
		}},
	}}},
	{"integer and filter", "{{ 42|safe }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.FilteredExpression{}, attrs{
			"Expression": _literal(parse.IntegerNode{}, int64(42)),
			"Filters": slice{
				filter{"safe", slice{}, attrs{}},
			},
		}},
	}}},
	{"negative integer and filter", "{{ -42|safe }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.FilteredExpression{}, attrs{
			"Expression": specs{parse.UnaryExpressionNode{}, attrs{
				"Negative": val{true},
				"Term":     _literal(parse.IntegerNode{}, int64(42)),
			}},
			"Filters": slice{
				filter{"safe", slice{}, attrs{}},
			},
		}},
	}}},
	{"logical expressions", "{{ true and false }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left":     _literal(parse.BoolNode{}, true),
			"Right":    _literal(parse.BoolNode{}, false),
			"Operator": _binOp("and"),
		}},
	}}},
	{"negated boolean", "{{ not true }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.NegationNode{}, attrs{
			"Term": _literal(parse.BoolNode{}, true),
		}},
	}}},
	{"negated logical expression", "{{ not false and true }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left": specs{parse.NegationNode{}, attrs{
				"Term": _literal(parse.BoolNode{}, false),
			}},
			"Right":    _literal(parse.BoolNode{}, true),
			"Operator": _binOp("and"),
		}},
	}}},
	{"negated logical expression with parenthesis", "{{ not (false and true) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.NegationNode{}, attrs{
			"Term": specs{parse.BinaryExpressionNode{}, attrs{
				"Left":     _literal(parse.BoolNode{}, false),
				"Right":    _literal(parse.BoolNode{}, true),
				"Operator": _binOp("and"),
			}},
		}},
	}}},
	{"logical expression with math comparison", "{{ 40 + 2 > 5 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left": specs{parse.BinaryExpressionNode{}, attrs{
				"Left":     _literal(parse.IntegerNode{}, int64(40)),
				"Right":    _literal(parse.IntegerNode{}, int64(2)),
				"Operator": _binOp("+"),
			}},
			"Right":    _literal(parse.IntegerNode{}, int64(5)),
			"Operator": _binOp(">"),
		}},
	}}},
	{"logical expression with filter", "{{ false and true|safe }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.BinaryExpressionNode{}, attrs{
			"Left": _literal(parse.BoolNode{}, false),
			"Right": specs{parse.FilteredExpression{}, attrs{
				"Expression": _literal(parse.BoolNode{}, true),
				"Filters": slice{
					filter{"safe", slice{}, attrs{}},
				},
			}},
			"Operator": _binOp("and"),
		}},
	}}},
	{"logical expression with parenthesis and filter", "{{ (false and true)|safe }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.FilteredExpression{}, attrs{
			"Expression": specs{parse.BinaryExpressionNode{}, attrs{
				"Left":     _literal(parse.BoolNode{}, false),
				"Right":    _literal(parse.BoolNode{}, true),
				"Operator": _binOp("and"),
			}},
			"Filters": slice{
				filter{"safe", slice{}, attrs{}},
			},
		}},
	}}},
	{"function", "{{ a_func(42) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.CallNode{}, attrs{
			"Func": specs{parse.NameNode{}, attrs{"Name": _token("a_func")}},
			"Args": slice{_literal(parse.IntegerNode{}, int64(42))},
		}},
	}}},
	{"method", "{{ an_obj.a_method(42) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.CallNode{}, attrs{
			"Func": specs{parse.GetItemNode{}, attrs{
				"Node": specs{parse.NameNode{}, attrs{"Name": _token("an_obj")}},
				"Arg":  val{"a_method"},
			}},
			"Args": slice{_literal(parse.IntegerNode{}, int64(42))},
		}},
	}}},
	{"function with filtered args", "{{ a_func(42|safe) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.CallNode{}, attrs{
			"Func": specs{parse.NameNode{}, attrs{"Name": _token("a_func")}},
			"Args": slice{
				specs{parse.FilteredExpression{}, attrs{
					"Expression": _literal(parse.IntegerNode{}, int64(42)),
					"Filters": slice{
						filter{"safe", slice{}, attrs{}},
					},
				}},
			},
		}},
	}}},
	{"variable and multiple filters", "{{ a_var|add(42)|safe }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.FilteredExpression{}, attrs{
			"Expression": specs{parse.NameNode{}, attrs{"Name": _token("a_var")}},
			"Filters": slice{
				filter{"add", slice{_literal(parse.IntegerNode{}, int64(42))}, attrs{}},
				filter{"safe", slice{}, attrs{}},
			},
		}},
	}}},
	{"variable and expression filters", "{{ a_var|add(40 + 2) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.FilteredExpression{}, attrs{
			"Expression": specs{parse.NameNode{}, attrs{"Name": _token("a_var")}},
			"Filters": slice{
				filter{"add", slice{
					specs{parse.BinaryExpressionNode{}, attrs{
						"Left":     _literal(parse.IntegerNode{}, int64(40)),
						"Right":    _literal(parse.IntegerNode{}, int64(2)),
						"Operator": _binOp("+"),
					}},
				}, attrs{}},
			},
		}},
	}}},
	{"variable and nested filters", "{{ a_var|add( 42|add(2) ) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.FilteredExpression{}, attrs{
			"Expression": specs{parse.NameNode{}, attrs{"Name": _token("a_var")}},
			"Filters": slice{
				filter{"add", slice{
					specs{parse.FilteredExpression{}, attrs{
						"Expression": _literal(parse.IntegerNode{}, int64(42)),
						"Filters": slice{
							filter{"add", slice{_literal(parse.IntegerNode{}, int64(2))}, attrs{}},
						},
					}},
				}, attrs{}},
			},
		}},
	}}},
	{"Test equal", "{{ 3 is equal 3 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.TestExpression{}, attrs{
			"Expression": _literal(parse.IntegerNode{}, int64(3)),
			"Test": specs{parse.TestCall{}, attrs{
				"Name": val{"equal"},
				"Args": slice{_literal(parse.IntegerNode{}, int64(3))},
			}},
		}},
	}}},
	{"Test equal parenthesis", "{{ 3 is equal(3) }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.TestExpression{}, attrs{
			"Expression": _literal(parse.IntegerNode{}, int64(3)),
			"Test": specs{parse.TestCall{}, attrs{
				"Name": val{"equal"},
				"Args": slice{_literal(parse.IntegerNode{}, int64(3))},
			}},
		}},
	}}},
	{"Test ==", "{{ 3 is == 3 }}", specs{parse.OutputNode{}, attrs{
		"Expression": specs{parse.TestExpression{}, attrs{
			"Expression": _literal(parse.IntegerNode{}, int64(3)),
			"Test": specs{parse.TestCall{}, attrs{
				"Name": val{"=="},
				"Args": slice{_literal(parse.IntegerNode{}, int64(3))},
			}},
		}},
	}}},
}

// func parseText(text string) (*nodeDocument, *Error) {
// 	tokens, err := lex("test", text)
// 	if err != nil {
// 		return nil, err
// 	}
// 	parser := newParser("test", tokens, &Template{
// 		set: &TemplateSet{},
// 	})
// 	return parser.parseDocument()
// }

func _deref(value reflect.Value) reflect.Value {
	for (value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr) && !value.IsNil() {
		value = value.Elem()
	}
	return value
}

type asserter interface {
	assert(t *testing.T, value reflect.Value)
}

type specs struct {
	typ   any
	attrs attrs
}

func (specs specs) assert(t *testing.T, value reflect.Value) {
	assert := testutils.NewAssert(t)
	value = _deref(value)
	// t.Logf("type(expected %+v, actual %+v)", reflect.TypeOf(specs.typ), value.Type())
	if !assert.Equal(reflect.TypeOf(specs.typ), value.Type()) {
		return
	}
	if specs.attrs != nil {
		specs.attrs.assert(t, value)
	}
}

type val struct {
	value any
}

func (val val) assert(t *testing.T, value reflect.Value) {
	assert := testutils.NewAssert(t)
	value = _deref(value)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		assert.Equal(val.value, value.Int())
	case reflect.Float32, reflect.Float64:
		assert.Equal(val.value, value.Float())
	case reflect.String:
		assert.Equal(val.value, value.String())
	case reflect.Bool:
		assert.Equal(val.value, value.Bool())
	case reflect.Slice:
		current, ok := val.value.(asserter)
		if assert.True(ok) {
			current.assert(t, value)
		}
	case reflect.Map:
		assert.Len(val.value, value.Len())
		v2 := reflect.ValueOf(val.value)

		iter := value.MapRange()
		for iter.Next() {
			assert.Equal(iter.Value(), v2.MapIndex(iter.Key()))
		}
	case reflect.Func:
		assert.Equal(value, reflect.ValueOf(val.value))
	default:
		t.Logf("Unknown value kind '%s'", value.Kind())
	}
}

func _literal(typ any, value any) asserter {
	return specs{typ, attrs{
		"Val": val{value},
	}}
}

func _token(value string) asserter {
	return specs{parse.Token{}, attrs{
		"Val": val{value},
	}}
}

func _binOp(value string) asserter {
	return specs{parse.BinOperatorNode{}, attrs{
		"Token": _token(value),
	}}
}

type attrs map[string]asserter

func (attrs attrs) assert(t *testing.T, value reflect.Value) {
	assert := testutils.NewAssert(t)
	for attr, specs := range attrs {
		field := value.FieldByName(attr)
		if assert.True(field.IsValid(), fmt.Sprintf("No field named '%s' found", attr)) {
			specs.assert(t, field)
		}
	}
}

type slice []asserter

func (slice slice) assert(t *testing.T, value reflect.Value) {
	assert := testutils.NewAssert(t)
	if assert.Equal(t, reflect.Slice, value.Kind()) {
		if assert.Equal(t, len(slice), value.Len()) {
			for idx, specs := range slice {
				specs.assert(t, value.Index(idx))
			}
		}
	}
}

type filter struct {
	name   string
	args   slice
	kwargs attrs
}

func (filter filter) assert(t *testing.T, value reflect.Value) {
	value = _deref(value)
	assert := testutils.NewAssert(t)
	assert.Equal(reflect.TypeOf(parse.FilterCall{}), value.Type())
	assert.Equal(filter.name, value.FieldByName("Name").String())
	args := value.FieldByName("Args")
	kwargs := value.FieldByName("Kwargs")
	if assert.Equal(len(filter.args), args.Len()) {
		for idx, specs := range filter.args {
			specs.assert(t, args.Index(idx))
		}
	}
	if assert.Equal(len(filter.kwargs), kwargs.Len()) {
		for key, specs := range filter.kwargs {
			specs.assert(t, args.MapIndex(reflect.ValueOf(key)))
		}
	}
}

func TestParser(t *testing.T) {
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			// t.Parallel()
			assert := testutils.NewAssert(t)
			tpl, err := parse.Parse(test.text)
			if err != nil {
				fmt.Println(err.Error(), test.text)
			}

			if assert.Nil(err, "unable to parse template: %s", err) {
				if assert.Equal(1, len(tpl.Nodes), "Expected one node") {
					test.expected.assert(t, reflect.ValueOf(tpl.Nodes[0]))
				} else {
					t.Logf("Nodes %+v", tpl.Nodes)
				}
			}
		})
	}
}
