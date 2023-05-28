package parse

import (
	debug "github.com/aisbergg/gonja/internal/debug/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

type StatementParser func(parser, args *Parser) Statement

// Tag = "{%" IDENT ARGS "%}"
func (p *Parser) ParseStatement() Statement {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	if p.Match(TokenBlockBegin) == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected '%s' , expected '{%%'", p.Current().Val)
	}

	name := p.Match(TokenName)
	if name == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "expected statement name, got '%s'", p.Current().Val)
	}

	// Check for the existing statement
	stmtParser, exists := p.Statements[name.Val]
	if !exists {
		// Does not exists
		errors.ThrowSyntaxError(name.ErrorToken(), "statement '%s' not found (or beginning not provided)", name)
	}

	// Check sandbox tag restriction
	// if _, isBanned := p.bannedStmts[tokenName.Val]; isBanned {
	// 	return nil, p.Error(fmt.Sprintf("Usage of statement '%s' is not allowed (sandbox restriction active).", tokenName.Val), tokenName)
	// }

	var args []*Token
	for !p.Stream.End() && p.Peek(TokenBlockEnd) == nil {
		// Add token to args
		args = append(args, p.Next())
		// p.Consume() // next token
	}

	// EOF?
	// if p.Remaining() == 0 {
	// 	return nil, p.Error("Unexpectedly reached EOF, no statement end found.", p.lastToken)
	// }

	if p.Match(TokenBlockEnd) == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "expected end of block '%s'", p.Config.BlockEndString)
	}

	argParser := NewParser(p.Config, NewStream(args))
	// argParser := newParser(p.name, argsToken, p.template)
	// if len(argsToken) == 0 {
	// 	// This is done to have nice EOF error messages
	// 	argParser.lastToken = tokenName
	// }

	p.Level++
	defer func() { p.Level-- }()
	return stmtParser(p, argParser)
}

// type StatementParser func(parser *Parser, args *Parser) (Stmt, error)

func (p *Parser) ParseStatementBlock() *StatementBlockNode {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	begin := p.Match(TokenBlockBegin)
	if begin == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected '%s', expected '%s'", p.Current(), p.Config.BlockStartString)
	}

	name := p.Match(TokenName)
	if name == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "expected statement name, got '%s'", p.Current().Val)
	}

	// Check for the existing statement
	stmtParser, exists := p.Statements[name.Val]
	if !exists {
		// Does not exists
		errors.ThrowSyntaxError(name.ErrorToken(), "statement '%s' not found (or beginning not provided)", name.Val)
	}

	// Check sandbox tag restriction
	// if _, isBanned := p.bannedStmts[tokenName.Val]; isBanned {
	// 	return nil, p.Error(fmt.Sprintf("Usage of statement '%s' is not allowed (sandbox restriction active).", tokenName.Val), tokenName)
	// }

	debug.Print("find args token")
	var args []*Token
	for !p.Stream.End() && p.Peek(TokenBlockEnd) == nil {
		args = append(args, p.Next())
	}

	// EOF?
	// if p.Remaining() == 0 {
	// 	return nil, p.Error("Unexpectedly reached EOF, no statement end found.", p.lastToken)
	// }

	end := p.Match(TokenBlockEnd)
	if end == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "expected end of block '%s'", p.Config.BlockEndString)
	}

	stream := NewStream(args)
	debug.Print("argparser")
	argParser := NewParser(p.Config, stream)
	// argParser := newParser(p.name, argsToken, p.template)
	// if len(argsToken) == 0 {
	// 	// This is done to have nice EOF error messages
	// 	argParser.lastToken = tokenName
	// }

	// p.template.level++
	// defer func() { p.template.level-- }()
	stmt := stmtParser(p, argParser)
	debug.Print("parsed expression: %s", stmt)
	return &StatementBlockNode{
		Location: begin,
		Name:     name.Val,
		Stmt:     stmt,
		LStrip:   begin.Val[len(begin.Val)-1] == '+',
		Trim: &Trim{
			Left:  begin.Val[len(begin.Val)-1] == '-',
			Right: end.Val[0] == '-',
		},
	}
}
