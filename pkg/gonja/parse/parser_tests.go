package parse

import (
	debug "github.com/aisbergg/gonja/internal/debug/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

func (p *Parser) ParseTest(expr Expression) Expression {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	current := p.Current()
	line := current.Line
	_ = line
	expr = p.ParseFilterExpression(expr)

	if p.MatchName("is") != nil {
		not := p.MatchName("not")
		ident := p.Next()
		test := &TestCall{
			Token:  ident,
			Name:   ident.Val,
			Args:   []Expression{},
			Kwargs: map[string]Expression{},
		}

		if ident.Val == "in" {
			// requires an expression as an argument
			test.Args = append(test.Args, p.ParseExpression())

		} else if p.Match(TokenLparen) != nil {
			// one or more args can be passed with parentheses, e.g.: {% if 9 is divisibleby(3) %}
			noMoreArgs := false
			for p.Match(TokenComma) != nil || p.Match(TokenRparen) == nil {
				// parse args and kwargs
				v := p.ParseExpression()
				if p.Match(TokenAssign) != nil {
					key := v.Position().Val
					test.Kwargs[key] = p.ParseExpression()
					noMoreArgs = true
				} else {
					if noMoreArgs {
						errors.ThrowSyntaxError(p.Current().ErrorToken(), "positional argument must be before keyword argument")
					}
					test.Args = append(test.Args, v)
				}
			}

		} else if p.Peek(TokenEOF, TokenBlockEnd, TokenVariableEnd, TokenRawEnd) == nil {
			// one arg can be passed without parentheses, e.g.: {% if 9 is divisibleby 3 %}
			if arg := p.ParseOptionalExpression(); arg != nil {
				test.Args = append(test.Args, arg)
			}
		}

		expr = &TestExpression{
			Expression: expr,
			Test:       test,
		}

		if not != nil {
			expr = &NegationNode{expr, not}
		}
	}

	debug.Print("parsed expression: %s", expr)
	return expr
}
