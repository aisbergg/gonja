package parse

import (
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
)

// Parser provides the means to parse a template document. The parser works on a
// token list and creates a node tree.
type Parser struct {
	Stream *Stream
	Config *Config

	Template        *TemplateNode
	Statements      map[string]StatementParser
	Level           int8
	TemplateParseFn TemplateParseFn
}

// NewParser creates a new parser for the given token stream.
func NewParser(cfg *Config, stream *Stream) *Parser {
	return &Parser{
		Stream: stream,
		Config: cfg,
	}
}

// Parse is an alias for ParseTemplate.
func (p *Parser) Parse() (*TemplateNode, error) {
	return p.ParseTemplate()
}

// Consume consumes the current token, thereby removing it from the stream.
func (p *Parser) Consume() {
	p.Stream.Next()
}

// Current returns the current token.
func (p *Parser) Current() *Token {
	return p.Stream.Current()
}

// Next returns and consumes the current token.
func (p *Parser) Next() *Token {
	return p.Stream.Next()
}

// End returns the last token in the stream.
func (p *Parser) End() bool {
	return p.Stream.End()
}

// Match returns and consumes the current token if it matches one of the given
// types.
func (p *Parser) Match(types ...TokenType) *Token {
	tok := p.Stream.Current()
	for _, t := range types {
		if tok.Type == t {
			p.Stream.Next()
			return tok
		}
	}
	return nil
}

// MatchName returns and advances the current token if it matches one of the
// given names.
func (p *Parser) MatchName(names ...string) *Token {
	t := p.Peek(TokenName)
	if t != nil {
		for _, name := range names {
			if t.Val == name {
				return p.Pop()
			}
		}
	}
	// if t != nil && t.Val == name { return p.Pop() }
	return nil
}

// Pop returns the current token and advances to the next.
func (p *Parser) Pop() *Token {
	t := p.Stream.Current()
	p.Stream.Next()
	return t
}

// Peek returns the next token without consuming the current one if it matches
// one of the given types.
func (p *Parser) Peek(types ...TokenType) *Token {
	tok := p.Stream.Current()
	for _, t := range types {
		if tok.Type == t {
			return tok
		}
	}
	return nil
}

// PeekName returns the next token without consuming the current one if it
// matches one of the given names.
func (p *Parser) PeekName(names ...string) *Token {
	t := p.Peek(TokenName)
	if t != nil {
		for _, name := range names {
			if t.Val == name {
				return t
			}
		}
	}
	return nil
}

// WrapUntil wraps all nodes between starting tag and "{% endtag %}" and
// provides one simple interface to execute the wrapped  It returns a
// parser to process provided arguments to the tag. Errors are returned as
// panics.
func (p *Parser) WrapUntil(names ...string) (*WrapperNode, *Parser) {
	wrapper := &WrapperNode{
		Location: p.Current(),
		Trim:     &Trim{},
	}

	var args []*Token

	for !p.Stream.End() {
		// New tag, check whether we have to stop wrapping here
		if begin := p.Match(TokenBlockBegin); begin != nil {
			ident := p.Peek(TokenName)

			if ident != nil {
				// We've found a (!) end-tag

				found := false
				for _, n := range names {
					if ident.Val == n {
						found = true
						break
					}
				}

				// We only process the tag if we've found an end tag
				if found {
					// Okay, endtag found.
					p.Consume() // '{%' tagname
					wrapper.Trim.Left = begin.Val[len(begin.Val)-1] == '-'
					wrapper.LStrip = begin.Val[len(begin.Val)-1] == '+'

					for {
						if end := p.Match(TokenBlockEnd); end != nil {
							// Okay, end the wrapping here
							wrapper.EndTag = ident.Val
							wrapper.Trim.Right = end.Val[0] == '-'
							stream := NewStream(args)
							return wrapper, NewParser(p.Config, stream)
						}
						t := p.Next()
						// p.Consume()
						if t == nil {
							errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected EOF")
						}
						args = append(args, t)
					}
				}
			}
			p.Stream.Backup()
		}

		// Otherwise process next element to be wrapped
		node := p.parseDocElement()
		wrapper.Nodes = append(wrapper.Nodes, node)
	}

	errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected EOF, expected any of '%s'", strings.Join(names, " or "))
	return nil, nil
}

// SkipUntil skips all nodes between starting tag and "{% endtag %}". Errors are
// returned as panics.
func (p *Parser) SkipUntil(names ...string) {
	for !p.End() {
		// New tag, check whether we have to stop wrapping here
		if p.Match(TokenBlockBegin) != nil {
			ident := p.Peek(TokenName)

			if ident != nil {
				// We've found a (!) end-tag

				found := false
				for _, n := range names {
					if ident.Val == n {
						found = true
						break
					}
				}

				// We only process the tag if we've found an end tag
				if found {
					// Okay, endtag found.
					p.Consume() // '{%' tagname

					for {
						if p.Match(TokenBlockEnd) != nil {
							// Done skipping, exit.
							return
						}
					}
				}
			} else {
				p.Stream.Backup()
			}
		}
		t := p.Next()
		if t == nil {
			errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected EOF")
		}
	}

	errors.ThrowSyntaxError(p.Current().ErrorToken(), "unexpected EOF, expected any of '%s'", strings.Join(names, " or "))
}

// -----------------------------------------------------------------------------

// Parse parses the given template string and returns the root node of the AST.
func Parse(input string) (*TemplateNode, error) {
	cfg := NewConfig()
	stream := Lex(input, cfg)
	p := NewParser(cfg, stream)
	return p.Parse()
}
