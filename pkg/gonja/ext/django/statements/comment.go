package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type CommentStmt struct {
	Location *parse.Token
}

func (stmt *CommentStmt) Position() *parse.Token { return stmt.Location }
func (stmt *CommentStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("Block(Line=%d Col=%d)", t.Line, t.Col)
}

func commentParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	commentNode := &CommentStmt{p.Current()}
	p.SkipUntil("endcomment")
	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "tag 'comment' does not take any argument.")
	}
	return commentNode
}

func init() {
	All.MustRegister("comment", commentParser)
}
