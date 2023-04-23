// The MIT License (MIT)
//
// Copyright (c) 2016 lestrrat
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package debug

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const indentPerLevel = 2

type state struct {
	indent int
	mu     sync.RWMutex
	out    io.Writer
}

func (s *state) Indent() int {
	s.mu.RLock()
	indent := s.indent
	s.mu.RUnlock()
	return indent
}

func (s *state) AddIndent(indent int) {
	s.mu.Lock()
	s.indent += indent
	if s.indent < 0 {
		s.indent = 0
	}
	s.mu.Unlock()
}

func (s *state) Write(msg []byte) {
	st.mu.Lock()
	_, _ = st.out.Write(msg)
	st.mu.Unlock()
}

type mGuard struct {
	errPtr *error
	indent int
	name   string
	msg    string
	start  time.Time
}

var st = &state{
	out: os.Stderr,
}

func FuncMarker(prefix string) MarkerGuard {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		panic("log.FuncMarker could not determine the name of caller function")
	}
	f := runtime.FuncForPC(pc)
	return Marker(prefix, f.Name())
}

var mGuardPool = sync.Pool{
	New: allocMGuard,
}

func allocMGuard() any {
	return &mGuard{}
}

func getMGuard() *mGuard {
	return mGuardPool.Get().(*mGuard)
}

func releaseMGuard(mg *mGuard) {
	mg.indent = 0
	mg.msg = ""
	mGuardPool.Put(mg)
}

func Marker(name, format string, args ...any) MarkerGuard {
	mg := getMGuard()
	mg.name = name
	mg.indent = st.Indent()
	mg.msg = fmt.Sprintf(format, args...)

	// format message
	var buf []byte
	buf = fmt.Appendf(buf, "|%s|", mg.name)
	for i := 0; i < mg.indent; i++ {
		buf = append(buf, ' ')
	}
	buf = fmt.Append(buf, "START ", mg.msg)
	buf = append(buf, '\n')

	st.Write(buf)
	st.AddIndent(indentPerLevel)

	return mg
}

func (mg *mGuard) BindError(errPtr *error) MarkerGuard {
	mg.errPtr = errPtr
	return mg
}

func (mg *mGuard) End() {
	// format message
	var buf []byte
	buf = fmt.Appendf(buf, "|%s|", mg.name)
	for i := 0; i < mg.indent; i++ {
		buf = append(buf, ' ')
	}
	buf = fmt.Append(buf, "END   ", mg.msg)
	if mg.errPtr != nil && *mg.errPtr != nil {
		buf = fmt.Appendf(buf, " (error=%s)", *mg.errPtr)
	}
	buf = append(buf, '\n')

	st.Write(buf)
	st.AddIndent(-indentPerLevel)

	releaseMGuard(mg)
}

func Printf(name, format string, args ...any) {
	// format message
	indent := st.Indent()
	var buf []byte
	buf = fmt.Appendf(buf, "|%s|", name)
	for i := 0; i < indent; i++ {
		buf = append(buf, ' ')
	}
	buf = fmt.Appendf(buf, format, args...)
	buf = append(buf, '\n')

	st.Write(buf)
}
