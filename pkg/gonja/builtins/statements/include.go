package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// IncludeStmt is a statement that includes another template.
type IncludeStmt struct {
	Location      *parse.Token
	Filename      string
	FilenameExpr  parse.Expression
	Template      *parse.TemplateNode
	IgnoreMissing bool
	WithContext   bool
	IsEmpty       bool
}

// Position returns the token position of the statement.
func (stmt *IncludeStmt) Position() *parse.Token { return stmt.Location }
func (stmt *IncludeStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("IncludeStmt(Filename=%s Line=%d Col=%d)", stmt.Filename, t.Line, t.Col)
}

// Execute executes the include statement.
func (stmt *IncludeStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	if stmt.IsEmpty {
		return
	}
	sub := r.Inherit()

	if stmt.FilenameExpr != nil {
		filename := r.Eval(stmt.FilenameExpr).String()
		included, err := r.Loader.GetTemplate(filename)
		if err != nil {
			if stmt.IgnoreMissing {
				return
			}
			errors.ThrowTemplateRuntimeError("unable to load template '%s': %s", filename, err)
		}
		sub.Template = included
		sub.Root = included.Root

	} else {
		sub.Root = stmt.Template
	}

	if err := sub.Execute(); err != nil {
		// pass error up the stack
		panic(err)
	}
}

type IncludeEmptyStmt struct{}

// func (node *IncludeEmptyStmt) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
// 	return nil
// }

func includeParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &IncludeStmt{
		Location: p.Current(),
	}

	if tok := args.Match(parse.TokenString); tok != nil {
		stmt.Filename = tok.Val
	} else {
		filename := args.ParseExpression()
		stmt.FilenameExpr = filename
	}

	if args.MatchName("ignore") != nil {
		if args.MatchName("missing") != nil {
			stmt.IgnoreMissing = true
		} else {
			args.Stream.Backup()
		}
	}

	if tok := args.MatchName("with", "without"); tok != nil {
		if args.MatchName("context") != nil {
			stmt.WithContext = tok.Val == "with"
		} else {
			args.Stream.Backup()
		}
	}

	// Preload static template
	if stmt.Filename != "" {
		tpl, err := p.TemplateParser(stmt.Filename)
		if err != nil {
			if stmt.IgnoreMissing {
				stmt.IsEmpty = true
			} else {
				errors.ThrowSyntaxError(parse.AsErrorToken(stmt.Location), "unable to parse included template '%s'", stmt.Filename)
			}
		} else {
			stmt.Template = tpl
		}
	}

	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "malformed 'include'-tag args.")
	}

	return stmt
}

func init() {
	All.MustRegister("include", includeParser)
}
