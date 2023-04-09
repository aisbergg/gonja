package parse

import (
	"fmt"
	"strings"
)

// Pos is an interface that wraps the Pos method.
type Pos interface {
	Pos() int
}

// Position describes an arbitrary source position including the file, line, and
// column location. A Position is valid if the line number is > 0.
type Position struct {
	Filename string // filename, if any
	Offset   int    // offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (byte count)
}

// IsValid reports whether the position is valid.
func (pos *Position) IsValid() bool { return pos.Line > 0 }

// Pos returns the current offset starting at 0.
func (pos *Position) Pos() int { return pos.Offset }

// String returns a string in one of several forms:
//
//	file:line:column    valid position with file name
//	file:line           valid position with file name but no column (column == 0)
//	line:column         valid position without file name
//	line                valid position without file name and no column (column == 0)
//	file                invalid position with file name
//	-                   invalid position without file name
func (pos Position) String() string {
	s := pos.Filename
	if pos.IsValid() {
		if s != "" {
			s += ":"
		}
		s += fmt.Sprintf("%d", pos.Line)
		if pos.Column != 0 {
			s += fmt.Sprintf(":%d", pos.Column)
		}
	}
	if s == "" {
		s = "-"
	}
	return s
}

// ColumnRowFromPos returns the column and row for a given character offset of
// an input string.
func ColumnRowFromPos(pos int, input string) (column, row int) {
	before := input[:pos]
	lines := strings.Split(before, "\n")
	length := len(lines)
	return length, len(lines[length-1]) + 1
}
