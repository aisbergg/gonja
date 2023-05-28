package errors

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

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

// stack is an array of stack frames stored in a human readable format.
type stack []stackFrame

func (s stack) String() string {
	builder := strings.Builder{}
	// pre-allocate a large buffer to avoid reallocations; some guesswork here:
	// Name: 40 per error
	// Location: 160 per error
	builder.Grow(len(s) * (40 + 160))
	for _, f := range s {
		builder.WriteString(f.Name)
		builder.WriteString("\n        ")
		builder.WriteString(f.File)
		builder.WriteRune(':')
		builder.WriteString(strconv.Itoa(f.Line))
		builder.WriteString(" pc=0x")
		builder.WriteString(strconv.FormatInt(int64(f.ProgramCounter), 16)) // format as hex
		builder.WriteRune('\n')
	}
	return builder.String()
}

// stackFrame stores a frame's runtime information in a human readable format.
type stackFrame struct {
	// Name of the function.
	Name string
	// File path where the function is defined.
	File string
	// Line number where the function is defined.
	Line int
	// ProgramCounter is the underlying program counter for the function.
	ProgramCounter uintptr
}

// getStackTrace returns a stack trace.
func getStackTrace(skip int) string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(int(2+skip), pcs)
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	stack := make(stack, 0, n)
	for {
		frame, more := frames.Next()
		stack = append(stack, stackFrame{
			Name:           frame.Function,
			File:           frame.File,
			Line:           frame.Line,
			ProgramCounter: frame.PC,
		})
		if !more {
			break
		}
	}
	return stack.String()
}
