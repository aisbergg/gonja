package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type cycleValue struct {
	node  *CycleStatement
	value exec.Value
}

type CycleStatement struct {
	position *parse.Token
	args     []parse.Expression
	idx      int
	asName   string
	silent   bool
}

var _ parse.Statement = (*CycleStatement)(nil)
var _ exec.Statement = (*CycleStatement)(nil)

func (stmt *CycleStatement) Position() *parse.Token { return stmt.position }
func (stmt *CycleStatement) String() string {
	t := stmt.Position()
	return fmt.Sprintf("CycleStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (cv *cycleValue) String() string {
	return cv.value.String()
}

func (stmt *CycleStatement) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	item := stmt.args[stmt.idx%len(stmt.args)]
	stmt.idx++

	val := r.Eval(item)
	if t, ok := val.Interface().(*cycleValue); ok {
		// {% cycle "test1" "test2"
		// {% cycle cycleitem %}

		// Update the cycle value with next value
		item := t.node.args[t.node.idx%len(t.node.args)]
		t.node.idx++

		t.value = r.Eval(item)
		if !t.node.silent {
			r.WriteString(val.String())
		}

	} else {
		// Regular call
		cycleValue := &cycleValue{
			node:  stmt,
			value: val,
		}

		if stmt.asName != "" {
			r.Ctx.Set(stmt.asName, cycleValue)
		}
		if !stmt.silent {
			r.WriteString(val.String())
		}
	}
}

// HINT: We're not supporting the old comma-separated list of expressions argument-style
func cycleParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	cycleNode := &CycleStatement{
		position: p.Current(),
	}

	for !args.End() {
		node := args.ParseExpression()
		cycleNode.args = append(cycleNode.args, node)

		if args.MatchName("as") != nil {
			// as

			name := args.Match(parse.TokenName)
			if name == nil {
				errors.ThrowSyntaxError(p.Current().ErrorToken(), "name (identifier) expected after 'as'")
			}
			cycleNode.asName = name.Val

			if args.MatchName("silent") != nil {
				cycleNode.silent = true
			}

			// Now we're finished
			break
		}
	}

	if !args.End() {
		errors.ThrowSyntaxError(p.Current().ErrorToken(), "malformed cycle-tag")
	}

	return cycleNode
}

func init() {
	All.MustRegister("cycle", cycleParser)
}
