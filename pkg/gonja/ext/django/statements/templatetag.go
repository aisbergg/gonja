package statements

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type TemplateTagStmt struct {
	Location *parse.Token
	content  string
}

func (stmt *TemplateTagStmt) Position() *parse.Token { return stmt.Location }
func (stmt *TemplateTagStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("TemplateTagStmt(Line=%d Col=%d)", t.Line, t.Col)
}

var templateTagMapping = map[string]string{
	"openblock":     "{%",
	"closeblock":    "%}",
	"openvariable":  "{{",
	"closevariable": "}}",
	"openbrace":     "{",
	"closebrace":    "}",
	"opencomment":   "{#",
	"closecomment":  "#}",
}

func (stmt *TemplateTagStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.WriteString(stmt.content)
}

func templateTagParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &TemplateTagStmt{}

	if argToken := args.Match(parse.TokenName); argToken != nil {
		output, found := templateTagMapping[argToken.Val]
		if !found {
			errors.ThrowSyntaxError(argToken.ErrorToken(), "argument not found")
		}
		stmt.content = output
	} else {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "identifier expected")
	}

	if !args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "malformed templatetag-tag argument")
	}

	return stmt
}

func init() {
	All.MustRegister("templatetag", templateTagParser)
}
