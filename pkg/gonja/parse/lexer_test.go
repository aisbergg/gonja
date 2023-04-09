package parse_test

import (
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type tok struct {
	typ parse.TokenType
	val string
}

func (t tok) String() string {
	return `"` + t.val + `"`
}

var (
	EOF            = tok{parse.TokenEOF, ""}
	varBegin       = tok{parse.TokenVariableBegin, "{{"}
	varEnd         = tok{parse.TokenVariableEnd, "}}"}
	blockBegin     = tok{parse.TokenBlockBegin, "{%"}
	blockBeginTrim = tok{parse.TokenBlockBegin, "{%-"}
	blockEnd       = tok{parse.TokenBlockEnd, "%}"}
	blockEndTrim   = tok{parse.TokenBlockEnd, "-%}"}
	lParen         = tok{parse.TokenLparen, "("}
	rParen         = tok{parse.TokenRparen, ")"}
	lBrace         = tok{parse.TokenLbrace, "{"}
	rBrace         = tok{parse.TokenRbrace, "}"}
	lBracket       = tok{parse.TokenLbracket, "["}
	rBracket       = tok{parse.TokenRbracket, "]"}
	space          = tok{parse.TokenWhitespace, " "}
)

func data(text string) tok {
	return tok{parse.TokenData, text}
}

func name(text string) tok {
	return tok{parse.TokenName, text}
}

func str(text string) tok {
	return tok{parse.TokenString, text}
}

func error(text string) tok {
	return tok{parse.TokenError, text}
}

var lexerCases = []struct {
	name     string
	input    string
	expected []tok
}{
	{"empty", "", []tok{EOF}},
	{"data", "Hello World", []tok{
		data("Hello World"),
		EOF,
	}},
	{"comment", "{# a comment #}", []tok{
		tok{parse.TokenCommentBegin, "{#"},
		data(" a comment "),
		tok{parse.TokenCommentEnd, "#}"},
		EOF,
	}},
	{"mixed comment", "Hello, {# comment #}World", []tok{
		data("Hello, "),
		tok{parse.TokenCommentBegin, "{#"},
		data(" comment "),
		tok{parse.TokenCommentEnd, "#}"},
		data("World"),
		EOF,
	}},
	{"simple variable", "{{ foo }}", []tok{
		varBegin,
		space,
		name("foo"),
		space,
		varEnd,
		EOF,
	}},
	{"basic math expression", "{{ (a - b) + c }}", []tok{
		varBegin, space,
		lParen, name("a"), space, tok{parse.TokenSub, "-"}, space, name("b"), rParen,
		space, tok{parse.TokenAdd, "+"}, space, name("c"),
		space, varEnd,
		EOF,
	}},
	{"blocks", "Hello.  {% if true %}World{% else %}Nobody{% endif %}", []tok{
		data("Hello.  "),
		blockBegin, space, name("if"), space, name("true"), space, blockEnd,
		data("World"),
		blockBegin, space, name("else"), space, blockEnd,
		data("Nobody"),
		blockBegin, space, name("endif"), space, blockEnd,
		EOF,
	}},
	{"blocks with trim control", "Hello.  {%- if true -%}World{%- else -%}Nobody{%- endif -%}", []tok{
		data("Hello.  "),
		blockBeginTrim, space, name("if"), space, name("true"), space, blockEndTrim,
		data("World"),
		blockBeginTrim, space, name("else"), space, blockEndTrim,
		data("Nobody"),
		blockBeginTrim, space, name("endif"), space, blockEndTrim,
		EOF,
	}},
	{"ignore tags in comment", "<html>{# ignore {% tags %} in comments ##}</html>", []tok{
		data("<html>"),
		tok{parse.TokenCommentBegin, "{#"},
		data(" ignore {% tags %} in comments #"),
		tok{parse.TokenCommentEnd, "#}"},
		data("</html>"),
		EOF,
	}},
	{"mixed content", "{# comment #}{% if foo -%} bar {%- elif baz %} bing{%endif    %}", []tok{
		tok{parse.TokenCommentBegin, "{#"},
		data(" comment "),
		tok{parse.TokenCommentEnd, "#}"},
		blockBegin, space, name("if"), space, name("foo"), space, blockEndTrim,
		data(" bar "),
		blockBeginTrim, space, name("elif"), space, name("baz"), space, blockEnd,
		data(" bing"),
		blockBegin, name("endif"), tok{parse.TokenWhitespace, "    "}, blockEnd,
		EOF,
	}},
	{"mixed tokens with doubles", "{{ +--+ /+//,|*/**=>>=<=< == }}", []tok{
		varBegin,
		space,
		tok{parse.TokenAdd, "+"}, tok{parse.TokenSub, "-"}, tok{parse.TokenSub, "-"}, tok{parse.TokenAdd, "+"},
		space,
		tok{parse.TokenDiv, "/"}, tok{parse.TokenAdd, "+"}, tok{parse.TokenFloordiv, "//"},
		tok{parse.TokenComma, ","},
		tok{parse.TokenPipe, "|"},
		tok{parse.TokenMul, "*"},
		tok{parse.TokenDiv, "/"},
		tok{parse.TokenPow, "**"},
		tok{parse.TokenAssign, "="},
		tok{parse.TokenGt, ">"},
		tok{parse.TokenGteq, ">="},
		tok{parse.TokenLteq, "<="},
		tok{parse.TokenLt, "<"},
		space,
		tok{parse.TokenEq, "=="},
		space,
		varEnd,
		EOF,
	}},
	{"delimiters", "{{ ([{}]()) }}", []tok{
		varBegin, space,
		lParen, lBracket, lBrace, rBrace, rBracket, lParen, rParen, rParen,
		space, varEnd,
		EOF,
	}},
	{"unbalanced delimiters", "{{ ([{]) }}", []tok{
		varBegin, space,
		lParen, lBracket, lBrace,
		error("Unbalanced delimiters, expected '}', got ']'"),
	}},
	{"unexpeced delimiter", "{{ ()) }}", []tok{
		varBegin, space,
		lParen, rParen,
		error("Unexpected delimiter ')'"),
	}},
	{"unbalance over end block", "{{ ({a:b, {a:b}}) }}", []tok{
		varBegin, space,
		lParen,
		lBrace, name("a"), tok{parse.TokenColon, ":"}, name("b"), tok{parse.TokenComma, ","},
		space,
		lBrace, name("a"), tok{parse.TokenColon, ":"}, name("b"), rBrace, rBrace,
		rParen,
		space, varEnd,
		EOF,
	}},
	{"string with double quote", `{{ "Hello, " + "World" }}`, []tok{
		varBegin, space,
		str("Hello, "),
		space, tok{parse.TokenAdd, "+"}, space,
		str("World"),
		space, varEnd,
		EOF,
	}},
	{"string with simple quote", `{{ 'Hello, ' + 'World' }}`, []tok{
		varBegin, space,
		str("Hello, "),
		space, tok{parse.TokenAdd, "+"}, space,
		str("World"),
		space, varEnd,
		EOF,
	}},
	{"single quotes inside double quotes string", `{{ "'quoted' test" }}`, []tok{
		varBegin, space, str("'quoted' test"), space, varEnd, EOF,
	}},
	{"escaped string", `{{ "Hello, \"World\"" }}`, []tok{
		varBegin, space,
		str(`Hello, "World"`),
		space, varEnd,
		EOF,
	}},
	{"escaped string mixed", `{{ "Hello,\n \'World\'" }}`, []tok{
		varBegin, space,
		str(`Hello,\n 'World'`),
		space, varEnd,
		EOF,
	}},
	{"if statement", `{% if 5.5 == 5.500000 %}5.5 is 5.500000{% endif %}`, []tok{
		blockBegin, space, name("if"), space,
		tok{parse.TokenFloat, "5.5"}, space, tok{parse.TokenEq, "=="}, space, tok{parse.TokenFloat, "5.500000"},
		space, blockEnd,
		data("5.5 is 5.500000"),
		blockBegin, space, name("endif"), space, blockEnd,
		EOF,
	}},
}

func tokenSlice(c chan *parse.Token) []*parse.Token {
	toks := []*parse.Token{}
	for token := range c {
		toks = append(toks, token)
	}
	return toks
}

func TestLexer(t *testing.T) {
	for _, lc := range lexerCases {
		test := lc
		t.Run(test.name, func(t *testing.T) {
			lexer := parse.NewLexer(test.input, parse.NewConfig())
			go lexer.Run()
			toks := tokenSlice(lexer.Tokens)

			assert := testutils.NewAssert(t)
			assert.Equal(len(test.expected), len(toks))
			actual := []tok{}
			for _, token := range toks {
				actual = append(actual, tok{token.Type, token.Val})
			}
			assert.Equal(test.expected, actual)
		})
	}
}

func streamResult(s *parse.Stream) []tok {
	out := []tok{}
	for !s.End() {
		token := s.Current()
		out = append(out, tok{token.Type, token.Val})
		s.Next()
	}
	return out
}

func asStreamResult(toks []tok) ([]tok, bool) {
	out := []tok{}
	isError := false
	for _, token := range toks {
		if token.typ == parse.TokenError {
			isError = true
			break
		}
		if token.typ != parse.TokenWhitespace && token.typ != parse.TokenEOF {
			out = append(out, token)
		}
	}
	return out, isError
}

func TestLex(t *testing.T) {
	for _, lc := range lexerCases {
		test := lc
		t.Run(test.name, func(t *testing.T) {
			stream := parse.Lex(test.input, parse.NewConfig())
			expected, _ := asStreamResult(test.expected)

			actual := streamResult(stream)

			assert := testutils.NewAssert(t)
			assert.Equal(len(expected), len(actual))
			assert.Equal(expected, actual)
		})
	}
}

func TestStreamSlice(t *testing.T) {
	for _, lc := range lexerCases {
		test := lc
		t.Run(test.name, func(t *testing.T) {
			lexer := parse.NewLexer(test.input, parse.NewConfig())
			go lexer.Run()
			toks := tokenSlice(lexer.Tokens)

			stream := parse.NewStream(toks)
			expected, _ := asStreamResult(test.expected)

			actual := streamResult(stream)

			assert := testutils.NewAssert(t)
			assert.Equal(len(expected), len(actual))
			assert.Equal(expected, actual)
		})
	}
}

const positionsCase = `Hello
{#
    Multiline comment
#}
World
`

func TestLexerPosition(t *testing.T) {
	assert := testutils.NewAssert(t)

	lexer := parse.NewLexer(positionsCase, parse.NewConfig())
	go lexer.Run()
	toks := tokenSlice(lexer.Tokens)
	assert.Equal([]*parse.Token{
		&parse.Token{parse.TokenData, "Hello\n", 0, 1, 1},
		&parse.Token{parse.TokenCommentBegin, "{#", 6, 2, 1},
		&parse.Token{parse.TokenData, "\n    Multiline comment\n", 8, 2, 3},
		&parse.Token{parse.TokenCommentEnd, "#}", 31, 4, 1},
		&parse.Token{parse.TokenData, "\nWorld\n", 33, 4, 3},
		&parse.Token{parse.TokenEOF, "", 40, 6, 1},
	}, toks)
}
