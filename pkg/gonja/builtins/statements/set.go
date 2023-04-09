package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type SetStmt struct {
	Location   *parse.Token
	Target     parse.Expression
	Expression parse.Expression
}

func (stmt *SetStmt) Position() *parse.Token { return stmt.Location }
func (stmt *SetStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("SetStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *SetStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	// Evaluate expression
	value := r.Eval(stmt.Expression)
	r.Current = stmt

	switch n := stmt.Target.(type) {
	case *parse.NameNode:
		r.Ctx.Set(n.Name.Val, value.Interface())

	case *parse.GetItemNode:
		target := r.Eval(n.Node)
		target.Set(n.Arg, value.Interface())

	default:
		errors.ThrowTemplateRuntimeError("illegal set target node %s", n)
	}
}

func setParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &SetStmt{
		Location: p.Current(),
	}

	// Parse variable name
	ident := args.ParseVariable()
	switch n := ident.(type) {
	case *parse.NameNode, *parse.CallNode, *parse.GetItemNode:
		stmt.Target = n
	default:
		errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "unexpected set target '%s'", n)
	}

	if args.Match(parse.TokenAssign) == nil {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "unexpected '%s', expected '='", args.Current().Val)
	}

	// Variable expression
	stmt.Expression = args.ParseExpression()

	// Remaining arguments
	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "malformed 'set'-tag args")
	}

	return stmt
}

func init() {
	All.MustRegister("set", setParser)
}
