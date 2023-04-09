package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type IfNotEqualStmt struct {
	Location    *parse.Token
	var1, var2  parse.Expression
	thenWrapper *parse.WrapperNode
	elseWrapper *parse.WrapperNode
}

func (stmt *IfNotEqualStmt) Position() *parse.Token { return stmt.Location }
func (stmt *IfNotEqualStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IfNotEqualStmt(Line=%d Col=%d)", t.Line, t.Col)
}

// func (node *IfNotEqualStmt) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
// r.Current = stmt
// 	r1, err := node.var1.Evaluate(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	r2, err := node.var2.Evaluate(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	result := !r1.EqualValueTo(r2)

// 	if result {
// 		return node.thenWrapper.Execute(ctx, writer)
// 	}
// 	if node.elseWrapper != nil {
// 		return node.elseWrapper.Execute(ctx, writer)
// 	}
// }

func ifNotEqualParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	ifnotequalNode := &IfNotEqualStmt{}

	// Parse two expressions
	ifnotequalNode.var1 = args.ParseExpression()
	ifnotequalNode.var2 = args.ParseExpression()

	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "ifequal only takes 2 args")
	}

	// Wrap then/else-blocks
	wrapper, endargs := p.WrapUntil("else", "endifnotequal")
	ifnotequalNode.thenWrapper = wrapper

	if !endargs.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(endargs.Current()), "arguments not allowed here")
	}

	if wrapper.EndTag == "else" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, endargs = p.WrapUntil("endifnotequal")
		ifnotequalNode.elseWrapper = wrapper

		if !endargs.End() {
			errors.ThrowSyntaxError(parse.AsErrorToken(endargs.Current()), "arguments not allowed here")
		}
	}

	return ifnotequalNode
}

func init() {
	All.MustRegister("ifnotequal", ifNotEqualParser)
}
