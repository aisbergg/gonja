package parse

import (
	"fmt"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// TokenType identifies the type of token.
type TokenType int

// Known tokens
const (
	TokenError TokenType = iota
	TokenAdd
	TokenAssign
	TokenColon
	TokenComma
	TokenDiv
	TokenDot
	TokenEq
	TokenFloordiv
	TokenGt
	TokenGteq
	TokenLbrace
	TokenLbracket
	TokenLparen
	TokenLt
	TokenLteq
	TokenMod
	TokenMul
	TokenNe
	TokenPipe
	TokenPow
	TokenRbrace
	TokenRbracket
	TokenRparen
	TokenSemicolon
	TokenSub
	TokenTilde
	TokenWhitespace
	TokenFloat
	TokenInteger
	TokenName
	TokenString
	TokenOperator
	TokenBlockBegin
	TokenBlockEnd
	TokenVariableBegin
	TokenVariableEnd
	TokenRawBegin
	TokenRawEnd
	TokenCommentBegin
	TokenCommentEnd
	TokenComment
	TokenLinestatementBegin
	TokenLinestatementEnd
	TokenLinecommentBegin
	TokenLinecommentEnd
	TokenLinecomment
	TokenData
	TokenInitial
	TokenEOF
)

func (t TokenType) String() string {
	names := map[TokenType]string{
		TokenError:              "Error",
		TokenAdd:                "Add",
		TokenAssign:             "Assign",
		TokenColon:              "Colon",
		TokenComma:              "Comma",
		TokenDiv:                "Div",
		TokenDot:                "Dot",
		TokenEq:                 "Eq",
		TokenFloordiv:           "Floordiv",
		TokenGt:                 "Gt",
		TokenGteq:               "Gteq",
		TokenLbrace:             "Lbrace",
		TokenLbracket:           "Lbracket",
		TokenLparen:             "Lparen",
		TokenLt:                 "Lt",
		TokenLteq:               "Lteq",
		TokenMod:                "Mod",
		TokenMul:                "Mul",
		TokenNe:                 "Ne",
		TokenPipe:               "Pipe",
		TokenPow:                "Pow",
		TokenRbrace:             "Rbrace",
		TokenRbracket:           "Rbracket",
		TokenRparen:             "Rparen",
		TokenSemicolon:          "Semicolon",
		TokenSub:                "Sub",
		TokenTilde:              "Tilde",
		TokenWhitespace:         "Whitespace",
		TokenFloat:              "Float",
		TokenInteger:            "Integer",
		TokenName:               "Name",
		TokenString:             "String",
		TokenOperator:           "Operator",
		TokenBlockBegin:         "BlockBegin",
		TokenBlockEnd:           "BlockEnd",
		TokenVariableBegin:      "VariableBegin",
		TokenVariableEnd:        "VariableEnd",
		TokenRawBegin:           "RawBegin",
		TokenRawEnd:             "RawEnd",
		TokenCommentBegin:       "CommentBegin",
		TokenCommentEnd:         "CommentEnd",
		TokenComment:            "Comment",
		TokenLinestatementBegin: "LinestatementBegin",
		TokenLinestatementEnd:   "LinestatementEnd",
		TokenLinecommentBegin:   "LinecommentBegin",
		TokenLinecommentEnd:     "LinecommentEnd",
		TokenLinecomment:        "Linecomment",
		TokenData:               "Data",
		TokenInitial:            "Initial",
		TokenEOF:                "EOF",
	}
	if name, ok := names[t]; ok {
		return name
	}

	return "Unknown"
}

// Token represents a unit of lexing
type Token struct {
	Type TokenType
	Val  string
	Pos  int
	Line int
	Col  int
}

func (t Token) String() string {
	val := t.Val
	if len(val) > 1000 {
		val = fmt.Sprintf("%s...%s", val[:10], val[len(val)-5:])
	}
	return fmt.Sprintf("<Token[%s] Val='%s' Pos=%d Line=%d Col=%d>", t.Type, val, t.Pos, t.Line, t.Col)
}

// ErrorToken converts the Token into an [errors.Token].
func (t Token) ErrorToken() *errors.Token {
	return &errors.Token{
		Val:  t.Val,
		Pos:  t.Pos,
		Line: t.Line,
		Col:  t.Col,
	}
}
