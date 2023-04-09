//go:build !(debug || debug_parse)

package exec

import "github.com/aisbergg/gonja/internal/log"

const Enabled = false

func FuncMarker() log.MarkerGuard { return log.NullMGuard{} }
func Print(_ string, _ ...any)    {}
