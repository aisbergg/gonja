package statements

import (
	"fmt"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type BlockStmt struct {
	Location *parse.Token
	Name     string
}

var _ parse.Statement = (*BlockStmt)(nil)
var _ exec.Statement = (*BlockStmt)(nil)

func (stmt *BlockStmt) Position() *parse.Token { return stmt.Location }
func (stmt *BlockStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("BlockStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *BlockStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	blocks := r.Root.GetBlocks(stmt.Name)
	block, blocks := blocks[0], blocks[1:]

	if block == nil {
		errors.ThrowTemplateRuntimeError("unable to find block '%s'", stmt.Name)
	}

	// XXX: missing the context of this template
	sub := r.Inherit()
	infos := &BlockInfos{Block: stmt, Renderer: sub, Blocks: blocks}

	sub.Ctx.Set("super", infos.super)
	sub.Ctx.Set("self", exec.Self(sub))

	if err := sub.ExecuteWrapper(block); err != nil {
		// pass error up the call stack
		panic(err)
	}
}

type BlockInfos struct {
	Block    *BlockStmt
	Renderer *exec.Renderer
	Blocks   []*parse.WrapperNode
	Root     *parse.TemplateNode
}

func (bi *BlockInfos) super() string {
	if len(bi.Blocks) <= 0 {
		errors.ThrowTemplateRuntimeError("failed to render block: super() can only be used in child templates")
	}
	r := bi.Renderer
	block, blocks := bi.Blocks[0], bi.Blocks[1:]
	sub := r.Inherit()
	var out strings.Builder
	sub.Out = &out
	infos := &BlockInfos{
		Block:    bi.Block,
		Renderer: sub,
		Blocks:   blocks,
	}
	sub.Ctx.Set("self", exec.Self(sub))
	sub.Ctx.Set("super", infos.super)
	if err := sub.ExecuteWrapper(block); err != nil {
		// pass error up the call stack
		panic(err)
	}
	return out.String()
}

func blockParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	block := &BlockStmt{
		Location: p.Current(),
	}
	if args.End() {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "tag 'block' requires an identifier")
	}

	name := args.Match(parse.TokenName)
	if name == nil {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "first argument for tag 'block' must be an identifier")
	}

	if !args.End() {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "tag 'block' takes exactly 1 argument (an identifier)")
	}

	wrapper, endargs := p.WrapUntil("endblock")
	if !endargs.End() {
		endName := endargs.Match(parse.TokenName)
		if endName != nil {
			if endName.Val != name.Val {
				errors.ThrowSyntaxError(p.Current().ErrorToken(), "name for 'endblock' must equal to 'block'-tag's name ('%s' != '%s').",
					name.Val, endName.Val)
			}
		}

		if endName == nil || !endargs.End() {
			errors.ThrowSyntaxError(p.Current().ErrorToken(), "either no or only one argument (identifier) allowed for 'endblock'")
		}
	}

	if !p.Template.Blocks.Exists(name.Val) {
		if err := p.Template.Blocks.Register(name.Val, wrapper); err != nil {
			errors.ThrowSyntaxError(block.Location.ErrorToken(), "failed to register block named '%s': %s", name.Val, err)
		}
	} else {
		errors.ThrowSyntaxError(block.Location.ErrorToken(), "block named '%s' already defined", name.Val)
	}

	block.Name = name.Val
	return block
}

func init() {
	All.MustRegister("block", blockParser)
}
