package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type RawStmt struct {
	Data *parse.DataNode
	// Content string
}

func (stmt *RawStmt) Position() *parse.Token { return stmt.Data.Position() }
func (stmt *RawStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("RawStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *RawStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	r.WriteString(stmt.Data.Data.Val)
}

func rawParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &RawStmt{}

	wrapper, _ := p.WrapUntil("endraw")
	node := wrapper.Nodes[0]
	data, ok := node.(*parse.DataNode)
	if ok {
		stmt.Data = data
	} else {
		errors.ThrowSyntaxError(parse.AsErrorToken(node.Position()), "raw statement can only contains a single data node")
	}

	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "raw statement doesn't accept parameters.")
	}

	return stmt
}

func init() {
	All.MustRegister("raw", rawParser)
}
