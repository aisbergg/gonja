package parse

import (
	log "github.com/aisbergg/gonja/internal/log/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// ParseFilterExpression parses an optional filter expression.
func (p *Parser) ParseFilterExpression(expr Expression) Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	if p.Peek(TokenPipe) != nil {
		filtered := &FilteredExpression{
			Expression: expr,
		}
		for p.Match(TokenPipe) != nil {
			// Parse one single filter
			filter := p.ParseFilter()

			// Check sandbox filter restriction
			// if _, isBanned := p.template.set.bannedFilters[filter.name]; isBanned {
			// 	return nil, p.Error(fmt.Sprintf("Usage of filter '%s' is not allowed (sandbox restriction active).", filter.name))
			// }

			filtered.Filters = append(filtered.Filters, filter)
		}
		expr = filtered
	}

	log.Print("parsed expression: %s", expr)
	return expr
}

// ParseExpression parses an expression.
func (p *Parser) ParseExpression() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	expr := p.parseLogicalExpression()
	expr = p.ParseFilterExpression(expr)

	log.Print("parsed expression: %s", expr)
	return expr
}

// ParseExpressionNode parses an expression node.
func (p *Parser) ParseExpressionNode() Node {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	tok := p.Match(TokenVariableBegin)
	if tok == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s' , expected '{{'", p.Current())
	}

	node := &OutputNode{
		Start: tok,
		Trim: &Trim{
			Left: tok.Val[len(tok.Val)-1] == '-',
		},
	}

	expr := p.ParseExpression()
	if expr == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "expected an expression")
	}
	node.Expression = expr

	tok = p.Match(TokenVariableEnd)
	if tok == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s' , expected '}}'", p.Current())
	}
	node.End = tok
	node.Trim.Right = tok.Val[0] == '-'

	log.Print("parsed expression: %s", expr)
	return node
}
