package parse

import (
	debug "github.com/aisbergg/gonja/internal/debug/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// ParseComment parses a comment and returns a CommentNode.
func (p *Parser) ParseComment() *CommentNode {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	tok := p.Match(TokenCommentBegin)
	if tok == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected '%s' , expected '%s'", p.Current(), p.Config.CommentStartString)
	}

	comment := &CommentNode{
		Start: tok,
		Trim:  &Trim{},
	}

	tok = p.Match(TokenData)
	if tok == nil {
		comment.Text = ""
	} else {
		comment.Text = tok.Val
	}

	tok = p.Match(TokenCommentEnd)
	if tok == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected '%s' , expected '%s'", p.Current(), p.Config.CommentEndString)
	}
	comment.End = tok

	debug.Print("parsed expression: %s", comment)
	return comment
}
