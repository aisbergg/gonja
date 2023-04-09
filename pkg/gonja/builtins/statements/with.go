package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type WithStmt struct {
	Location *parse.Token
	Pairs    map[string]parse.Expression
	Wrapper  *parse.WrapperNode
}

func (stmt *WithStmt) Position() *parse.Token { return stmt.Location }
func (stmt *WithStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("WithStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *WithStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	sub := r.Inherit()

	for key, value := range stmt.Pairs {
		val := r.Eval(value)
		sub.Ctx.Set(key, val)
	}

	if err := sub.ExecuteWrapper(stmt.Wrapper); err != nil {
		// pass error up the stack
		panic(err)
	}
}

func withParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &WithStmt{
		Location: p.Current(),
		Pairs:    map[string]parse.Expression{},
	}

	wrapper, endargs := p.WrapUntil("endwith")
	stmt.Wrapper = wrapper

	if !endargs.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(endargs.Current()), "arguments not allowed here")
	}

	for !args.End() {
		key := args.Match(parse.TokenName)
		if key == nil {
			errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "expected an identifier")
		}
		if args.Match(parse.TokenAssign) == nil {
			errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "unexpected '%s', expected '='", args.Current().Val)
		}
		value := args.ParseExpression()
		stmt.Pairs[key.Val] = value

		if args.Match(parse.TokenComma) == nil {
			break
		}
	}

	if !args.End() {
		return nil
	}

	return stmt
}

func init() {
	All.MustRegister("with", withParser)
}
