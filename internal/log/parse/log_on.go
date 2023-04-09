//go:build debug || debug_parse

package parse

import "github.com/aisbergg/gonja/internal/log"

const (
	Enabled = true
	name    = "PAR"
)

func FuncMarker() log.MarkerGuard {
	return log.FuncMarker(name)
}

func Print(format string, args ...any) {
	log.Printf(name, format, args...)
}
