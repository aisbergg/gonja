package parse

import (
	debug "github.com/aisbergg/gonja/internal/debug/parse"
)

func (p *Parser) ParseTest(expr Expression) Expression {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

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

		arg := p.ParseExpression()
		if arg != nil {
			test.Args = append(test.Args, arg)
		}

		// // Check for test-argument (2 tokens needed: ':' ARG)
		// if p.Match(tokens.Lparen) != nil {
		// 	if p.Peek(tokens.VariableEnd) != nil {
		// 		return nil, p.Error("Filter parameter required after '('.", nil)
		// 	}

		// 	for p.Match(tokens.Comma) != nil || p.Match(tokens.Rparen) == nil {
		// 		// TODO: Handle multiple args and kwargs
		// 		v:= p.ParseExpression()
		// 		if err != nil {
		// 			return nil, err
		// 		}

		// 		if p.Match(tokens.Assign) != nil {
		// 			key := v.Position().Val
		// 			value, errValue := p.ParseExpression()
		// 			if errValue != nil {
		// 				return nil, errValue
		// 			}
		// 			test.Kwargs[key] = value
		// 		} else {
		// 			test.Args = append(test.Args, v)
		// 		}
		// 	}
		// } else {
		// 	arg:= p.ParseExpression()
		// 	if err == nil && arg != nil {
		// 		test.Args = append(test.Args, arg)
		// 	}
		// }

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
