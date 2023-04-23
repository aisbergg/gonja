package django

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	debug "github.com/aisbergg/gonja/internal/debug/exec"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	u "github.com/aisbergg/gonja/pkg/gonja/utils"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var Filters = exec.FilterSet{
	"escapejs":           filterEscapejs,
	"add":                filterAdd,
	"addslashes":         filterAddslashes,
	"capfirst":           filterCapfirst,
	"cut":                filterCut,
	"date":               filterDate,
	"default_if_none":    filterDefaultIfNone,
	"floatformat":        filterFloatformat,
	"get_digit":          filterGetdigit,
	"iriencode":          filterIriencode,
	"linebreaks":         filterLinebreaks,
	"linebreaksbr":       filterLinebreaksbr,
	"linenumbers":        filterLinenumbers,
	"ljust":              filterLjust,
	"make_list":          filterMakelist,
	"phone2numeric":      filterPhone2numeric,
	"pluralize":          filterPluralize,
	"rjust":              filterRjust,
	"split":              filterSplit,
	"stringformat":       filterStringformat,
	"time":               filterDate, // time uses filterDate (same golang-format,
	"truncatechars":      filterTruncatechars,
	"truncatechars_html": filterTruncatecharsHTML,
	"truncatewords":      filterTruncatewords,
	"truncatewords_html": filterTruncatewordsHTML,
	"yesno":              filterYesno,
}

func filterTruncatecharsHelper(s string, newLen int) string {
	runes := []rune(s)
	if newLen < len(runes) {
		if newLen >= 3 {
			return fmt.Sprintf("%s...", string(runes[:newLen-3]))
		}
		// Not enough space for the ellipsis
		return string(runes[:newLen])
	}
	return string(runes)
}

func filterTruncateHTMLHelper(value string, newOutput *bytes.Buffer, cond func() bool, fn func(c rune, s int, idx int) int, finalize func()) {
	vLen := len(value)
	var tagStack []string
	idx := 0

	for idx < vLen && !cond() {
		c, s := utf8.DecodeRuneInString(value[idx:])
		if c == utf8.RuneError {
			idx += s
			continue
		}

		if c == '<' {
			newOutput.WriteRune(c)
			idx += s // consume "<"

			if idx+1 < vLen {
				if value[idx] == '/' {
					// Close tag

					newOutput.WriteString("/")

					tag := ""
					idx++ // consume "/"

					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}
						tag += string(c2)
						idx += size2
					}

					if len(tagStack) > 0 {
						// Ideally, the close tag is TOP of tag stack
						// In malformed HTML, it must not be, so iterate through the stack and remove the tag
						for i := len(tagStack) - 1; i >= 0; i-- {
							if tagStack[i] == tag {
								// Found the tag
								tagStack[i] = tagStack[len(tagStack)-1]
								tagStack = tagStack[:len(tagStack)-1]
								break
							}
						}
					}

					newOutput.WriteString(tag)
					newOutput.WriteString(">")
				} else {
					// Open tag

					tag := ""

					params := false
					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						newOutput.WriteRune(c2)

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}

						if !params {
							if c2 == ' ' {
								params = true
							} else {
								tag += string(c2)
							}
						}

						idx += size2
					}

					// Add tag to stack
					tagStack = append(tagStack, tag)
				}
			}
		} else {
			idx = fn(c, s, idx)
		}
	}

	finalize()

	for i := len(tagStack) - 1; i >= 0; i-- {
		tag := tagStack[i]
		// Close everything from the regular tag stack
		newOutput.WriteString(fmt.Sprintf("</%s>", tag))
	}
}

func filterTruncatechars(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: truncatechars(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("truncatechars(num_chars)", p.Error())
	}
	debug.Print("call filter with evaluated args: truncatechars(%s)", p.String())

	s := in.String()
	newLen := p.Args[0].Integer()
	return e.ValueFactory.NewValue(filterTruncatecharsHelper(s, newLen), false)
}

func filterTruncatecharsHTML(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: truncatechars_html(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("truncatechars_html(num_chars)", p.Error())
	}
	debug.Print("call filter with evaluated args: truncatechars_html(%s)", p.String())

	value := in.String()
	newLen := u.Max(params.Args[0].Integer()-3, 0)

	newOutput := bytes.NewBuffer(nil)

	textcounter := 0

	filterTruncateHTMLHelper(value, newOutput, func() bool {
		return textcounter >= newLen
	}, func(c rune, s int, idx int) int {
		textcounter++
		newOutput.WriteRune(c)

		return idx + s
	}, func() {
		if textcounter >= newLen && textcounter < len(value) {
			newOutput.WriteString("...")
		}
	})

	return e.ValueFactory.NewValue(newOutput.String(), true)
}

func filterTruncatewords(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: truncatewords(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("truncatewords(num_chars)", p.Error())
	}
	debug.Print("call filter with evaluated args: truncatewords(%s)", p.String())

	words := strings.Fields(in.String())
	n := p.Args[0].Integer()
	if n <= 0 {
		return e.ValueFactory.NewValue("", false)
	}
	nlen := u.Min(len(words), n)
	out := make([]string, 0, nlen)
	for i := 0; i < nlen; i++ {
		out = append(out, words[i])
	}

	if n < len(words) {
		out = append(out, "...")
	}

	return e.ValueFactory.NewValue(strings.Join(out, " "), false)
}

func filterTruncatewordsHTML(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: truncatewords_html(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("truncatewords_html(num_chars)", p.Error())
	}
	debug.Print("call filter with evaluated args: truncatewords_html(%s)", p.String())

	value := in.String()
	newLen := u.Max(p.Args[0].Integer(), 0)

	newOutput := bytes.NewBuffer(nil)

	wordcounter := 0

	filterTruncateHTMLHelper(value, newOutput, func() bool {
		return wordcounter >= newLen
	}, func(_ rune, _ int, idx int) int {
		// Get next word
		wordFound := false

		for idx < len(value) {
			c2, size2 := utf8.DecodeRuneInString(value[idx:])
			if c2 == utf8.RuneError {
				idx += size2
				continue
			}

			if c2 == '<' {
				// HTML tag start, don't consume it
				return idx
			}

			newOutput.WriteRune(c2)
			idx += size2

			if c2 == ' ' || c2 == '.' || c2 == ',' || c2 == ';' {
				// Word ends here, stop capturing it now
				break
			} else {
				wordFound = true
			}
		}

		if wordFound {
			wordcounter++
		}

		return idx
	}, func() {
		if wordcounter >= newLen {
			newOutput.WriteString("...")
		}
	})

	return e.ValueFactory.NewValue(newOutput.String(), true)
}

func filterEscapejs(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: escapejs(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("escapejs()", p.Error())
	}
	return e.ValueFactory.NewValue(template.JSEscapeString(in.String()), false)
}

func filterAdd(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: add(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("add(value)", p.Error())
	}
	debug.Print("call filter with evaluated args: add(%s)", p.String())

	param := p.Args[0]
	if in.IsNumber() && param.IsNumber() {
		if in.IsFloat() || param.IsFloat() {
			return e.ValueFactory.NewValue(in.Float()+param.Float(), false)
		}
		return e.ValueFactory.NewValue(in.Integer()+param.Integer(), false)
	}
	// If in/param is not a number, we're relying on the
	// Value's String() conversion and just add them both together
	return e.ValueFactory.NewValue(in.String()+param.String(), false)
}

func filterAddslashes(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: addslashes(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("addslashes()", p.Error())
	}
	output := strings.Replace(in.String(), "\\", "\\\\", -1)
	output = strings.Replace(output, "\"", "\\\"", -1)
	output = strings.Replace(output, "'", "\\'", -1)
	return e.ValueFactory.NewValue(output, false)
}

func filterCut(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: cut(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("cut(pattern)", p.Error())
	}
	debug.Print("call filter with evaluated args: cut(%s)", p.String())

	return e.ValueFactory.NewValue(strings.Replace(in.String(), params.Args[0].String(), "", -1), false)
}

func filterDefaultIfNone(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: default_if_none(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("default_if_none(default_value)", p.Error())
	}
	debug.Print("call filter with evaluated args: default_if_none(%s)", p.String())

	if in.IsNil() {
		return p.Args[0]
	}
	return in
}

func filterFloatformat(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: floatformat(%s)", params.String())
	p := params.ExpectKwArgs([]*exec.Kwarg{{"format", nil}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("floatformat(format=nil)", p.Error())
	}
	debug.Print("call filter with evaluated args: floatformat(%s)", p.String())

	val := in.Float()
	format := p.GetKwarg("format")

	decimals := -1
	if !format.IsNil() {
		// Any argument provided?
		decimals = format.Integer()
	}

	// if the argument is not a number (e. g. empty), the default
	// behavior is trim the result
	trim := !format.IsNumber()

	if decimals <= 0 {
		// argument is negative or zero, so we
		// want the output being trimmed
		decimals = -decimals
		trim = true
	}

	if trim {
		// Remove zeroes
		if float64(int(val)) == val {
			return e.ValueFactory.NewValue(in.Integer(), false)
		}
	}

	return e.ValueFactory.NewValue(strconv.FormatFloat(val, 'f', decimals, 64), false)
}

func filterGetdigit(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: get_digit(%s)", params.String())
	p := params.ExpectKwArgs([]*exec.Kwarg{
		{"digit", nil},
	})
	if p.IsError() {
		errors.ThrowFilterArgumentError("get_digit(digit=0)", p.Error())
	}
	debug.Print("call filter with evaluated args: get_digit(%s)", p.String())

	if !in.IsNumber() {
		errors.ThrowFilterArgumentError("get_digit(digit=0)", "get_digit(digit=0) requires a number as input")
	}
	i := p.GetKwarg("digit").Integer()
	s := in.String()
	l := len(s)
	if i <= 0 || i > l {
		return in
	}
	n, _ := strconv.Atoi(s[i : i+1])
	return e.ValueFactory.NewValue(n, false)
}

func filterIriencode(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: iriencode(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("iriencode()", p.Error())
	}

	return e.ValueFactory.NewValue(u.IRIEncode(in.String()), false)
}

func filterMakelist(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: make_list(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("make_list()", p.Error())
	}

	s := in.String()
	result := make([]string, 0, len(s))
	for _, c := range s {
		result = append(result, string(c))
	}
	return e.ValueFactory.NewValue(result, false)
}

func filterCapfirst(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: capfirst(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("capfirst()", p.Error())
	}

	if in.Len() <= 0 {
		return e.ValueFactory.NewValue("", false)
	}
	t := in.String()
	r, size := utf8.DecodeRuneInString(t)
	return e.ValueFactory.NewValue(strings.ToUpper(string(r))+t[size:], false)
}

func filterDate(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: date(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("date(format)", p.Error())
	}
	debug.Print("call filter with evaluated args: date(%s)", p.String())

	t, isTime := in.Interface().(time.Time)
	if !isTime {
		errors.ThrowFilterArgumentError("date", "filter input argument must be of type 'time.Time'")
	}
	return e.ValueFactory.NewValue(t.Format(p.Args[0].String()), false)
}

func filterLinebreaks(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: linebreaks(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("linebreaks()", p.Error())
	}

	if in.Len() == 0 {
		return in
	}

	var b bytes.Buffer

	// Newline = <br />
	// Double newline = <p>...</p>
	lines := strings.Split(in.String(), "\n")
	lenlines := len(lines)

	opened := false

	for idx, line := range lines {

		if !opened {
			b.WriteString("<p>")
			opened = true
		}

		b.WriteString(line)

		if idx < lenlines-1 && strings.TrimSpace(lines[idx]) != "" {
			// We've not reached the end
			if strings.TrimSpace(lines[idx+1]) == "" {
				// Next line is empty
				if opened {
					b.WriteString("</p>")
					opened = false
				}
			} else {
				b.WriteString("<br />")
			}
		}
	}

	if opened {
		b.WriteString("</p>")
	}

	return e.ValueFactory.NewValue(b.String(), false)
}

func filterSplit(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: split(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("split(pattern)", p.Error())
	}
	debug.Print("call filter with evaluated args: split(%s)", p.String())

	chunks := strings.Split(in.String(), params.Args[0].String())
	return e.ValueFactory.NewValue(chunks, false)
}

func filterLinebreaksbr(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: linebreaksbr(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("linebreaksbr()", p.Error())
	}

	return e.ValueFactory.NewValue(strings.Replace(in.String(), "\n", "<br />", -1), false)
}

func filterLinenumbers(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: linenumbers(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("linenumbers()", p.Error())
	}

	lines := strings.Split(in.String(), "\n")
	output := make([]string, 0, len(lines))
	for idx, line := range lines {
		output = append(output, fmt.Sprintf("%d. %s", idx+1, line))
	}
	return e.ValueFactory.NewValue(strings.Join(output, "\n"), false)
}

func filterLjust(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: split(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("split(pattern)", p.Error())
	}
	debug.Print("call filter with evaluated args: split(%s)", p.String())

	times := params.Args[0].Integer() - in.Len()
	if times < 0 {
		times = 0
	}
	return e.ValueFactory.NewValue(fmt.Sprintf("%s%s", in.String(), strings.Repeat(" ", times)), false)
}

func filterStringformat(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: stringformat(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("stringformat(format)", p.Error())
	}
	debug.Print("call filter with evaluated args: stringformat(%s)", p.String())

	return e.ValueFactory.NewValue(fmt.Sprintf(p.Args[0].String(), in.Interface()), false)
}

// https://en.wikipedia.org/wiki/Phoneword
var filterPhone2numericMap = map[string]string{
	"a": "2", "b": "2", "c": "2", "d": "3", "e": "3", "f": "3", "g": "4", "h": "4", "i": "4", "j": "5", "k": "5",
	"l": "5", "m": "6", "n": "6", "o": "6", "p": "7", "q": "7", "r": "7", "s": "7", "t": "8", "u": "8", "v": "8",
	"w": "9", "x": "9", "y": "9", "z": "9",
}

func filterPhone2numeric(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: phone2numeric(%s)", params.String())
	if p := params.ExpectNothing(); p.IsError() {
		errors.ThrowFilterArgumentError("phone2numeric()", p.Error())
	}

	sin := in.String()
	for k, v := range filterPhone2numericMap {
		sin = strings.Replace(sin, k, v, -1)
		sin = strings.Replace(sin, strings.ToUpper(k), v, -1)
	}
	return e.ValueFactory.NewValue(sin, false)
}

func filterPluralize(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: pluralize(%s)", params.String())
	p := params.ExpectKwArgs([]*exec.Kwarg{{"suffix", "s"}})
	if p.IsError() {
		errors.ThrowFilterArgumentError("pluralize(suffix='s')", p.Error())
	}
	debug.Print("call filter with evaluated args: pluralize(%s)", p.String())

	if !in.IsNumber() {
		errors.ThrowFilterArgumentError("pluralize(suffix='s')", "only works on numbers")
	}

	suffix := p.GetKwarg("suffix").String()
	endings := strings.Split(suffix, ",")
	if len(endings) > 2 {
		errors.ThrowFilterArgumentError("pluralize(suffix='s')", "only 2 endings are allowed, one for singular and one for plural")
	}
	if len(endings) == 2 {
		if in.Integer() != 1 {
			// ending for plural
			return e.ValueFactory.NewValue(endings[1], false)
		}
		return e.ValueFactory.NewValue(endings[0], false)
	}

	// only plural ending is given
	if in.Integer() != 1 {
		return e.ValueFactory.NewValue(endings[0], false)
	}
	return e.ValueFactory.NewValue("", false)
}

func filterRjust(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: rjust(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("rjust(format)", p.Error())
	}
	debug.Print("call filter with evaluated args: rjust(%s)", p.String())

	return e.ValueFactory.NewValue(fmt.Sprintf(fmt.Sprintf("%%%ds", p.Args[0].Integer()), in.String()), false)
}

func filterYesno(e *exec.Evaluator, in exec.Value, params *exec.VarArgs) exec.Value {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("call filter with raw args: yesno(%s)", params.String())
	p := params.ExpectArgs(1)
	if p.IsError() {
		errors.ThrowFilterArgumentError("yesno(arg)", p.Error())
	}
	debug.Print("call filter with evaluated args: yesno(%s)", p.String())

	choices := map[int]string{
		0: "yes",
		1: "no",
		2: "maybe",
	}
	param := p.Args[0]
	paramString := param.String()
	customChoices := strings.Split(paramString, ",")
	if len(paramString) > 0 {
		if len(customChoices) > 3 {
			errors.ThrowFilterArgumentError("yesno(arg)", "accepts only 3 options, got: '%s'.", paramString)
		}
		if len(customChoices) < 2 {
			errors.ThrowFilterArgumentError("yesno(arg)", "accepts no or at least 2 options, got: '%s'.", paramString)
		}

		// Map to the options now
		choices[0] = customChoices[0]
		choices[1] = customChoices[1]
		if len(customChoices) == 3 {
			choices[2] = customChoices[2]
		}
	}

	// maybe
	if in.IsNil() {
		return e.ValueFactory.NewValue(choices[2], false)
	}

	// yes
	if in.IsTrue() {
		return e.ValueFactory.NewValue(choices[0], false)
	}

	// no
	return e.ValueFactory.NewValue(choices[1], false)
}
