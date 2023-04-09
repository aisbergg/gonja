//go:build debug || debug_parse

package lex

import "github.com/aisbergg/gonja/internal/log"

const (
	Enabled = true
	name    = "LEX"
)

func FuncMarker() log.MarkerGuard {
	return log.FuncMarker(name)
}

func Print(format string, args ...any) {
	log.Printf(name, format, args...)
}
