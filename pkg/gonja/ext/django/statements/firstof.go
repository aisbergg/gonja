package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type FirstofStmt struct {
	Location *parse.Token
	Args     []parse.Expression
}

func (stmt *FirstofStmt) Position() *parse.Token { return stmt.Location }
func (stmt *FirstofStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("FirstofStmt(Args=%s, Line=%d Col=%d)", stmt.Args, t.Line, t.Col)
}

func (stmt *FirstofStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	for _, arg := range stmt.Args {
		val := r.Eval(arg)

		if val.IsTrue() {
			r.RenderValue(val)
			return
		}
	}
}

func firstofParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &FirstofStmt{
		Location: p.Current(),
	}

	for !args.End() {
		node := args.ParseExpression()
		stmt.Args = append(stmt.Args, node)
	}

	return stmt
}

func init() {
	All.MustRegister("firstof", firstofParser)
}
