package statements

import (
	"fmt"

	debug "github.com/aisbergg/gonja/internal/debug/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type IfStmt struct {
	Location   *parse.Token
	conditions []parse.Expression
	wrappers   []*parse.WrapperNode
}

var (
	_ parse.Statement = (*IfStmt)(nil)
	_ exec.Statement  = (*IfStmt)(nil)
)

func (stmt *IfStmt) Position() *parse.Token { return stmt.Location }
func (stmt *IfStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IfStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *IfStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	for i, condition := range stmt.conditions {
		result := r.Eval(condition)

		if result.Bool() {
			if err := r.ExecuteWrapper(stmt.wrappers[i]); err != nil {
				panic(err)
			}
			return
		}
	}

	// else block (has no condition)
	if len(stmt.wrappers) > len(stmt.conditions) {
		if err := r.ExecuteWrapper(stmt.wrappers[len(stmt.wrappers)-1]); err != nil {
			panic(err)
		}
	}
}

func ifParser(p, args *parse.Parser) parse.Statement {
	if debug.Enabled {
		fm := debug.FuncMarker()
		defer fm.End()
	}
	debug.Print("parse: %s", p.Current())

	ifNode := &IfStmt{
		Location: args.Current(),
	}

	// Parse first and main IF condition
	condition := args.ParseExpression()
	ifNode.conditions = append(ifNode.conditions, condition)

	if !args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "if-condition is malformed")
	}

	// Check the rest
	for {
		wrapper, tagArgs := p.WrapUntil("elif", "else", "endif")
		ifNode.wrappers = append(ifNode.wrappers, wrapper)

		if wrapper.EndTag == "elif" {
			// elif can take a condition
			condition = tagArgs.ParseExpression()
			ifNode.conditions = append(ifNode.conditions, condition)

			if !tagArgs.End() {
				errors.ThrowSyntaxError(tagArgs.Current().ErrorToken(), "elif-condition is malformed")
			}
		} else {
			if !tagArgs.End() {
				// else/endif can't take any conditions
				errors.ThrowSyntaxError(tagArgs.Current().ErrorToken(), "arguments not allowed here")
			}
		}

		if wrapper.EndTag == "endif" {
			break
		}
	}

	debug.Print("parsed expression: %s", ifNode)
	return ifNode
}

func init() {
	All.MustRegister("if", ifParser)
}
