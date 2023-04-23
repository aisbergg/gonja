package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type ExtendsStmt struct {
	Location    *parse.Token
	Filename    string
	WithContext bool
}

var _ parse.Statement = (*ExtendsStmt)(nil)

func (stmt *ExtendsStmt) Position() *parse.Token { return stmt.Location }
func (stmt *ExtendsStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("ExtendsStmt(Filename=%s Line=%d Col=%d)", stmt.Filename, t.Line, t.Col)
}

func extendsParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &ExtendsStmt{
		Location: p.Current(),
	}

	if p.Level > 1 {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "the 'extends' statement can only be defined at root level")
	}

	if p.Template.Parent != nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "the template can only be extended once")
	}

	if filename := args.Match(parse.TokenString); filename != nil {
		stmt.Filename = filename.Val
		tpl, err := p.TemplateParseFn(stmt.Filename)
		if err != nil {
			errors.ThrowSyntaxError(p.Current().ErrorToken(), "unable to load parent template '%s': %s", stmt.Filename, err)
		}
		p.Template.Parent = tpl

	} else {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "tag 'extends' requires a template filename as string.")
	}

	if tok := args.MatchName("with", "without"); tok != nil {
		if args.MatchName("context") != nil {
			stmt.WithContext = tok.Val == "with"
		} else {
			args.Stream.Backup()
		}
	}

	if !args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "tag 'extends' does only take 1 argument.")
	}

	return stmt
}

func init() {
	All.MustRegister("extends", extendsParser)
}
