package parse

import (
	debug "github.com/aisbergg/gonja/internal/debug/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// ParseFilter parses a filter.
func (p *Parser) ParseFilter() *FilterCall {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())
	identToken := p.Match(TokenName)

	// Check filter ident
	if identToken == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "filter name must be an identifier")
	}

	filter := &FilterCall{
		Token:  identToken,
		Name:   identToken.Val,
		Args:   []Expression{},
		Kwargs: map[string]Expression{},
	}

	if p.Match(TokenLparen) != nil {
		noMoreArgs := false
		for p.Match(TokenComma) != nil || p.Match(TokenRparen) == nil {
			// parse args and kwargs
			v := p.ParseExpression()
			if p.Match(TokenAssign) != nil {
				key := v.Position().Val
				filter.Kwargs[key] = p.ParseExpression()
				noMoreArgs = true
			} else {
				if noMoreArgs {
					errors.ThrowSyntaxError(p.Current().ErrorToken(), "positional argument must be before keyword argument")
				}
				filter.Args = append(filter.Args, v)
			}
		}
	}

	debug.Print("parsed expression: %s", filter)
	return filter
}
