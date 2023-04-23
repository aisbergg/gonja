package parse

import "fmt"

// Stream represents a stream of tokens that can be navigated sequentially. The
// stream allows to access each token one after the other and also goings back
// and forth.
type Stream struct {
	it       TokenIterator
	previous *Token
	current  *Token
	next     *Token
	backup   *Token
	buffer   []*Token
	tokens   []*Token
}

// TokenIterator is an interface for token providers.
type TokenIterator interface {
	Next() *Token
}

// chanIterator is a token iterator that reads from a channel.
type chanIterator struct {
	input chan *Token
}

// ChanIterator creates a new channel iterator.
func ChanIterator(input chan *Token) TokenIterator {
	return &chanIterator{input}
}

// Next returns the next token.
func (ci *chanIterator) Next() *Token {
	return <-ci.input
}

// slicIterator is a token iterator that reads from a slice.
type sliceIterator struct {
	input []*Token
	idx   int
}

// SliceIterator creates a new slice iterator.
func SliceIterator(input []*Token) TokenIterator {
	length := len(input)
	var last *Token
	if length > 0 {
		last = input[length-1]
	}
	if last == nil || last.Type != TokenEOF {
		input = append(input, &Token{Type: TokenEOF})
	}
	return &sliceIterator{input, 0}
}

// Next returns the next token.
func (si *sliceIterator) Next() *Token {
	if si.idx < len(si.input) {
		tok := si.input[si.idx]
		si.idx++
		return tok
	}
	return nil
}

// NewStream creates a new stream.
func NewStream(input any) *Stream {
	var it TokenIterator

	switch t := input.(type) {
	case chan *Token:
		it = ChanIterator(t)
	case []*Token:
		it = SliceIterator(t)
	default:
		panic(fmt.Errorf("[BUG] unsupported stream input type '%T'", t))
	}

	s := &Stream{
		it:     it,
		buffer: []*Token{},
		tokens: []*Token{},
	}
	s.init()
	return s
}

// init initializes the stream.
func (s *Stream) init() {
	s.current = s.nonIgnored()
	if !s.End() {
		s.next = s.nonIgnored()
	}
}

// nonIgnored returns the next non-whitespace token.
func (s *Stream) nonIgnored() *Token {
	var tok *Token
	for tok = s.it.Next(); tok.Type == TokenWhitespace; tok = s.it.Next() {
	} //revive:disable-line:empty-block
	return tok
}

// consume consumes the current token and returns it.
func (s *Stream) consume() *Token {
	s.previous = s.current
	s.current = s.next
	if s.backup != nil {
		s.next = s.backup
		s.backup = nil
	} else if s.End() {
		s.next = nil
	} else {
		s.next = s.nonIgnored()
	}
	return s.previous
}

// Current returns the current token.
func (s *Stream) Current() *Token {
	return s.current
}

// Next returns the next token and consumes it.
func (s *Stream) Next() *Token {
	return s.consume()
}

// EOF returns true if the end of the stream has been reached.
func (s *Stream) EOF() bool {
	return s.current.Type == TokenEOF
}

// IsError returns true if the stream is in an error state.
func (s *Stream) IsError() bool {
	return s.current.Type == TokenError
}

// End returns true if the end of the stream has been reached.
func (s *Stream) End() bool {
	return s.EOF() || s.IsError()
}

// Peek returns the next token without consuming it.
func (s *Stream) Peek() *Token {
	return s.next
}

// Backup backs up the stream to the previous token.
func (s *Stream) Backup() {
	if s.previous == nil {
		panic(fmt.Errorf("[BUG] can't backup"))
	}
	s.backup = s.next
	s.next = s.current
	s.current = s.previous
	s.previous = nil
}
