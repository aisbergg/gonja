package builtins

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	debug "github.com/aisbergg/gonja/internal/debug/exec"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	u "github.com/aisbergg/gonja/pkg/gonja/utils"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Filters export all builtin filters
var Filters = exec.FilterSet{
	"abs":            filterAbs,
	"attr":           filterAttr,
	"batch":          filterBatch,
	"capitalize":     filterCapitalize,
	"center":         filterCenter,
	"d":              filterDefault,
	"default":        filterDefault,
	"dictsort":       filterDictSort,
	"e":              filterEscape,
	"escape":         filterEscape,
	"filesizeformat": filterFileSize,
	"first":          filterFirst,
	"float":          filterFloat,
	"forceescape":    filterForceEscape,
	"format":         filterFormat,
	"groupby":        filterGroupBy,
	"indent":         filterIndent,
	"int":            filterInteger,
	"join":           filterJoin,
	"last":           filterLast,
	"length":         filterLength,
	"list":           filterList,
	"lower":          filterLower,
	"map":            filterMap,
	"max":            filterMax,
	"min":            filterMin,
	"pprint":         filterPPrint,
	"random":         filterRandom,
	"reject":         filterReject,
	"rejectattr":     filterRejectAttr,
	"replace":        filterReplace,
	"reverse":        filterReverse,
	"round":          filterRound,
	"safe":           filterSafe,
	"select":         filterSelect,
	"selectattr":     filterSelectAttr,
	"slice":          filterSlice,
	"sort":           filterSort,
	"string":         filterString,
	"striptags":      filterStriptags,
	"sum":            filterSum,
	"title":          filterTitle,
	"tojson":         filterToJSON,
	"trim":           filterTrim,
	"truncate":       filterTruncate,
	"unique":         filterUnique,
	"upper":          filterUpper,
	"urlencode":      filterUrlencode,
	"urlize":         filterUrlize,
	"wordcount":      filterWordcount,
	"wordwrap":       filterWordwrap,
	"xmlattr":        filterXMLAttr,
}

func filterAbs(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: abs(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("abs()", p.Error())
	}

	if in.IsInteger() {
		asInt := in.Integer()
		if asInt < 0 {
			return e.ValueFactory.NewValue(-asInt, false)
		}
		return in
	} else if in.IsFloat() {
		return e.ValueFactory.NewValue(math.Abs(in.Float()), false)
	}
	return e.ValueFactory.NewValue(math.Abs(in.Float()), false) // nothing to do here, just to keep track of the safe application
}

func filterAttr(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: attr(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("attr(key)", p.Error())
	}
	debug.Print("call filter with evaluated args: default(%s)", p.String())

	attr := p.First().String()
	value := in.GetItem(attr)
	return value
}

func filterBatch(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: batch(%s)", params.String())
	p := params.Expect(1, []*exec.Kwarg{{"fill_with", nil}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("batch(linecount, fill_with=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: batch(%s)", p.String())

	size := p.First().Integer()
	out := []exec.Value{}
	var row []exec.Value
	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		if math.Mod(float64(idx), float64(size)) == 0 {
			if row != nil {
				out = append(out, e.ValueFactory.NewValue(row, false))
			}
			row = []exec.Value{}
		}
		row = append(row, key)
		return true
	}, func() {})
	if len(row) > 0 {
		fillWith := p.GetKwarg("fill_with")
		if !fillWith.IsNil() {
			for len(row) < size {
				row = append(row, fillWith)
			}
		}
		out = append(out, e.ValueFactory.NewValue(row, false))
	}
	return e.ValueFactory.NewValue(out, false)
}

func filterCapitalize(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: capitalize(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("capitalize()", p.Error())
	}

	if in.Len() <= 0 {
		return e.ValueFactory.NewValue("", false)
	}
	t := in.String()
	r, size := utf8.DecodeRuneInString(t)
	return e.ValueFactory.NewValue(strings.ToUpper(string(r))+strings.ToLower(t[size:]), false)
}

func filterCenter(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: center(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{{"width", 80}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("center(width=80)", p.Error())
	}
	debug.Print("call filter with evaluated args: default(%s)", p.String())

	width := p.GetKwarg("width").Integer()
	slen := in.Len()
	if width <= slen {
		return in
	}

	spaces := width - slen
	left := spaces/2 + spaces%2
	right := spaces / 2

	return e.ValueFactory.NewValue(
		fmt.Sprintf("%s%s%s", strings.Repeat(" ", left), in.String(), strings.Repeat(" ", right)),
		false,
	)
}

func filterDefault(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: default(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{{"default_value", ""}, {"boolean", false}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("default(default_value='', boolean=false)", p.Error())
	}
	debug.Print("call filter with evaluated args: default(%s)", p.String())

	falsy := p.GetKwarg("boolean")
	if falsy.Bool() && !in.IsTrue() {
		return p.GetKwarg("default_value")
	} else if in.IsNil() {
		return p.GetKwarg("default_value")
	}
	return in
}

func sortByKey(in exec.Value, caseSensitive bool, reverse bool) [][2]exec.Value {
	out := [][2]exec.Value{}
	in.IterateOrder(func(idx, count int, key, value exec.Value) bool {
		out = append(out, [2]exec.Value{key, value})
		return true
	}, func() {}, reverse, true, caseSensitive)
	return out
}

func sortByValue(in exec.Value, caseSensitive, reverse bool) [][2]exec.Value {
	out := [][2]exec.Value{}
	items := in.Items()
	var sorter func(i, j int) bool
	switch {
	case caseSensitive && reverse:
		sorter = func(i, j int) bool {
			return items[i].Value.String() > items[j].Value.String()
		}
	case caseSensitive && !reverse:
		sorter = func(i, j int) bool {
			return items[i].Value.String() < items[j].Value.String()
		}
	case !caseSensitive && reverse:
		sorter = func(i, j int) bool {
			return strings.ToLower(items[i].Value.String()) > strings.ToLower(items[j].Value.String())
		}
	case !caseSensitive && !reverse:
		sorter = func(i, j int) bool {
			return strings.ToLower(items[i].Value.String()) < strings.ToLower(items[j].Value.String())
		}
	}
	sort.Slice(items, sorter)
	for _, item := range items {
		out = append(out, [2]exec.Value{item.Key, item.Value})
	}
	return out
}

func filterDictSort(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: dictsort(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"case_sensitive", false},
		{"by", "key"},
		{"reverse", false},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("dictsort(case_sensitive=false, by='key', reverse=false)", p.Error())
	}
	debug.Print("call filter with evaluated args: dictsort(%s)", p.String())

	caseSensitive := p.GetKwarg("case_sensitive").Bool()
	by := p.GetKwarg("by").String()
	reverse := p.GetKwarg("reverse").Bool()

	switch by {
	case "key":
		return e.ValueFactory.NewValue(sortByKey(in, caseSensitive, reverse), false)
	case "value":
		return e.ValueFactory.NewValue(sortByValue(in, caseSensitive, reverse), false)
	}
	errors.ThrowFilterArgumentError("dictsort(case_sensitive=false, by='key', reverse=false)", "'by' should be either 'key' or 'value'")
	return nil
}

func filterEscape(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: escape(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("escape()", p.Error())
	}

	if in.IsSafe() {
		return in
	}
	return e.ValueFactory.NewValue(in.Escaped(), true)
}

var (
	bytesPrefixes  = []string{"kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	binaryPrefixes = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
)

func filterFileSize(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: filesizeformat(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{{"binary", false}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("filesizeformat(binary=false)", p.Error())
	}
	debug.Print("call filter with evaluated args: default(%s)", p.String())

	bytes := in.Float()
	binary := p.GetKwarg("binary").Bool()
	var base float64
	var prefixes []string
	if binary {
		base = 1024.0
		prefixes = binaryPrefixes
	} else {
		base = 1000.0
		prefixes = bytesPrefixes
	}
	if bytes == 1.0 {
		return e.ValueFactory.NewValue("1 Byte", false)
	} else if bytes < base {
		return e.ValueFactory.NewValue(fmt.Sprintf("%.0f Bytes", bytes), false)
	} else {
		var i int
		var unit float64
		var prefix string
		for i, prefix = range prefixes {
			unit = math.Pow(base, float64(i+2))
			if bytes < unit {
				return e.ValueFactory.NewValue(fmt.Sprintf("%.1f %s", (base*bytes/unit), prefix), false)
			}
		}
		return e.ValueFactory.NewValue(fmt.Sprintf("%.1f %s", (base*bytes/unit), prefix), false)
	}
}

func filterFirst(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: first(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("first()", p.Error())
	}

	if in.IsSliceable() && in.Len() > 0 {
		return in.Index(0)
	}
	return e.ValueFactory.NewValue("", false)
}

func filterFloat(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: float(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("float()", p.Error())
	}

	return e.ValueFactory.NewValue(in.Float(), false)
}

func filterForceEscape(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: forceescape(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("forceescape()", p.Error())
	}

	return e.ValueFactory.NewValue(in.Escaped(), true)
}

func filterFormat(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: format(%s)", params.String())

	args := []any{}
	for _, arg := range params.Args {
		args = append(args, arg.Interface())
	}
	return e.ValueFactory.NewValue(fmt.Sprintf(in.String(), args...), false)
}

// XXX: 'default' and 'case_sensitive' need to be implemented
func filterGroupBy(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: groupby(%s)", params.String())
	p := params.Expect(1, []*exec.Kwarg{
		{"default", nil},
		{"case_sensitive", false},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("groupby(attribute, default=nil, case_sensitive=false)", p.Error())
	}
	debug.Print("call filter with evaluated args: groupby(%s)", p.String())

	field := p.First().String()
	groups := map[any][]exec.Value{}
	groupers := []any{}

	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		attr := key.GetItem(field)
		lst, exists := groups[attr.Interface()]
		if !exists {
			lst = []exec.Value{}
			groupers = append(groupers, attr.Interface())
		}
		lst = append(lst, key)
		groups[attr.Interface()] = lst
		return true
	}, func() {})

	out := []map[string]exec.Value{}
	for _, grouper := range groupers {
		out = append(out, map[string]exec.Value{
			"grouper": e.ValueFactory.NewValue(grouper, false),
			"list":    e.ValueFactory.NewValue(groups[grouper], false),
		})
	}
	return e.ValueFactory.NewValue(out, false)
}

func filterIndent(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: indent(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"width", 4},
		{"first", false},
		{"blank", false},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("indent(width=4, first=false, blank=false)", p.Error())
	}
	debug.Print("call filter with evaluated args: indent(%s)", p.String())

	width := p.GetKwarg("width").Integer()
	first := p.GetKwarg("first").Bool()
	blank := p.GetKwarg("blank").Bool()
	indent := strings.Repeat(" ", width)
	lines := strings.Split(in.String(), "\n")
	// start := 1
	// if first {start = 0}
	var out strings.Builder
	for idx, line := range lines {
		if line == "" && !blank {
			out.WriteByte('\n')
			continue
		}
		if idx > 0 || first {
			out.WriteString(indent)
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return e.ValueFactory.NewValue(out.String(), false)
}

func filterInteger(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: int(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("int()", p.Error())
	}

	return e.ValueFactory.NewValue(in.Integer(), false)
}

func filterJoin(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: join(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"d", ""},
		{"attribute", nil},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("join(d='', attribute=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: join(%s)", p.String())

	if !in.IsSliceable() {
		return in
	}
	sep := p.GetKwarg("d").String()
	sl := make([]string, 0, in.Len())
	for i := 0; i < in.Len(); i++ {
		sl = append(sl, in.Index(i).String())
	}
	return e.ValueFactory.NewValue(strings.Join(sl, sep), false)
}

func filterLast(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: last(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("last()", p.Error())
	}

	if in.IsSliceable() && in.Len() > 0 {
		return in.Index(in.Len() - 1)
	}
	return e.ValueFactory.NewValue("", false)
}

func filterLength(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: length(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("length()", p.Error())
	}

	return e.ValueFactory.NewValue(in.Len(), false)
}

func filterList(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: list(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("list()", p.Error())
	}

	if in.IsString() {
		out := []string{}
		for _, r := range in.String() {
			out = append(out, string(r))
		}
		return e.ValueFactory.NewValue(out, false)
	}
	out := []exec.Value{}
	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		out = append(out, key)
		return true
	}, func() {})
	return e.ValueFactory.NewValue(out, false)
}

func filterLower(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: lower(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("lower()", p.Error())
	}

	return e.ValueFactory.NewValue(strings.ToLower(in.String()), false)
}

func filterMap(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: map(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"filter", ""},
		{"attribute", nil},
		{"default", nil},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("map(filter='', attribute=nil, default=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: map(%s)", p.String())

	filter := p.GetKwarg("filter").String()
	attribute := p.GetKwarg("attribute").String()
	defaultVal := p.GetKwarg("default")
	out := []exec.Value{}
	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		val := key
		if len(attribute) > 0 {
			attr := val.GetItem(attribute)
			if exec.IsDefined(attr) {
				val = attr
			} else if defaultVal != nil {
				val = defaultVal
			} else {
				return true
			}
		}
		if len(filter) > 0 {
			val = e.ExecuteFilterByName(filter, val, exec.NewVarArgs(e.ValueFactory))
		}
		out = append(out, val)
		return true
	}, func() {})
	return e.ValueFactory.NewValue(out, false)
}

func filterMax(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: max(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"case_sensitive", false},
		{"attribute", nil},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("max(case_sensitive=false, attribute=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: max(%s)", p.String())

	caseSensitive := p.GetKwarg("case_sensitive").Bool()
	attribute := p.GetKwarg("attribute").String()

	var max exec.Value
	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		val := key
		if len(attribute) > 0 {
			attr := val.GetItem(attribute)
			if exec.IsDefined(attr) {
				val = attr
			} else {
				val = nil
			}
		}
		if max == nil {
			max = val
			return true
		}
		if val == nil || max == nil {
			return true
		}
		switch {
		case max.IsFloat() || max.IsInteger() && val.IsFloat() || val.IsInteger():
			if val.Float() > max.Float() {
				max = val
			}
		case max.IsString() && val.IsString():
			if !caseSensitive && strings.ToLower(val.String()) > strings.ToLower(max.String()) {
				max = val
			} else if caseSensitive && val.String() > max.String() {
				max = val
			}
		default:
			errors.ThrowFilterArgumentError("max()", "%T and %T are not comparable", max.Interface(), val.Interface())

		}
		return true
	}, func() {})

	if max == nil {
		return e.ValueFactory.NewValue("", false)
	}
	return max
}

func filterMin(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: min(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"case_sensitive", false},
		{"attribute", nil},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("min(case_sensitive=false, attribute=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: min(%s)", p.String())

	caseSensitive := p.GetKwarg("case_sensitive").Bool()
	attribute := p.GetKwarg("attribute").String()

	var min exec.Value
	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		val := key
		if len(attribute) > 0 {
			attr := val.GetItem(attribute)
			if exec.IsDefined(attr) {
				val = attr
			} else {
				val = nil
			}
		}
		if min == nil {
			min = val
			return true
		}
		if val == nil || min == nil {
			return true
		}
		switch {
		case min.IsFloat() || min.IsInteger() && val.IsFloat() || val.IsInteger():
			if val.Float() < min.Float() {
				min = val
			}
		case min.IsString() && val.IsString():
			if !caseSensitive && strings.ToLower(val.String()) < strings.ToLower(min.String()) {
				min = val
			} else if caseSensitive && val.String() < min.String() {
				min = val
			}
		default:
			errors.ThrowFilterArgumentError("min()", "%T and %T are not comparable", min.Interface(), val.Interface())
		}
		return true
	}, func() {})

	if min == nil {
		return e.ValueFactory.NewValue("", false)
	}
	return min
}

func filterPPrint(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: pprint(%s)", params.String())
	p := params.ExpectNothing()
	if p.IsError() {
		errors.ThrowFilterArgumentError("pprint()", p.Error())
	}
	debug.Print("call filter with evaluated args: pprint(%s)", p.String())

	b, err := json.MarshalIndent(in.Interface(), "", "  ")
	if err != nil {
		errors.ThrowFilterArgumentError("pprint()", "unable to pretty print '%s'", in.String())
	}
	return e.ValueFactory.NewValue(string(b), true)
}

func filterRandom(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: random(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("random()", p.Error())
	}

	if !in.IsSliceable() || in.Len() <= 0 {
		return in
	}
	i := rand.Intn(in.Len())
	return in.Index(i)
}

func filterReject(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: reject(%s)", params.String())

	var test func(exec.Value) bool
	if len(params.Args) == 0 {
		// Reject truthy value
		test = func(in exec.Value) bool {
			return in.IsTrue()
		}
	} else {
		name := params.First().String()
		testParams := &exec.VarArgs{
			Args:   params.Args[1:],
			Kwargs: params.Kwargs,
		}
		test = func(in exec.Value) bool {
			out := e.ExecuteTestByName(name, in, testParams)
			return out.IsTrue()
		}
	}

	out := []exec.Value{}

	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		if !test(key) {
			out = append(out, key)
		}
		return true
	}, func() {})

	return e.ValueFactory.NewValue(out, false)
}

func filterRejectAttr(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: rejectattr(%s)", params.String())
	if len(params.Args) < 1 {
		errors.ThrowFilterArgumentError("rejectattr(*args, **kwargs)", "at least one attribute needs to be specified")
	}

	var test func(exec.Value) exec.Value
	attribute := params.First().String()
	if len(params.Args) == 1 {
		// Reject truthy value
		test = func(in exec.Value) exec.Value {
			return in.GetItem(attribute)
		}
	} else {
		name := params.Args[1].String()
		testParams := &exec.VarArgs{
			Args:   params.Args[2:],
			Kwargs: params.Kwargs,
		}
		test = func(in exec.Value) exec.Value {
			attr := in.GetItem(attribute)
			out := e.ExecuteTestByName(name, attr, testParams)
			return out
		}
	}

	out := []exec.Value{}

	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		result := test(key)
		if !result.IsTrue() {
			out = append(out, key)
		}
		return true
	}, func() {})

	return e.ValueFactory.NewValue(out, false)
}

func filterReplace(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: replace(%s)", params.String())
	p := params.Expect(2, []*exec.Kwarg{{"count", nil}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("replace(old, new, count=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: replace(%s)", p.String())

	old := p.Args[0].String()
	new := p.Args[1].String()
	count := p.GetKwarg("count")
	if count.IsNil() {
		return e.ValueFactory.NewValue(strings.ReplaceAll(in.String(), old, new), false)
	}
	return e.ValueFactory.NewValue(strings.Replace(in.String(), old, new, count.Integer()), false)
}

func filterReverse(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: reverse(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("reverse()", p.Error())
	}

	if in.IsString() {
		var out strings.Builder
		in.IterateOrder(func(idx, count int, key, value exec.Value) bool {
			out.WriteString(key.String())
			return true
		}, func() {}, true, false, false)
		return e.ValueFactory.NewValue(out.String(), false)
	}
	out := []exec.Value{}
	in.IterateOrder(func(idx, count int, key, value exec.Value) bool {
		out = append(out, key)
		return true
	}, func() {}, true, true, false)
	return e.ValueFactory.NewValue(out, false)
}

func filterRound(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: round(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"precision", 0},
		{"method", "common"},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("round(precision=0, method='common')", p.Error())
	}
	debug.Print("call filter with evaluated args: round(%s)", p.String())

	method := p.GetKwarg("method").String()
	var op func(float64) float64
	switch method {
	case "common":
		op = math.Round
	case "floor":
		op = math.Floor
	case "ceil":
		op = math.Ceil
	default:
		errors.ThrowFilterArgumentError("round(precision=0, method='common')", "unknown method '%s', must be one of 'common, 'floor', 'ceil'", method)
	}
	value := in.Float()
	factor := float64(10 * p.GetKwarg("precision").Integer())
	if factor > 0 {
		value = value * factor
	}
	value = op(value)
	if factor > 0 {
		value = value / factor
	}
	return e.ValueFactory.NewValue(value, false)
}

func filterSafe(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: safe(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("safe()", p.Error())
	}

	// create a new [Value] container with the same value but safe
	return e.ValueFactory.NewValue(in.Interface(), true)
}

func filterSelect(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: select(%s)", params.String())

	var test func(exec.Value) bool
	if len(params.Args) == 0 {
		// Reject truthy value
		test = func(in exec.Value) bool {
			return in.IsTrue()
		}
	} else {
		name := params.First().String()
		testParams := &exec.VarArgs{
			Args:   params.Args[1:],
			Kwargs: params.Kwargs,
		}
		test = func(in exec.Value) bool {
			out := e.ExecuteTestByName(name, in, testParams)
			return out.IsTrue()
		}
	}

	out := []exec.Value{}

	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		if test(key) {
			out = append(out, key)
		}
		return true
	}, func() {})

	return e.ValueFactory.NewValue(out, false)
}

func filterSelectAttr(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: selectattr(%s)", params.String())
	if len(params.Args) < 1 {
		errors.ThrowFilterArgumentError("selectattr(*args, **kwargs)", "at least one attribute needs to be specified")
	}

	var test func(exec.Value) exec.Value
	attribute := params.First().String()
	if len(params.Args) == 1 {
		// Reject truthy value
		test = func(in exec.Value) exec.Value {
			return in.GetItem(attribute)
		}
	} else {
		name := params.Args[1].String()
		testParams := &exec.VarArgs{
			Args:   params.Args[2:],
			Kwargs: params.Kwargs,
		}
		test = func(in exec.Value) exec.Value {
			attr := in.GetItem(attribute)
			out := e.ExecuteTestByName(name, attr, testParams)
			return out
		}
	}

	out := []exec.Value{}

	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		result := test(key)
		if result.IsTrue() {
			out = append(out, key)
		}
		return true
	}, func() {})

	return e.ValueFactory.NewValue(out, false)
}

// XXX: make this filter behave like the python one
func filterSlice(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: slice(%s)", params.String())
	p := params.Expect(1, []*exec.Kwarg{{"fill_with", nil}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("slice(slices, fill_with=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: slice(%s)", p.String())

	// XXX: this stuff is original
	comp := strings.Split(params.Args[0].String(), ":")
	if len(comp) != 2 {
		return e.ValueFactory.NewValue(fmt.Errorf("Slice string must have the format 'from:to' [from/to can be omitted, but the ':' is required]"), false)
	}

	if !in.IsSliceable() {
		return in
	}

	from := e.ValueFactory.NewValue(comp[0], false).Integer()
	to := in.Len()

	if from > to {
		from = to
	}

	vto := e.ValueFactory.NewValue(comp[1], false).Integer()
	if vto >= from && vto <= in.Len() {
		to = vto
	}

	return in.Slice(from, to)
}

func filterSort(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: sort(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{{"reverse", false}, {"case_sensitive", false}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("sort(reverse=false, case_sensitive=false)", p.Error())
	}
	debug.Print("call filter with evaluated args: sort(%s)", p.String())

	reverse := p.GetKwarg("reverse").Bool()
	caseSensitive := p.GetKwarg("case_sensitive").Bool()
	out := []exec.Value{}
	in.IterateOrder(func(idx, count int, key, value exec.Value) bool {
		out = append(out, key)
		return true
	}, func() {}, reverse, true, caseSensitive)
	return e.ValueFactory.NewValue(out, false)
}

func filterString(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: string(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("string()", p.Error())
	}

	return e.ValueFactory.NewValue(in.String(), false)
}

var reStriptags = regexp.MustCompile(`<[^>]*?>`)

func filterStriptags(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: striptags(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("striptags()", p.Error())
	}

	s := in.String()

	// Strip all tags
	s = reStriptags.ReplaceAllString(s, "")

	return e.ValueFactory.NewValue(strings.TrimSpace(s), false)
}

func filterSum(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: sum(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{{"attribute", nil}, {"start", 0}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("sum(attribute=nil, start=0)", p.Error())
	}
	debug.Print("call filter with evaluated args: sum(%s)", p.String())

	attribute := p.GetKwarg("attribute")
	sum := p.GetKwarg("start").Float()
	var err error

	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		if attribute.IsString() {
			val := key
			for _, attr := range strings.Split(attribute.String(), ".") {
				val = val.GetItem(attr)
			}
			if val.IsNumber() {
				sum += val.Float()
			}
		} else if attribute.IsInteger() {
			value := key.GetItem(attribute.Integer())
			sum += value.Float()
		} else {
			sum += key.Float()
		}
		return true
	}, func() {})

	if err != nil {
		return e.ValueFactory.NewValue(err, false)
	} else if sum == math.Trunc(sum) {
		return e.ValueFactory.NewValue(int64(sum), false)
	}
	return e.ValueFactory.NewValue(sum, false)
}

func filterTitle(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: title(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("title()", p.Error())
	}

	if !in.IsString() {
		return e.ValueFactory.NewValue("", false)
	}
	return e.ValueFactory.NewValue(strings.Title(strings.ToLower(in.String())), false)
}

func filterTrim(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: trim(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("trim()", p.Error())
	}

	return e.ValueFactory.NewValue(strings.TrimSpace(in.String()), false)
}

func filterToJSON(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: tojson(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{{"indent", nil}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("tojson(indent=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: tojson(%s)", p.String())

	indent := p.GetKwarg("indent")
	var out string
	if indent.IsNil() {
		b, err := json.Marshal(in.Interface())
		if err != nil {
			errors.ThrowFilterArgumentError("tojson(indent=nil)", "unable to marhsall to json: %s", err.Error())
		}
		out = string(b)
	} else if indent.IsInteger() {
		b, err := json.MarshalIndent(in.Interface(), "", strings.Repeat(" ", indent.Integer()))
		if err != nil {
			errors.ThrowFilterArgumentError("tojson(indent=nil)", "unable to marhsall to json: %s", err.Error())
		}
		out = string(b)
	} else {
		errors.ThrowFilterArgumentError("tojson(indent=nil)", "expected an integer for 'indent', got '%s'", indent.String())
	}
	return e.ValueFactory.NewValue(out, true)
}

func filterTruncate(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: truncate(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"length", 255},
		{"killwords", false},
		{"end", "..."},
		{"leeway", 0},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("truncate(length=255, killwords=false, end='...', leeway=0)", p.Error())
	}
	debug.Print("call filter with evaluated args: truncate(%s)", p.String())

	source := in.String()
	length := p.GetKwarg("length").Integer()
	leeway := p.GetKwarg("leeway").Integer()
	killwords := p.GetKwarg("killwords").Bool()
	end := p.GetKwarg("end").String()
	rEnd := []rune(end)
	fullLength := length + leeway
	runes := []rune(source)

	if length < len(rEnd) {
		errors.ThrowFilterArgumentError("truncate(length=255, killwords=false, end='...', leeway=0)", "expected length >= %d, got %d", len(rEnd), length)
	}

	if len(runes) <= fullLength {
		return e.ValueFactory.NewValue(source, false)
	}

	atLength := string(runes[:length-len(rEnd)])
	if !killwords {
		atLength = strings.TrimRightFunc(atLength, func(r rune) bool {
			return !unicode.IsSpace(r)
		})
		atLength = strings.TrimRight(atLength, " \n\t")
	}
	return e.ValueFactory.NewValue(fmt.Sprintf("%s%s", atLength, end), false)
}

func filterUnique(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: unique(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{{"case_sensitive", false}, {"attribute", nil}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("unique(case_sensitive=false, attribute=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: unique(%s)", p.String())

	caseSensitive := p.GetKwarg("case_sensitive").Bool()
	attribute := p.GetKwarg("attribute")

	out := exec.ValuesList{}
	tracker := map[any]bool{}
	var err error

	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		val := key
		if attribute.IsString() {
			attr := attribute.String()
			val = key.GetItem(attr)
		}
		tracked := val.Interface()
		if !caseSensitive && val.IsString() {
			tracked = strings.ToLower(val.String())
		}
		if _, contains := tracker[tracked]; !contains {
			tracker[tracked] = true
			out = append(out, key)
		}
		return true
	}, func() {})

	if err != nil {
		return e.ValueFactory.NewValue(err, false)
	}
	return e.ValueFactory.NewValue(out, false)
}

func filterUpper(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: upper(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("upper()", p.Error())
	}

	return e.ValueFactory.NewValue(strings.ToUpper(in.String()), false)
}

func filterUrlencode(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: urlencode(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("urlencode()", p.Error())
	}

	return e.ValueFactory.NewValue(url.QueryEscape(in.String()), false)
}

// TODO: This regexp could do some work
var filterUrlizeURLRegexp = regexp.MustCompile(`((((http|https)://)|www\.|((^|[ ])[0-9A-Za-z_\-]+(\.com|\.net|\.org|\.info|\.biz|\.de))))(?U:.*)([ ]+|$)`)
var filterUrlizeEmailRegexp = regexp.MustCompile(`(\w+@\w+\.\w{2,4})`)

func filterUrlizeHelper(input string, trunc int, rel string, target string) (string, error) {
	var soutErr error
	sout := filterUrlizeURLRegexp.ReplaceAllStringFunc(input, func(raw_url string) string {
		var prefix string
		var suffix string
		if strings.HasPrefix(raw_url, " ") {
			prefix = " "
		}
		if strings.HasSuffix(raw_url, " ") {
			suffix = " "
		}

		raw_url = strings.TrimSpace(raw_url)

		url := u.IRIEncode(raw_url)

		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("http://%s", url)
		}

		title := raw_url

		if trunc > 3 && len(title) > trunc {
			title = fmt.Sprintf("%s...", title[:trunc-3])
		}

		title = u.HTMLEscape(title)

		attrs := ""
		if len(target) > 0 {
			attrs = fmt.Sprintf(` target="%s"`, target)
		}

		rels := []string{}
		cleanedRel := strings.Trim(strings.Replace(rel, "noopener", "", -1), " ")
		if len(cleanedRel) > 0 {
			rels = append(rels, cleanedRel)
		}
		rels = append(rels, "noopener")
		rel = strings.Join(rels, " ")

		return fmt.Sprintf(`%s<a href="%s" rel="%s"%s>%s</a>%s`, prefix, url, rel, attrs, title, suffix)
	})
	if soutErr != nil {
		return "", soutErr
	}

	sout = filterUrlizeEmailRegexp.ReplaceAllStringFunc(sout, func(mail string) string {
		title := mail

		if trunc > 3 && len(title) > trunc {
			title = fmt.Sprintf("%s...", title[:trunc-3])
		}

		return fmt.Sprintf(`<a href="mailto:%s">%s</a>`, mail, title)
	})
	return sout, nil
}

func filterUrlize(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: urlize(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"trim_url_limit", nil},
		{"nofollow", false},
		{"target", nil},
		{"rel", nil},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("urlize(trim_url_limit=nil, nofollow=false, target=nil, rel=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: urlize(%s)", p.String())

	truncate := -1
	if param := p.GetKwarg("trim_url_limit"); param.IsInteger() {
		truncate = param.Integer()
	}
	rel := p.GetKwarg("rel")
	target := p.GetKwarg("target")

	s, err := filterUrlizeHelper(in.String(), truncate, rel.String(), target.String())
	if err != nil {
		return e.ValueFactory.NewValue(err, false)
	}

	return e.ValueFactory.NewValue(s, false)
}

func filterWordcount(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: wordcount(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("wordcount()", p.Error())
	}

	return e.ValueFactory.NewValue(len(strings.Fields(in.String())), false)
}

func filterWordwrap(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: wordwrap(%s)", params.String())
	p := params.Expect(0, []*exec.Kwarg{
		{"width", 79},
		{"break_long_words", true},
		{"wrapstring", true},
		{"break_on_hyphens", true},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("wordwrap(width=79, break_long_words=True, wrapstring=True, break_on_hyphens=True)", p.Error())
	}
	debug.Print("call filter with evaluated args: wordwrap(%s)", p.String())

	words := strings.Fields(in.String())
	wordsLen := len(words)
	wrapAt := params.Args[0].Integer()
	if wrapAt <= 0 {
		return in
	}

	linecount := wordsLen/wrapAt + wordsLen%wrapAt
	lines := make([]string, 0, linecount)
	for i := 0; i < linecount; i++ {
		lines = append(lines, strings.Join(words[wrapAt*i:u.Min(wrapAt*(i+1), wordsLen)], " "))
	}
	return e.ValueFactory.NewValue(strings.Join(lines, "\n"), false)
}

func filterXMLAttr(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: xmlattr(%s)", params.String())
	p := params.ExpectKwArgs([]*exec.Kwarg{{"autospace", true}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("xmlattr(autoescape=true)", p.Error())
	}
	debug.Print("call filter with evaluated args: xmlattr(%s)", p.String())

	autospace := p.GetKwarg("autospace").Bool()
	kvs := []string{}
	in.Iterate(func(idx, count int, key, value exec.Value) bool {
		if !value.IsTrue() {
			return true
		}
		kv := fmt.Sprintf(`%s="%s"`, key.Escaped(), value.Escaped())
		kvs = append(kvs, kv)
		return true
	}, func() {})
	out := strings.Join(kvs, " ")
	if autospace {
		out = " " + out
	}
	return e.ValueFactory.NewValue(out, false)
}
