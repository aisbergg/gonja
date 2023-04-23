package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// ImportStmt is a statement that imports a template and makes its macros
// available.
type ImportStmt struct {
	Location     *parse.Token
	Filename     string
	FilenameExpr parse.Expression
	As           string
	WithContext  bool
	Template     *parse.TemplateNode
}

var _ parse.Statement = (*ImportStmt)(nil)
var _ exec.Statement = (*ImportStmt)(nil)

// Position returns the position of the statement.
func (stmt *ImportStmt) Position() *parse.Token { return stmt.Location }
func (stmt *ImportStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("ImportStmt(Line=%d Col=%d)", t.Line, t.Col)
}

// Execute executes the import statement.
func (stmt *ImportStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	var imported map[string]*parse.MacroNode
	macros := map[string]exec.Macro{}

	if stmt.FilenameExpr != nil {
		filename := r.Eval(stmt.FilenameExpr).String()
		tpl, err := r.TemplateLoadFn(filename)
		if err != nil {
			errors.ThrowTemplateRuntimeError("unable to load template '%s': %s", filename, err)
		}
		imported = tpl.Root.Macros

	} else {
		imported = stmt.Template.Macros
	}

	for name, macro := range imported {
		fn := exec.MacroNodeToFunc(macro, r)
		macros[name] = fn
	}

	r.Ctx.Set(stmt.As, macros)
}

// FromImportStmt is a statement that imports macros from another template.
type FromImportStmt struct {
	Location     *parse.Token
	Filename     string
	FilenameExpr parse.Expression
	WithContext  bool
	Template     *parse.TemplateNode
	As           map[string]string
	Macros       map[string]*parse.MacroNode // alias/name -> macro instance
}

// Position returns the position of the statement.
func (stmt *FromImportStmt) Position() *parse.Token { return stmt.Location }
func (stmt *FromImportStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("FromImportStmt(Line=%d Col=%d)", t.Line, t.Col)
}

// Execute executes the import statement.
func (stmt *FromImportStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	var imported map[string]*parse.MacroNode

	if stmt.FilenameExpr != nil {
		filename := r.Eval(stmt.FilenameExpr).String()
		tpl, err := r.TemplateLoadFn(filename)
		if err != nil {
			errors.ThrowTemplateRuntimeError("unable to load template '%s': %s", filename, err)
		}
		imported = tpl.Root.Macros

	} else {
		imported = stmt.Template.Macros
	}

	for alias, name := range stmt.As {
		node := imported[name]
		fn := exec.MacroNodeToFunc(node, r)
		r.Ctx.Set(alias, fn)
	}
}

func importParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &ImportStmt{
		Location: p.Current(),
		// Macros:   map[string]*parse.Macro{},
	}

	if args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "you must at least specify one macro to import.")
	}

	if tok := args.Match(parse.TokenString); tok != nil {
		stmt.Filename = tok.Val
	} else {
		expr := args.ParseExpression()
		stmt.FilenameExpr = expr
	}
	if args.MatchName("as") == nil {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "expected 'as' keyword, got '%s'", args.Current().Val)
	}

	alias := args.Match(parse.TokenName)
	if alias == nil {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "expected macro alias name (identifier), got '%s'", args.Current().Val)
	}
	stmt.As = alias.Val

	if tok := args.MatchName("with", "without"); tok != nil {
		if args.MatchName("context") != nil {
			stmt.WithContext = tok.Val == "with"
		} else {
			args.Stream.Backup()
		}
	}

	// Preload static template
	if stmt.Filename != "" {
		tpl, err := p.TemplateParseFn(stmt.Filename)
		if err != nil {
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "unable to parse imported template '%s'", stmt.Filename)
		}
		stmt.Template = tpl
	}

	return stmt
}

func fromParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &FromImportStmt{
		Location: p.Current(),
		As:       map[string]string{},
		// Macros:   map[string]*parse.Macro{},
	}

	if args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "you must at least specify one macro to import")
	}

	if tok := args.Match(parse.TokenString); tok != nil {
		stmt.Filename = tok.Val
	} else {
		filename := args.ParseExpression()
		stmt.FilenameExpr = filename
	}

	if args.MatchName("import") == nil {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "expected 'import' keyword, got '%s'", args.Current().Val)
	}

	for !args.End() {
		name := args.Match(parse.TokenName)
		if name == nil {
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "expected macro name (identifier), got '%s'", args.Current().Val)
		}

		// asName := macroNameToken.Val
		if args.MatchName("as") != nil {
			alias := args.Match(parse.TokenName)
			if alias == nil {
				errors.ThrowSyntaxError(args.Current().ErrorToken(), "expected macro alias name (identifier), got '%s'", args.Current().Val)
			}
			// asName = aliasToken.Val
			stmt.As[alias.Val] = name.Val
		} else {
			stmt.As[name.Val] = name.Val
		}

		// macroInstance, has := tpl.exportedMacros[macroNameToken.Val]
		// if !has {
		// 	return nil, args.Error(fmt.Sprintf("Macro '%s' not found (or not exported) in '%s'.", macroNameToken.Val,
		// 		stmt.filename), macroNameToken)
		// }

		// stmt.macros[asName] = macroInstance
		if tok := args.MatchName("with", "without"); tok != nil {
			if args.MatchName("context") != nil {
				stmt.WithContext = tok.Val == "with"
				break
			} else {
				args.Stream.Backup()
			}
		}

		if args.End() {
			break
		}

		if args.Match(parse.TokenComma) == nil {
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "unexpected '%s', expected ','", args.Current().Val)
		}
	}

	// Preload static template
	if stmt.Filename != "" {
		tpl, err := p.TemplateParseFn(stmt.Filename)
		if err != nil {
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "unable to parse imported template '%s'", stmt.Filename)
		}
		stmt.Template = tpl
	}

	return stmt
}

func init() {
	All.MustRegister("import", importParser)
	All.MustRegister("from", fromParser)
}
