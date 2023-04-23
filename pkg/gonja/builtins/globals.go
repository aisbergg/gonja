package builtins

import (
	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/utils"
)

var Globals = map[string]any{
	"cycler":    Cycler,
	"dict":      Dict,
	"joiner":    Joiner,
	"lipsum":    Lipsum,
	"namespace": Namespace,
	"range":     Range,
}

func Range(va *exec.VarArgs) <-chan int {
	var (
		start = 0
		stop  = -1
		step  = 1
	)
	switch len(va.Args) {
	case 1:
		stop = va.Args[0].Integer()
	case 2:
		start = va.Args[0].Integer()
		stop = va.Args[1].Integer()
	case 3:
		start = va.Args[0].Integer()
		stop = va.Args[1].Integer()
		step = va.Args[2].Integer()
		// default:
		// 	return nil, errors.New("range expect signature range([start, ]stop[, step])")
	}
	chnl := make(chan int)
	go func() {
		for i := start; i < stop; i += step {
			chnl <- i
		}

		// Ensure that at the end of the loop we close the channel!
		close(chnl)
	}()
	return chnl
}

func Dict(va *exec.VarArgs) exec.Value {
	dict := exec.NewDict()
	for _, kv := range va.Kwargs {
		dict.Pairs = append(dict.Pairs, &exec.Pair{
			Key:   va.ValueFactory.NewValue(kv.Key, false),
			Value: kv.Value,
		})
	}
	return va.ValueFactory.NewValue(dict, false)
}

type cycler struct {
	values  []string
	idx     int
	getters map[string]any
}

func (c *cycler) Reset() {
	c.idx = 0
	c.getters["current"] = c.values[c.idx]
}

func (c *cycler) Next() string {
	c.idx++
	value := c.getters["current"].(string)
	if c.idx >= len(c.values) {
		c.idx = 0
	}
	c.getters["current"] = c.values[c.idx]
	return value
}

func Cycler(va *exec.VarArgs) exec.Value {
	c := &cycler{}
	for _, arg := range va.Args {
		c.values = append(c.values, arg.String())
	}
	c.getters = map[string]any{
		"next":  c.Next,
		"reset": c.Reset,
	}
	c.Reset()
	return va.ValueFactory.NewValue(c.getters, false)
}

type joiner struct {
	sep   string
	first bool
}

func (j *joiner) String() string {
	if !j.first {
		j.first = true
		return ""
	}
	return j.sep
}

func Joiner(va *exec.VarArgs) exec.Value {
	p := va.ExpectKwArgs([]*exec.Kwarg{{"sep", ","}})
	if p.IsError() {
		errors.ThrowTemplateRuntimeError("wrong signature for 'joiner': %s", p.Error())
	}
	sep := p.GetKwarg("sep").String()
	j := &joiner{sep: sep}
	return va.ValueFactory.NewValue(j.String, false)
}

func Namespace(va *exec.VarArgs) map[string]any {
	ns := map[string]any{}
	for _, kv := range va.Kwargs {
		ns[kv.Key] = kv.Value
	}
	return ns
}

func Lipsum(va *exec.VarArgs) exec.Value {
	p := va.ExpectKwArgs([]*exec.Kwarg{
		{"n", 5},
		{"html", true},
		{"min", 20},
		{"max", 100},
	})
	if p.IsError() {
		errors.ThrowTemplateRuntimeError("wrong signature for 'lipsum': %s", p.Error())
	}
	n := p.GetKwarg("n").Integer()
	html := p.GetKwarg("html").Bool()
	min := p.GetKwarg("min").Integer()
	max := p.GetKwarg("max").Integer()
	return va.ValueFactory.NewValue(utils.Lipsum(n, html, min, max), true)
}
