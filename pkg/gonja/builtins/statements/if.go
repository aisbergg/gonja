package statements

import (
	"fmt"

	log "github.com/aisbergg/gonja/internal/log/parse"
	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type IfStmt struct {
	Location   *parse.Token
	conditions []parse.Expression
	wrappers   []*parse.WrapperNode
}

func (stmt *IfStmt) Position() *parse.Token { return stmt.Location }
func (stmt *IfStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IfStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *IfStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	for i, condition := range stmt.conditions {
		result := r.Eval(condition)

		if result.IsTrue() {
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

func ifParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	ifNode := &IfStmt{
		Location: args.Current(),
	}

	// Parse first and main IF condition
	condition := args.ParseExpression()
	ifNode.conditions = append(ifNode.conditions, condition)

	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "if-condition is malformed")
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
				errors.ThrowSyntaxError(parse.AsErrorToken(tagArgs.Current()), "elif-condition is malformed")
			}
		} else {
			if !tagArgs.End() {
				// else/endif can't take any conditions
				errors.ThrowSyntaxError(parse.AsErrorToken(tagArgs.Current()), "arguments not allowed here")
			}
		}

		if wrapper.EndTag == "endif" {
			break
		}
	}

	log.Print("parsed expression: %s", ifNode)
	return ifNode
}

func init() {
	All.MustRegister("if", ifParser)
}
