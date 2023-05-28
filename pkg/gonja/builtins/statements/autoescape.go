package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type AutoescapeStmt struct {
	Wrapper    *parse.WrapperNode
	Autoescape bool
}

var (
	_ parse.Statement = (*AutoescapeStmt)(nil)
	_ exec.Statement  = (*AutoescapeStmt)(nil)
)

func (stmt *AutoescapeStmt) Position() *parse.Token { return stmt.Wrapper.Position() }
func (stmt *AutoescapeStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("AutoescapeStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *AutoescapeStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	sub := r.Inherit()
	sub.Autoescape = stmt.Autoescape

	err := sub.ExecuteWrapper(stmt.Wrapper)
	if err != nil {
		// pass error up the call stack
		panic(err)
	}
}

func autoescapeParser(p, args *parse.Parser) parse.Statement {
	stmt := &AutoescapeStmt{}

	wrapper, _ := p.WrapUntil("endautoescape")
	stmt.Wrapper = wrapper

	modeToken := args.Match(parse.TokenName)
	if modeToken == nil {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "a mode is required for autoescape statement")
	}
	if modeToken.Val == "true" {
		stmt.Autoescape = true
	} else if modeToken.Val == "false" {
		stmt.Autoescape = false
	} else {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "only 'true' or 'false' is valid as an autoescape statement.")
	}

	if !args.Stream.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "malformed autoescape statement args")
	}

	return stmt
}

func init() {
	All.MustRegister("autoescape", autoescapeParser)
}
