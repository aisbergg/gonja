package parse

import (
	debug "github.com/aisbergg/gonja/internal/debug/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// TemplateParseFn is a function that parses a template string and returns a
// node tree.
type TemplateParseFn func(string) (*TemplateNode, error)

// Doc = { ( Filter | Tag | HTML ) }
func (p *Parser) parseDocElement() Node {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	t := p.Current()

	switch t.Type {
	case TokenData:
		n := &DataNode{Data: t}
		p.Consume() // consume HTML element
		return n
	case TokenEOF:
		p.Consume()
		return nil
	case TokenCommentBegin:
		return p.ParseComment()
	case TokenVariableBegin:
		return p.ParseExpressionNode()
	case TokenBlockBegin:
		return p.ParseStatementBlock()
	}
	errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected token (only HTML/tags/filters in templates allowed)")
	return nil
}

// ParseTemplate parses a template and returns the root node of the AST.
func (p *Parser) ParseTemplate() (tpl *TemplateNode, err error) {
	// catch all syntax errors and rethrow others
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(errors.TemplateSyntaxError); ok {
				err = rerr
			} else {
				panic(r)
			}
		}
	}()

	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse template: %s", p.Current())

	tpl = &TemplateNode{
		Blocks: BlockSet{},
		Macros: map[string]*MacroNode{},
	}
	p.Template = tpl

	for !p.Stream.End() {
		node := p.parseDocElement()
		if node != nil {
			tpl.Nodes = append(tpl.Nodes, node)
		}
	}
	return tpl, nil
}
