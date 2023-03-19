package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/nodes"
	"github.com/aisbergg/gonja/parser"
	"github.com/aisbergg/gonja/tokens"
)

type IfEqualStmt struct {
	Location    *tokens.Token
	var1, var2  nodes.Expression
	thenWrapper *nodes.Wrapper
	elseWrapper *nodes.Wrapper
}

func (stmt *IfEqualStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *IfEqualStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IfEqualStmt(Line=%d Col=%d)", t.Line, t.Col)
}

// func (node *IfEqualStmt) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
// 	r1, err := node.var1.Evaluate(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	r2, err := node.var2.Evaluate(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	result := r1.EqualValueTo(r2)

// 	if result {
// 		return node.thenWrapper.Execute(ctx, writer)
// 	}
// 	if node.elseWrapper != nil {
// 		return node.elseWrapper.Execute(ctx, writer)
// 	}
// 	return nil
// }

func ifEqualParser(p *parser.Parser, args *parser.Parser) (nodes.Statement, error) {
	ifequalNode := &IfEqualStmt{}

	// Parse two expressions
	var1, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	var2, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	ifequalNode.var1 = var1
	ifequalNode.var2 = var2

	if !args.End() {
		return nil, args.Error("ifequal only takes 2 args.", nil)
	}

	// Wrap then/else-blocks
	wrapper, endargs, err := p.WrapUntil("else", "endifequal")
	if err != nil {
		return nil, err
	}
	ifequalNode.thenWrapper = wrapper

	if !endargs.End() {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	if wrapper.EndTag == "else" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, endargs, err = p.WrapUntil("endifequal")
		if err != nil {
			return nil, err
		}
		ifequalNode.elseWrapper = wrapper

		if !endargs.End() {
			return nil, endargs.Error("Arguments not allowed here.", nil)
		}
	}

	return ifequalNode, nil
}

func init() {
	All.MustRegister("ifequal", ifEqualParser)
}
