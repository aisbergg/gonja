package statements

import (
	"fmt"
	"math"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type WidthRatioStmt struct {
	Location     *parse.Token
	current, max parse.Expression
	width        parse.Expression
	ctxName      string
}

func (stmt *WidthRatioStmt) Position() *parse.Token { return stmt.Location }
func (stmt *WidthRatioStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("WidthRatioStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *WidthRatioStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	current := r.Eval(stmt.current)
	max := r.Eval(stmt.max)
	width := r.Eval(stmt.width)
	value := int(math.Ceil(current.Float()/max.Float()*width.Float() + 0.5))
	if stmt.ctxName == "" {
		r.WriteString(fmt.Sprintf("%d", value))
	} else {
		r.Ctx.Set(stmt.ctxName, value)
	}
}

func widthratioParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &WidthRatioStmt{
		Location: p.Current(),
	}

	stmt.current = args.ParseExpression()
	stmt.max = args.ParseExpression()
	stmt.width = args.ParseExpression()

	if args.MatchName("as") != nil {
		// Name follows
		nameToken := args.Match(parse.TokenName)
		if nameToken == nil {
			// return nil, args.Error("Expected name (identifier).", nil)
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "expected name (identifier)")
		}
		stmt.ctxName = nameToken.Val
	}

	if !args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "malformed widthratio-tag args")
	}

	return stmt
}

func init() {
	All.MustRegister("widthratio", widthratioParser)
}
