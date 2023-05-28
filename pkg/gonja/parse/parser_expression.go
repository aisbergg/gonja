package parse

import (
	debug "github.com/aisbergg/gonja/internal/debug/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// ParseFilterExpression parses an optional filter expression.
func (p *Parser) ParseFilterExpression(expr Expression) Expression {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

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

	debug.Print("parsed expression: %s", expr)
	return expr
}

// ParseExpression parses an expression.
func (p *Parser) ParseExpression() Expression {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	expr := p.parseLogicalExpression()
	expr = p.ParseFilterExpression(expr)

	debug.Print("parsed expression: %s", expr)
	return expr
}

// ParseExpression parses an expression.
func (p *Parser) ParseOptionalExpression() Expression {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(errors.TemplateSyntaxError); ok {
				return
			} else {
				panic(r)
			}
		}
	}()

	return p.ParseExpression()
}

// ParseExpressionNode parses an expression node.
func (p *Parser) ParseExpressionNode() Node {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	tok := p.Match(TokenVariableBegin)
	if tok == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected '%s' , expected '{{'", p.Current().Val)
	}

	node := &OutputNode{
		Start: tok,
		Trim: &Trim{
			Left: tok.Val[len(tok.Val)-1] == '-',
		},
	}

	expr := p.ParseExpression()
	if expr == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "expected an expression")
	}
	expr = p.ParseInlineIf(expr)
	node.Expression = expr

	tok = p.Match(TokenVariableEnd)
	if tok == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected '%s' , expected '}}'", p.Current().Val)
	}
	node.End = tok
	node.Trim.Right = tok.Val[0] == '-'

	debug.Print("parsed expression: %s", expr)
	return node
}

// ParseInlineIf parses an inline if-statement.
func (p *Parser) ParseInlineIf(expr Expression) Expression {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	trueExpr := expr
	if p.PeekName("if") != nil {
		tok := p.Pop()

		// parse condition
		var condTok []*Token
		for !p.Stream.End() && p.Peek(TokenVariableEnd) == nil && p.PeekName("else") == nil {
			condTok = append(condTok, p.Next())
		}
		if len(condTok) == 0 {
			errors.ThrowSyntaxError(p.Current().ErrorToken(), "expected a condition for inline if")
		}
		if p.MatchName("else") == nil {
			errors.ThrowSyntaxError(p.Current().ErrorToken(), "expected 'else', got '%s'", p.Current().Val)
		}
		stream := NewStream(condTok)
		debug.Print("condition parser")
		condParser := NewParser(p.Config, stream)
		condition := condParser.ParseExpression()

		// parse false expression
		flaseExpr := p.ParseExpression()
		expr = &InlineIfExpressionNode{
			Location:  tok,
			Condition: condition,
			TrueExpr:  trueExpr,
			FalseExpr: flaseExpr,
		}
		debug.Print("created inline if node: %s", expr)
	}

	return expr
}
