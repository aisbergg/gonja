package exec

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// TemplateLoader is an interface for loading templates by name.
type TemplateLoader interface {
	GetTemplate(string) (*Template, error)
}

// Template is the central template object. It represents a parsed template and
// is used to evaluate it.
type Template struct {
	Reader io.Reader
	Source string

	Env    *EvalConfig
	Loader TemplateLoader

	Tokens *parse.Stream
	Parser *parse.Parser

	Root   *parse.TemplateNode
	Macros MacroSet
}

// NewTemplate creates a new template.
func NewTemplate(name string, source string, cfg *EvalConfig) (*Template, error) {
	// Create the template
	t := &Template{
		Env:    cfg,
		Source: source,
		Tokens: parse.Lex(source, cfg.Config),
	}

	// Parse it
	t.Parser = parse.NewParser(cfg.Config, t.Tokens)
	t.Parser.Statements = *t.Env.Statements
	t.Parser.TemplateParser = t.Env.GetTemplate
	root, err := t.Parser.Parse()
	if err != nil {
		return nil, err
	}
	t.Root = root

	return t, nil
}

// execute executes the template with the given context and writes the rendered
// template to out.
func (tpl *Template) execute(ctx any, out io.StringWriter) (err error) {
	resolver := NewResolver(tpl.Env.Undefined, tpl.Env.CustomGetters)
	rootCtx := NewContext(tpl.Env.Globals, ctx, resolver)
	excCtx := rootCtx.Inherit()

	var builder strings.Builder
	renderer := NewRenderer(excCtx, resolver, &builder, tpl.Env, tpl)

	err = renderer.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	if _, err = out.WriteString(renderer.String()); err != nil {
		return fmt.Errorf("failed to write out template: %w", err)
	}
	return nil
}

// newBufferAndExecute executes the template with the given context and returns
// the rendered template as a newly created bytes.Buffer.
func (tpl *Template) newBufferAndExecute(ctx map[string]any) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	// Create output buffer
	// We assume that the rendered template will be 30% larger
	// buffer := bytes.NewBuffer(make([]byte, 0, int(float64(tpl.size)*1.3)))
	if err := tpl.execute(ctx, &buffer); err != nil {
		return nil, err
	}
	return &buffer, nil
}

// ExecuteBytes executes the template with the given context and returns the
// rendered template as []byte.
func (tpl *Template) ExecuteBytes(ctx map[string]any) ([]byte, error) {
	buffer, err := tpl.newBufferAndExecute(ctx)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Execute executes the template with the given context and returns the rendered
// template as a string.
func (tpl *Template) Execute(ctx any) (string, error) {
	var b strings.Builder
	err := tpl.execute(ctx, &b)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
