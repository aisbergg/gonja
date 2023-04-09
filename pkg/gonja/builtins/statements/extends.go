package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type ExtendsStmt struct {
	Location    *parse.Token
	Filename    string
	WithContext bool
}

func (stmt *ExtendsStmt) Position() *parse.Token { return stmt.Location }
func (stmt *ExtendsStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("ExtendsStmt(Filename=%s Line=%d Col=%d)", stmt.Filename, t.Line, t.Col)
}

func (stmt *ExtendsStmt) Execute(r *exec.Renderer) error {
	r.Current = stmt
	return nil
}

func extendsParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &ExtendsStmt{
		Location: p.Current(),
	}

	if p.Level > 1 {
		errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "the 'extends' statement can only be defined at root level")
	}

	if p.Template.Parent != nil {
		errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "this template has already one parent")
	}

	// var filename parse.Node
	if filename := args.Match(parse.TokenString); filename != nil {
		stmt.Filename = filename.Val
		tpl, err := p.TemplateParser(stmt.Filename)
		if err != nil {
			errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "unable to parse parent template '%s'", stmt.Filename)
		}
		p.Template.Parent = tpl
	} else {
		errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "tag 'extends' requires a template filename as string.")
	}

	if tok := args.MatchName("with", "without"); tok != nil {
		if args.MatchName("context") != nil {
			stmt.WithContext = tok.Val == "with"
		} else {
			args.Stream.Backup()
		}
	}

	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "tag 'extends' does only take 1 argument.")
	}

	return stmt
}

func init() {
	All.MustRegister("extends", extendsParser)
}
