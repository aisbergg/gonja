package parse

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/aisbergg/gonja/internal/log/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// parseNumber parses a number.
func (p *Parser) parseNumber() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())
	t := p.Match(TokenInteger, TokenFloat)
	if t == nil {
		errors.ThrowSyntaxError(AsErrorToken(t), "expected a number")
	}

	if t.Type == TokenInteger {
		i, err := strconv.Atoi(t.Val)
		if err != nil {
			errors.ThrowSyntaxError(AsErrorToken(p.Current()), err.Error())
		}
		nr := &IntegerNode{
			Location: t,
			Val:      i,
		}
		return nr
	}
	f, err := strconv.ParseFloat(t.Val, 64)
	if err != nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), err.Error())
	}
	fr := &FloatNode{
		Location: t,
		Val:      f,
	}
	return fr
}

// parseString parses a string.
func (p *Parser) parseString() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())
	t := p.Match(TokenString)
	if t == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "expected a string")
	}
	str := strconv.Quote(t.Val)
	replaced := strings.Replace(str, `\\`, `\`, -1)
	newstr, err := strconv.Unquote(replaced)
	if err != nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), err.Error())
	}
	sr := &StringNode{
		Location: t,
		Val:      newstr,
	}
	return sr
}

// parseCollection parses a collection.
func (p *Parser) parseCollection() Expression {
	switch p.Current().Type {
	case TokenLbracket:
		return p.parseList()
	case TokenLparen:
		return p.parseTuple()
	case TokenLbrace:
		return p.parseDict()
	default:
		return nil
	}
}

// parseList parses a list.
func (p *Parser) parseList() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())
	t := p.Match(TokenLbracket)
	if t == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s', expected '['", t.Val)
	}

	if p.Match(TokenRbracket) != nil {
		// Empty list
		return &ListNode{t, []Expression{}}
	}

	expr := p.ParseExpression()
	list := []Expression{expr}

	for p.Match(TokenComma) != nil {
		if p.Peek(TokenRbracket) != nil {
			// Trailing coma
			break
		}
		expr := p.ParseExpression()
		if expr == nil {
			errors.ThrowSyntaxError(AsErrorToken(p.Current()), "expected a value")
		}
		list = append(list, expr)
	}

	if p.Match(TokenRbracket) == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s', expected ']'", t.Val)
	}

	return &ListNode{t, list}
}

// parseTuple parses a tuple.
func (p *Parser) parseTuple() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())
	t := p.Match(TokenLparen)
	if t == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s', expected '('", t.Val)
	}
	expr := p.ParseExpression()
	list := []Expression{expr}

	trailingComa := false

	for p.Match(TokenComma) != nil {
		if p.Peek(TokenRparen) != nil {
			// Trailing coma
			trailingComa = true
			break
		}
		expr := p.ParseExpression()
		if expr == nil {
			errors.ThrowSyntaxError(AsErrorToken(p.Current()), "expected a value")
		}
		list = append(list, expr)
	}

	if p.Match(TokenRparen) == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unbalanced parenthesis '()'")
	}

	if len(list) > 1 || trailingComa {
		return &TupleNode{t, list}
	}
	return expr
}

// parsePair parses a pair.
func (p *Parser) parsePair() *PairNode {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())
	key := p.ParseExpression()

	if p.Match(TokenColon) == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s', expected ':'", p.Current())
	}
	value := p.ParseExpression()
	return &PairNode{
		Key:   key,
		Value: value,
	}
}

// parseDict parses a dict.
func (p *Parser) parseDict() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())
	t := p.Match(TokenLbrace)
	if t == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s', expected '{'", p.Current())
	}

	dict := &DictNode{
		Token: t,
		Pairs: []*PairNode{},
	}

	if p.Peek(TokenRbrace) == nil {
		pair := p.parsePair()
		dict.Pairs = append(dict.Pairs, pair)
	}

	for p.Match(TokenComma) != nil {
		pair := p.parsePair()
		dict.Pairs = append(dict.Pairs, pair)
	}

	if p.Match(TokenRbrace) == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s', expected '}'", p.Current())
	}

	return dict
}

// ParseVariable parses a variable.
func (p *Parser) ParseVariable() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	t := p.Match(TokenName)
	if t == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "expected an identifier")
	}

	switch t.Val {
	case "true", "True":
		br := &BoolNode{
			Location: t,
			Val:      true,
		}
		return br
	case "false", "False":
		br := &BoolNode{
			Location: t,
			Val:      false,
		}
		return br
	}

	var variable Node = &NameNode{t}

	for !p.Stream.EOF() {
		if dot := p.Match(TokenDot); dot != nil {
			getitem := &GetItemNode{
				Location: dot,
				Node:     variable,
			}
			tok := p.Match(TokenName, TokenInteger)
			if tok == nil {
				errors.ThrowSyntaxError(AsErrorToken(p.Current()), "expected an identifier or an integer")
			}
			switch tok.Type {
			case TokenName:
				getitem.Arg = tok.Val
			case TokenInteger:
				i, err := strconv.Atoi(tok.Val)
				if err != nil {
					errors.ThrowSyntaxError(AsErrorToken(p.Current()), err.Error())
				}
				getitem.Index = i
			default:
				panic(fmt.Errorf("BUG: token '%s' not allowed here.", p.Current()))
			}
			variable = getitem
			continue

		} else if bracket := p.Match(TokenLbracket); bracket != nil {
			getitem := &GetItemNode{
				Location: bracket,
				Node:     variable,
			}
			tok := p.Match(TokenString, TokenInteger)
			if tok == nil {
				errors.ThrowSyntaxError(AsErrorToken(p.Current()), "expected a string or an integer")
			}
			switch tok.Type {
			case TokenString:
				getitem.Arg = tok.Val
			case TokenInteger:
				i, err := strconv.Atoi(tok.Val)
				if err != nil {
					errors.ThrowSyntaxError(AsErrorToken(p.Current()), err.Error())

				}
				getitem.Index = i
			default:
				panic(fmt.Errorf("BUG: token '%s' not allowed here", p.Current()))
			}
			variable = getitem
			if p.Match(TokenRbracket) == nil {
				errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unbalanced bracket '[]'")
			}
			continue

		} else if lparen := p.Match(TokenLparen); lparen != nil {
			call := &CallNode{
				Location: lparen,
				Func:     variable,
				Args:     []Expression{},
				Kwargs:   map[string]Expression{},
			}
			// if p.Peek(tokens.VariableEnd) != nil {
			// 	return nil, p.Error("Filter parameter required after '('.")
			// }

			for p.Match(TokenComma) != nil || p.Match(TokenRparen) == nil {
				// TODO: Handle multiple args and kwargs
				v := p.ParseExpression()

				if p.Match(TokenAssign) != nil {
					key := v.Position().Val
					call.Kwargs[key] = p.ParseExpression()
				} else {
					call.Args = append(call.Args, v)
				}
			}
			variable = call
			// We're done parsing the function call, next variable part
			continue
		}

		// No dot or function call? Then we're done with the variable parsing
		break
	}

	return variable
}

// ParseVariableOrLiteral parses a variable or a literal.
func (p *Parser) ParseVariableOrLiteral() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())
	t := p.Current()

	if t == nil {
		errors.ThrowSyntaxError(AsErrorToken(t), "unexpected EOF, expected a number, string, keyword or identifier")
	}

	// Is first part a number or a string, there's nothing to resolve (because there's only to return the value then)
	switch t.Type {
	case TokenInteger, TokenFloat:
		return p.parseNumber()

	case TokenString:
		return p.parseString()

	case TokenLparen, TokenLbrace, TokenLbracket:
		return p.parseCollection()

	case TokenName:
		return p.ParseVariable()
	}

	errors.ThrowSyntaxError(AsErrorToken(p.Current()), "expected a number, string, keyword or identifier")
	return nil
}
