package statements

import (
	"fmt"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type IfChangedStmt struct {
	Location    *parse.Token
	watchedExpr []parse.Expression
	lastValues  []exec.Value
	lastContent string
	thenWrapper *parse.WrapperNode
	elseWrapper *parse.WrapperNode
}

var _ parse.Statement = (*IfChangedStmt)(nil)
var _ exec.Statement = (*IfChangedStmt)(nil)

func (stmt *IfChangedStmt) Position() *parse.Token { return stmt.Location }
func (stmt *IfChangedStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IfChangedStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *IfChangedStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	if len(stmt.watchedExpr) == 0 {
		// Check against own rendered body
		var out strings.Builder
		sub := r.Inherit()
		sub.Out = &out
		err := sub.ExecuteWrapper(stmt.thenWrapper)
		if err != nil {
			panic(err)
		}

		str := out.String()
		if stmt.lastContent != str {
			// Rendered content changed, output it
			r.WriteString(str)
			stmt.lastContent = str
		}
	} else {
		nowValues := make([]exec.Value, 0, len(stmt.watchedExpr))
		for _, expr := range stmt.watchedExpr {
			val := r.Eval(expr)
			nowValues = append(nowValues, val)
		}

		// Compare old to new values now
		changed := len(stmt.lastValues) == 0

		for idx, oldVal := range stmt.lastValues {
			if !oldVal.EqualValueTo(nowValues[idx]) {
				changed = true
				break // we can stop here because ONE value changed
			}
		}

		stmt.lastValues = nowValues

		if changed {
			// Render thenWrapper
			err := r.ExecuteWrapper(stmt.thenWrapper)
			if err != nil {
				// pass error up the call stack
				panic(err)
			}
		} else {
			// Render elseWrapper
			err := r.ExecuteWrapper(stmt.elseWrapper)
			if err != nil {
				// pass error up the call stack
				panic(err)
			}
		}
	}
}

func ifchangedParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &IfChangedStmt{
		Location: p.Current(),
	}

	for !args.End() {
		// Parse condition
		expr := args.ParseExpression()
		stmt.watchedExpr = append(stmt.watchedExpr, expr)
	}

	if !args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "ifchanged-arguments are malformed")
	}

	// Wrap then/else-blocks
	wrapper, endargs := p.WrapUntil("else", "endifchanged")
	stmt.thenWrapper = wrapper

	if !endargs.End() {
		errors.ThrowSyntaxError(endargs.Current().ErrorToken(), "arguments not allowed here")
	}

	if wrapper.EndTag == "else" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, endargs = p.WrapUntil("endifchanged")
		stmt.elseWrapper = wrapper

		if !endargs.End() {
			errors.ThrowSyntaxError(endargs.Current().ErrorToken(), "arguments not allowed here")
		}
	}

	return stmt
}

func init() {
	All.MustRegister("ifchanged", ifchangedParser)
}
