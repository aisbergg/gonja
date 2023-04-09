package exec

import (
	"fmt"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// TrimState stores and applies the trim policy.
type TrimState struct {
	Should      bool
	ShouldBlock bool
	Buffer      *strings.Builder
}

func (ts *TrimState) TrimBlocks(r rune) bool {
	if ts.ShouldBlock {
		switch r {
		case '\n':
			ts.ShouldBlock = false
			return true
		case ' ', '\t':
			return true
		default:
			return false
		}
	}
	return false
}

// Renderer is a node visitor in charge of rendering a template.
type Renderer struct {
	*EvalConfig
	Ctx      *Context
	Resolver *Resolver
	Template *Template
	Root     *parse.TemplateNode
	Current  parse.Node
	Out      *strings.Builder
	Trim     *TrimState
}

// NewRenderer initialize a new renderer
func NewRenderer(ctx *Context, resolver *Resolver, out *strings.Builder, cfg *EvalConfig, tpl *Template) *Renderer {
	var buffer strings.Builder
	r := &Renderer{
		EvalConfig: cfg,
		Ctx:        ctx,
		Resolver:   resolver,
		Template:   tpl,
		Root:       tpl.Root,
		Out:        out,
		Trim:       &TrimState{Buffer: &buffer},
	}
	r.Ctx.Set("self", Self(r))
	return r
}

// Inherit creates a new sub renderer.
func (r *Renderer) Inherit() *Renderer {
	sub := &Renderer{
		EvalConfig: r.EvalConfig.Inherit(),
		Ctx:        r.Ctx.Inherit(),
		Resolver:   r.Resolver,
		Template:   r.Template,
		Current:    r.Current,
		Root:       r.Root,
		Out:        r.Out,
		Trim:       r.Trim,
	}
	return sub
}

// Flush flushes the contents of the buffer to the final output.
func (r *Renderer) Flush(lstrip bool) {
	r.FlushAndTrim(false, lstrip)
}

// FlushAndTrim trims the contents of the buffer according to the trim policy
// and flushes it to the final output.
func (r *Renderer) FlushAndTrim(trim, lstrip bool) {
	txt := r.Trim.Buffer.String()
	if r.LstripBlocks && !lstrip {
		lines := strings.Split(txt, "\n")
		last := lines[len(lines)-1]
		lines[len(lines)-1] = strings.TrimLeft(last, " \t")
		txt = strings.Join(lines, "\n")
	}
	if trim {
		txt = strings.TrimRight(txt, " \t\n")
	}
	r.Out.WriteString(txt)
	r.Trim.Buffer.Reset()
}

// WriteString applies the trim policy on the given string and writes it to the
// buffer.
func (r *Renderer) WriteString(txt string) int {
	if r.TrimBlocks {
		txt = strings.TrimLeftFunc(txt, r.Trim.TrimBlocks)
	}
	if r.Trim.Should {
		txt = strings.TrimLeft(txt, " \t\n")
		if len(txt) > 0 {
			r.Trim.Should = false
		}
	}
	l, err := r.Trim.Buffer.WriteString(txt)
	if err != nil {
		errors.ThrowTemplateRuntimeError("unable to write to buffer: %s", err)
	}
	return l
}

// RenderValue renders a single value.
func (r *Renderer) RenderValue(value *Value) {
	if r.Autoescape && value.IsString() && !value.Safe {
		r.WriteString(value.Escaped())
	} else {
		r.WriteString(value.String())
	}
}

func (r *Renderer) StartTag(trim *parse.Trim, lstrip bool) {
	if trim == nil {
		r.Flush(lstrip)
	} else {
		r.FlushAndTrim(trim.Left, lstrip)
	}
	r.Trim.Should = false
}

func (r *Renderer) EndTag(trim *parse.Trim) {
	if trim == nil {
		return
	}
	r.Trim.Should = trim.Right
}

func (r *Renderer) Tag(trim *parse.Trim, lstrip bool) {
	r.StartTag(trim, lstrip)
	r.EndTag(trim)
}

// walk steps through the major pieces of the template structure and generates
// the output.
func (r *Renderer) walk(node parse.Node) {
	r.Current = node
	switch n := node.(type) {
	case *parse.DataNode:
		r.WriteString(n.Data.Val)

	case *parse.OutputNode:
		r.StartTag(n.Trim, false)
		value := r.Eval(n.Expression)
		r.RenderValue(value)
		r.EndTag(n.Trim)

	case *parse.StatementBlockNode:
		r.Tag(n.Trim, n.LStrip)
		r.Trim.ShouldBlock = r.TrimBlocks
		// Silently ignore non executable statements
		if stmt, ok := n.Stmt.(Statement); ok {
			stmt.Execute(r, n)
		}

	case *parse.CommentNode:
		r.Tag(n.Trim, false)

	case *parse.WrapperNode:
		for _, node := range n.Nodes {
			r.walk(node)
		}

	case *parse.TemplateNode:
		for _, node := range n.Nodes {
			r.walk(node)
		}

	default:
		panic(fmt.Errorf("BUG: cannot walk unknown node '%s'", r.Current))
	}
}

// ExecuteWrapper wraps the parse.Wrapper execution logic
func (r *Renderer) ExecuteWrapper(wrapper *parse.WrapperNode) (err error) {
	sub := r.Inherit()
	sub.Current = wrapper

	// catch all runtime errors and rethrow others
	defer func() {
		if rec := recover(); rec != nil {
			if rerr, ok := rec.(errors.TemplateRuntimeError); ok {
				rerr.Enrich(parse.AsErrorToken(sub.Current.Position()))
				err = rerr
			} else {
				panic(rec)
			}
		}
	}()
	sub.walk(wrapper)
	sub.Tag(wrapper.Trim, wrapper.LStrip)
	r.Trim.ShouldBlock = r.TrimBlocks
	return nil
}

func (r *Renderer) LStrip() {}

// Execute performs the render process by visiting every node and turning them
// into text.
func (r *Renderer) Execute() (err error) {
	// catch all runtime errors and rethrow others
	defer func() {
		if rec := recover(); rec != nil {
			if rerr, ok := rec.(errors.TemplateRuntimeError); ok {
				rerr.Enrich(parse.AsErrorToken(r.Current.Position()))
				err = rerr
			} else {
				panic(rec)
			}
		}
	}()

	// Determine the parent to be executed (for template inheritance)
	root := r.Root
	for root.Parent != nil {
		root = root.Parent
	}
	r.walk(root)
	r.Flush(false)
	return nil
}

func (r *Renderer) String() string {
	r.Flush(false)
	out := r.Out.String()
	if !r.KeepTrailingNewline {
		out = strings.TrimSuffix(out, "\n")
	}
	return out
}
