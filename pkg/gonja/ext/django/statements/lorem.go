package statements

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
	"github.com/aisbergg/gonja/pkg/gonja/utils"
)

type LoremStmt struct {
	Location *parse.Token
	count    int    // number of paragraphs
	method   string // w = words, p = HTML paragraphs, b = plain-text (default is b)
	random   bool   // does not use the default paragraph "Lorem ipsum dolor sit amet, ..."
}

func (stmt *LoremStmt) Position() *parse.Token { return stmt.Location }
func (stmt *LoremStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("LoremStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *LoremStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	lorem, err := utils.LoremIpsum(stmt.count, stmt.method)
	if err != nil {
		errors.ThrowTemplateRuntimeError("unable to execute 'lorem' statement: %s", err)
	}
	r.WriteString(lorem)
}

func loremParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &LoremStmt{
		Location: p.Current(),
		count:    1,
		method:   "b",
	}

	if countToken := args.Match(parse.TokenInteger); countToken != nil {
		stmt.count, _ = strconv.Atoi(countToken.Val)
	}

	if methodToken := args.Match(parse.TokenName); methodToken != nil {
		if methodToken.Val != "w" && methodToken.Val != "p" && methodToken.Val != "b" {
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "lorem-method must be either 'w', 'p' or 'b'")
		}

		stmt.method = methodToken.Val
	}

	if args.MatchName("random") != nil {
		stmt.random = true
	}

	if !args.End() {
		// return nil, args.Error("Malformed lorem-tag args.", nil)
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "malformed lorem-tag args")
	}

	return stmt
}

func init() {
	rand.Seed(time.Now().Unix())

	All.MustRegister("lorem", loremParser)
}
