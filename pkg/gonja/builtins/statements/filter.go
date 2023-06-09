package statements

import (
	"fmt"
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

// FilterStmt is a statement that applies a filter chain to the output of a
// previous statement.
type FilterStmt struct {
	position    *parse.Token
	bodyWrapper *parse.WrapperNode
	filterChain []*parse.FilterCall
}

var (
	_ parse.Statement = (*FilterStmt)(nil)
	_ exec.Statement  = (*FilterStmt)(nil)
)

// Position returns the token position of the statement.
func (stmt *FilterStmt) Position() *parse.Token { return stmt.position }

func (stmt *FilterStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("FilterStmt(Line=%d Col=%d)", t.Line, t.Col)
}

// Execute executes the filter statement.
func (stmt *FilterStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	var out strings.Builder
	sub := r.Inherit()
	sub.Out = &out

	if err := sub.ExecuteWrapper(stmt.bodyWrapper); err != nil {
		// pass error up the call stack
		panic(err)
	}

	value := r.ValueFactory.Value(out.String())
	for _, call := range stmt.filterChain {
		value = r.Evaluator().ExecuteFilter(call, value)
	}
	r.WriteString(value.String())
}

func filterParser(p, args *parse.Parser) parse.Statement {
	stmt := &FilterStmt{
		position: p.Current(),
	}

	wrapper, _ := p.WrapUntil("endfilter")
	stmt.bodyWrapper = wrapper

	for !args.End() {
		filterCall := args.ParseFilter()
		stmt.filterChain = append(stmt.filterChain, filterCall)

		if args.Match(parse.TokenPipe) == nil {
			break
		}
	}

	if !args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "malformed filter-tag args")
	}

	return stmt
}

func init() {
	All.MustRegister("filter", filterParser)
}
