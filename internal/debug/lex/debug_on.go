//go:build debug || debug_parse

package lex

import "github.com/aisbergg/gonja/internal/debug"

const (
	Enabled = true
	name    = "LEX"
)

func FuncMarker() debug.MarkerGuard {
	return debug.FuncMarker(name)
}

func Print(format string, args ...any) {
	debug.Printf(name, format, args...)
}
