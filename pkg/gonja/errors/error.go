package errors

import "fmt"

// TemplateError is a generic template error.
type TemplateError interface {
	error
	TemplateError()
}

// Token is a token representation for error reporting.
type Token struct {
	Val  string
	Pos  int
	Line int
	Col  int
}

// String returns a string representation of the token.
func (t Token) String() string {
	val := t.Val
	if len(val) > 1000 {
		val = fmt.Sprintf("%s...%s", val[:10], val[len(val)-5:])
	}
	return fmt.Sprintf("<Token Val='%s' Pos=%d Line=%d Col=%d>", val, t.Pos, t.Line, t.Col)
}
