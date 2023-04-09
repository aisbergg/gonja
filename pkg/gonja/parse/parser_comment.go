package parse

import (
	log "github.com/aisbergg/gonja/internal/log/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// ParseComment parses a comment and returns a CommentNode.
func (p *Parser) ParseComment() *CommentNode {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	tok := p.Match(TokenCommentBegin)
	if tok == nil {
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s' , expected '%s'", p.Current(), p.Config.CommentStartString)
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
		errors.ThrowSyntaxError(AsErrorToken(p.Current()), "unexpected '%s' , expected '%s'", p.Current(), p.Config.CommentEndString)
	}
	comment.End = tok

	log.Print("parsed expression: %s", comment)
	return comment
}
