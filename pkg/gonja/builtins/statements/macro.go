package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type MacroStmt struct {
	*parse.MacroNode
}

// func (stmt *MacroStmt) Position() *tokens.Token { return stmt.Location }
func (stmt *MacroStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("MacroStmt(Macro=%s Line=%d Col=%d)", stmt.MacroNode, t.Line, t.Col)
}

func (stmt *MacroStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	macro := exec.MacroNodeToFunc(stmt.MacroNode, r)
	r.Ctx.Set(stmt.Name, macro)
}

// func (node *MacroStmt) call(ctx *exec.Context, args ...*exec.Value) *exec.Value {
// 	// argsCtx := make(exec.Context)

// 	// for k, v := range node.args {
// 	// 	if v == nil {
// 	// 		// User did not provided a default value
// 	// 		argsCtx[k] = nil
// 	// 	} else {
// 	// 		// Evaluate the default value
// 	// 		valueExpr, err := v.Evaluate(ctx)
// 	// 		if err != nil {
// 	// 			ctx.Logf(err.Error())
// 	// 			return AsSafeValue(err.Error())
// 	// 		}

// 	// 		argsCtx[k] = valueExpr
// 	// 	}
// 	// }

// 	// if len(args) > len(node.argsOrder) {
// 	// 	// Too many arguments, we're ignoring them and just logging into debug mode.
// 	// 	err := ctx.Error(fmt.Sprintf("Macro '%s' called with too many arguments (%d instead of %d).",
// 	// 		node.name, len(args), len(node.argsOrder)), nil).updateFromTokenIfNeeded(ctx.template, node.position)

// 	// 	ctx.Logf(err.Error()) // TODO: This is a workaround, because the error is not returned yet to the Execution()-methods
// 	// 	return AsSafeValue(err.Error())
// 	// }

// 	// // Make a context for the macro execution
// 	// macroCtx := NewChildExecutionContext(ctx)

// 	// // Register all arguments in the private context
// 	// macroCtx.Private.Update(argsCtx)

// 	// for idx, argValue := range args {
// 	// 	macroCtx.Private[node.argsOrder[idx]] = argValue.Interface()
// 	// }

// 	// var b bytes.Buffer
// 	// err := node.wrapper.Execute(macroCtx, &b)
// 	// if err != nil {
// 	// 	return AsSafeValue(err.updateFromTokenIfNeeded(ctx.template, node.position).Error())
// 	// }

// 	// return AsSafeValue(b.String())
// 	return nil
// }

func macroParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &parse.MacroNode{
		Location: p.Current(),
		Args:     []string{},
		Kwargs:   []*parse.PairNode{},
	}

	name := args.Match(parse.TokenName)
	if name == nil {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "macro-tag needs at least an identifier as name.")
	}
	stmt.Name = name.Val

	if args.Match(parse.TokenLparen) == nil {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "unexpected '%s', expected '('", args.Current().Val)
	}

	for args.Match(parse.TokenRparen) == nil {
		argName := args.Match(parse.TokenName)
		if argName == nil {
			errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "expected argument name as identifier.")
		}

		if args.Match(parse.TokenAssign) != nil {
			// Default expression follows
			expr := args.ParseExpression()
			stmt.Kwargs = append(stmt.Kwargs, &parse.PairNode{
				Key:   &parse.StringNode{argName, argName.Val},
				Value: expr,
			})
			// stmt.Kwargs[argName.Val] = expr
		} else {
			stmt.Args = append(stmt.Args, argName.Val)
		}

		if args.Match(parse.TokenRparen) != nil {
			break
		}
		if args.Match(parse.TokenComma) == nil {
			errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "unexpected '%s', expected ',' or ')'", args.Current().Val)
		}
	}

	// if args.MatchName("export") != nil {
	// 	stmt.exported = true
	// }

	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(args.Current()), "malformed macro-tag.")
	}

	// Body wrapping
	wrapper, endargs := p.WrapUntil("endmacro")
	stmt.Wrapper = wrapper

	if !endargs.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(endargs.Current()), "arguments not allowed here")
	}

	p.Template.Macros[stmt.Name] = stmt

	// if stmt.exported {
	// 	// Now register the macro if it wants to be exported
	// 	_, has := p.template.exportedMacros[stmt.name]
	// 	if has {
	// 		return nil, p.Error(fmt.Sprintf("another macro with name '%s' already exported", stmt.name), start)
	// 	}
	// 	p.template.exportedMacros[stmt.name] = stmt
	// }

	return &MacroStmt{stmt}
}

func init() {
	All.MustRegister("macro", macroParser)
}
