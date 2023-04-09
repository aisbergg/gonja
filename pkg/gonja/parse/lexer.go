package parse

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// rEOF is an arbitrary value for End Of File
const rEOF = -1

// escapeString is a replacement map for escaping quotes.
var escapedStrings = map[string]string{
	`\"`: `"`,
	`\'`: `'`,
}

// lexFn represents the state of the scanner
// as a function that returns the next state.
type lexFn func() lexFn

// Lexer holds the state of the scanner.
type Lexer struct {
	Input string // the string being scanned.
	Start int    // start position of this item.
	Pos   int    // current position in the input.
	Width int    // width of last rune read from input.
	Line  int    // Current line in the input
	Col   int    // Current position in the line
	// Position Position // Current lexing position in the input
	Config        *Config     // The lexer configuration
	Tokens        chan *Token // channel of scanned tokens.
	delimiters    []rune
	RawStatements rawStmt
	rawEnd        *regexp.Regexp
}

// TODO: set from env
type rawStmt map[string]*regexp.Regexp

// NewLexer creates a new scanner for the input string.
func NewLexer(input string, cfg *Config) *Lexer {
	return &Lexer{
		Input:  input,
		Tokens: make(chan *Token),
		Config: cfg,
		RawStatements: rawStmt{
			"raw":     regexp.MustCompile(fmt.Sprintf(`%s\s*endraw`, cfg.BlockStartString)),
			"comment": regexp.MustCompile(fmt.Sprintf(`%s\s*endcomment`, cfg.BlockStartString)),
		},
	}
}

// Lex lexes the input and returns a stream of tokens.
func Lex(input string, cfg *Config) *Stream {
	l := NewLexer(input, cfg)
	go l.Run()
	return NewStream(l.Tokens)
}

// errorf returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating Lexer.Run.
func (l *Lexer) errorf(format string, args ...any) lexFn {
	l.Tokens <- &Token{
		Type: TokenError,
		Val:  fmt.Sprintf(format, args...),
		Pos:  l.Pos,
	}
	return nil
}

// Position return the current position in the input
func (l *Lexer) Position() *Position {
	return &Position{
		Offset: l.Pos,
		Line:   l.Line,
		Column: l.Col,
	}
}

// Current returns the current token.
func (l *Lexer) Current() string {
	return l.Input[l.Start:l.Pos]
}

// Run lexes the input by executing state functions until the state is nil.
func (l *Lexer) Run() {
	for state := l.lexData; state != nil; {
		state = state()
	}
	close(l.Tokens) // No more tokens will be delivered.
}

// next returns the next rune in the input.
func (l *Lexer) next() (rune rune) {
	if l.Pos >= len(l.Input) {
		l.Width = 0
		return rEOF
	}
	rune, l.Width = utf8.DecodeRuneInString(l.Input[l.Pos:])
	l.Pos += l.Width
	if rune == '\n' {
		l.Line++
		l.Col = 1
	}
	return rune
}

// emit passes a Token back to the client.
func (l *Lexer) emit(t TokenType) {
	l.processAndEmit(t, nil)
}

// processAndEmit processes the current token and emits it.
func (l *Lexer) processAndEmit(t TokenType, fn func(string) string) {
	line, col := ColumnRowFromPos(l.Start, l.Input)
	val := l.Input[l.Start:l.Pos]
	if fn != nil {
		val = fn(val)
	}
	l.Tokens <- &Token{
		Type: t,
		Val:  val,
		Pos:  l.Start,
		Line: line,
		Col:  col,
	}
	l.Start = l.Pos
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.Start = l.Pos
}

// backup steps back one rune. Can be called only once per call of next.
func (l *Lexer) backup() {
	l.Pos -= l.Width
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	} //revive:disable-line:empty-block
	l.backup()
}

// pushDelimiter pushes a delimiter onto the delimiter stack.
func (l *Lexer) pushDelimiter(r rune) {
	l.delimiters = append(l.delimiters, r)
}

// hasPrefix returns true if the remaining input starts with the given prefix.
func (l *Lexer) hasPrefix(prefix string) bool {
	return strings.HasPrefix(l.Input[l.Pos:], prefix)
}

// popDelimiter pops the last pushed delimiter.
func (l *Lexer) popDelimiter(r rune) bool {
	if len(l.delimiters) == 0 {
		l.errorf("Unexpected delimiter '%c'", r)
		return false
	}
	last := len(l.delimiters) - 1
	expected := l.delimiters[last]
	if r != expected {
		l.errorf("Unbalanced delimiters, expected '%c', got '%c'", expected, r)
		return false
	}
	// l.delimiters[last] = nil // Erase element (write zero value)
	l.delimiters = l.delimiters[:last]
	return true
}

// expectDelimiter returns whether or not we are expecting r as the next delimiter.
func (l *Lexer) expectDelimiter(r rune) bool {
	if len(l.delimiters) == 0 {
		return false
	}
	expected := l.delimiters[len(l.delimiters)-1]
	return r == expected
}

// lexData scans until the end of the block or until the end of the input.
func (l *Lexer) lexData() lexFn {
	for {
		if l.hasPrefix(l.Config.CommentStartString) {
			if l.Pos > l.Start {
				l.emit(TokenData)
			}
			return l.lexComment
		}

		if l.hasPrefix(l.Config.VariableStartString) {
			if l.Pos > l.Start {
				l.emit(TokenData)
			}
			return l.lexVariable
		}

		if l.hasPrefix(l.Config.BlockStartString) {
			if l.Pos > l.Start {
				l.emit(TokenData)
			}
			return l.lexBlock
		}

		if l.next() == rEOF {
			break
		}
	}
	// Correctly reached EOF.
	if l.Pos > l.Start {
		l.emit(TokenData)
	}
	l.emit(TokenEOF) // Useful to make EOF a token.
	return nil       // Stop the run loop.
}

// remaining returns the remaining input.
func (l *Lexer) remaining() string {
	return l.Input[l.Pos:]
}

// lexRaw scans a raw statement.
func (l *Lexer) lexRaw() lexFn {
	loc := l.rawEnd.FindStringIndex(l.remaining())
	if loc == nil {
		return l.errorf("unable to find raw closing statement")
	}
	l.Pos += loc[0]
	l.emit(TokenData)
	l.rawEnd = nil
	return l.lexBlock
}

// lexComment scans a comment.
func (l *Lexer) lexComment() lexFn {
	l.Pos += len(l.Config.CommentStartString)
	l.emit(TokenCommentBegin)
	i := strings.Index(l.Input[l.Pos:], l.Config.CommentEndString)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.Pos += i
	l.emit(TokenData)
	l.Pos += len(l.Config.CommentEndString)
	l.emit(TokenCommentEnd)
	return l.lexData
}

// lexVariable scans a variable.
func (l *Lexer) lexVariable() lexFn {
	l.Pos += len(l.Config.VariableStartString)
	l.accept("-")
	l.emit(TokenVariableBegin)
	return l.lexExpression
}

// lexVariableEnd scans the end of a variable.
func (l *Lexer) lexVariableEnd() lexFn {
	l.accept("-")
	l.Pos += len(l.Config.VariableEndString)
	l.emit(TokenVariableEnd)
	return l.lexData
}

// lexBlock lexes a block.
func (l *Lexer) lexBlock() lexFn {
	l.Pos += len(l.Config.BlockStartString)
	l.accept("+-")
	l.emit(TokenBlockBegin)
	for isSpace(l.peek()) {
		l.next()
	}
	if len(l.Current()) > 0 {
		l.emit(TokenWhitespace)
	}
	stmt := l.nextIdentifier()
	l.emit(TokenName)
	re, exists := l.RawStatements[stmt]
	if exists {
		l.rawEnd = re
	}
	return l.lexExpression
}

// lexBlockEnd scans the end of a block.
func (l *Lexer) lexBlockEnd() lexFn {
	l.accept("-")
	l.Pos += len(l.Config.BlockEndString)
	l.emit(TokenBlockEnd)
	if l.rawEnd != nil {
		return l.lexRaw
	}
	return l.lexData
}

// lexExpression scans the next token of the input.
func (l *Lexer) lexExpression() lexFn {
	for {
		if !l.expectDelimiter(l.peek()) {
			if l.hasPrefix(l.Config.VariableEndString) { // && l.expectDelimiter(l.peek()) {
				return l.lexVariableEnd
			}

			// if this is the rightDelim, but we are expecting the next char as a delimiter
			// then skip marking this as rightDelim.  This allows us to have, eg, '}}' as
			// part of a literal inside a var block.
			// if strings.HasPrefix(l.input[l.pos:], l.rightDelim) && !l.shouldExpectDelim(l.peek()) {
			// 	l.pos += Pos(len(l.rightDelim))
			// 	l.emitRight()
			// 	return lexText
			// }

			if l.hasPrefix(l.Config.BlockEndString) {
				return l.lexBlockEnd
			}
		}

		r := l.next()
		// remaining := l.Input[l.Pos:]
		switch {
		case isSpace(r):
			return l.lexSpace
		case isNumeric(r):
			return l.lexNumber
		case isAlphaNumeric(r):
			return l.lexIdentifier
		}

		switch r {
		case '"', '\'':
			l.backup()
			return l.lexString
		case ',':
			l.emit(TokenComma)
		case '|':
			l.emit(TokenPipe)
			// if l.accept("|") {
			// 	l.emit(Or)
			// } else {
			// }
		case '+':
			l.emit(TokenAdd)
		case '-':
			if l.hasPrefix(l.Config.BlockEndString) {
				l.backup()
				return l.lexBlockEnd
			} else if l.hasPrefix(l.Config.VariableEndString) {
				l.backup()
				return l.lexVariableEnd
			} else {
				l.emit(TokenSub)
			}
		case '~':
			l.emit(TokenTilde)
		case ':':
			l.emit(TokenColon)
		case '.':
			l.emit(TokenDot)
		case '%':
			l.emit(TokenMod)
		case '/':
			if l.accept("/") {
				l.emit(TokenFloordiv)
			} else {
				l.emit(TokenDiv)
			}
		case '<':
			if l.accept("=") {
				l.emit(TokenLteq)
			} else {
				l.emit(TokenLt)
			}
		case '>':
			if l.accept("=") {
				l.emit(TokenGteq)
			} else {
				l.emit(TokenGt)
			}
		case '*':
			if l.accept("*") {
				l.emit(TokenPow)
			} else {
				l.emit(TokenMul)
			}
		case '!':
			if l.accept("=") {
				l.emit(TokenNe)
			} else {
				// l.emit(Not)
				l.errorf("Unexpected '!'")
			}
		// case '&':
		// 	if l.accept("&") {
		// 		l.emit(And)
		// 	} else {
		// 		return nil
		// 	}
		case '=':
			if l.accept("=") {
				l.emit(TokenEq)
			} else {
				l.emit(TokenAssign)
			}
		case '(':
			l.emit(TokenLparen)
			l.pushDelimiter(')')
		case '{':
			l.emit(TokenLbrace)
			l.pushDelimiter('}')
		case '[':
			l.emit(TokenLbracket)
			l.pushDelimiter(']')
		case ')':
			if !l.popDelimiter(')') {
				return nil
			}
			l.emit(TokenRparen)
		case '}':
			if !l.popDelimiter('}') {
				return nil
			}
			l.emit(TokenRbrace)
		case ']':
			if !l.popDelimiter(']') {
				return nil
			}
			l.emit(TokenRbracket)
		}
	}
}

// lexSpace scans a run of space characters.
func (l *Lexer) lexSpace() lexFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(TokenWhitespace)
	return l.lexExpression
}

// nextIdentifier advances to the next identifier.
func (l *Lexer) nextIdentifier() string {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			// l.emit(Name)
			return l.Current()
		}
	}
}

// lexIdentifier scans an alphanumeric that serves as an identifier.
func (l *Lexer) lexIdentifier() lexFn {
	l.nextIdentifier()
	l.emit(TokenName)
	return l.lexExpression
}

// lexNumber scans a number: decimal, or float
func (l *Lexer) lexNumber() lexFn {
	tokType := TokenInteger
	for {
		switch r := l.next(); {
		case isNumeric(r):
			// absorb
		case r == '.':
			if tokType != TokenFloat {
				tokType = TokenFloat
			} else {
				l.errorf("two dots in numeric token")
			}
		case isAlphaNumeric(r) && tokType == TokenInteger:
			return l.lexIdentifier
		default:
			l.backup()
			l.emit(tokType)
			return l.lexExpression
		}
	}
}

// unescape unescapes a string that is quoted by `"` or `'`
func unescape(str string) string {
	str = str[1 : len(str)-1]
	for escaped, unescaped := range escapedStrings {
		str = strings.ReplaceAll(str, escaped, unescaped)
	}
	return str
}

// lexString scans a quoted string. The initial quote is already consumed.
func (l *Lexer) lexString() lexFn {
	quote := l.next() // should be either ' or "
	var prev rune
	for r := l.next(); r != quote || prev == '\\'; r, prev = l.next(), r {
	} //revive:disable-line:empty-block
	l.processAndEmit(TokenString, unescape)
	return l.lexExpression
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isNumeric reports whether r is a digit.
func isNumeric(r rune) bool {
	return unicode.IsDigit(r)
}
