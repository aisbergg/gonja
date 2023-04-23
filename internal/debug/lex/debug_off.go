//go:build !(debug || debug_parse)

package lex

import "github.com/aisbergg/gonja/internal/debug"

const Enabled = false

func FuncMarker() debug.MarkerGuard { return debug.NullMGuard{} }
func Print(_ string, _ ...any)      {}
