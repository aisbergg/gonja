package statements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type SpacelessStmt struct {
	Location *parse.Token
	wrapper  *parse.WrapperNode
}

func (stmt *SpacelessStmt) Position() *parse.Token { return stmt.Location }
func (stmt *SpacelessStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("SpacelessStmt(Line=%d Col=%d)", t.Line, t.Col)
}

var spacelessRegexp = regexp.MustCompile(`(?U:(<.*>))([\t\n\v\f\r ]+)(?U:(<.*>))`)

func (stmt *SpacelessStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	var out strings.Builder

	sub := r.Inherit()
	sub.Out = &out
	if err := sub.ExecuteWrapper(stmt.wrapper); err != nil {
		panic(err)
	}

	s := out.String()
	// Repeat this recursively
	changed := true
	for changed {
		s2 := spacelessRegexp.ReplaceAllString(s, "$1$3")
		changed = s != s2
		s = s2
	}

	r.WriteString(s)
}

func spacelessParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &SpacelessStmt{
		Location: p.Current(),
	}

	wrapper, _ := p.WrapUntil("endspaceless")
	stmt.wrapper = wrapper

	if !args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "malformed spaceless-tag args")
	}

	return stmt
}

func init() {
	All.MustRegister("spaceless", spacelessParser)
}
