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

	// // Get the appropriate filter function and bind it
	// filterFn, exists := filters[identToken.Val]
	// if !exists {
	// 	return nil, p.Error(fmt.Sprintf("Filter '%s' does not exist.", identToken.Val), identToken)
	// }

	// filter.filterFunc = filterFn

	// Check for filter-argument (2 tokens needed: ':' ARG)
	if p.Match(TokenLparen) != nil {
		if p.Peek(TokenVariableEnd) != nil {
			errors.ThrowSyntaxError(p.Current().ErrorToken(), "filter parameter required after '('")
		}

		for p.Match(TokenComma) != nil || p.Match(TokenRparen) == nil {
			// TODO: Handle multiple args and kwargs
			v := p.ParseExpression()

			if p.Match(TokenAssign) != nil {
				key := v.Position().Val
				filter.Kwargs[key] = p.ParseExpression()
			} else {
				filter.Args = append(filter.Args, v)
			}
		}
	}

	debug.Print("parsed expression: %s", filter)
	return filter
}
