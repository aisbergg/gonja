package gonja

import "github.com/aisbergg/gonja/pkg/gonja/exec"

var (
	Undefined       exec.UndefinedFunc = exec.NewUndefinedValue
	StrictUndefined exec.UndefinedFunc = exec.NewStrictUndefinedValue
)
