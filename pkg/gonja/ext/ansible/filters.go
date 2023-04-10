package ansible

import (
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

// Filters is a set of filters that are available in the Ansible Jinja2
// implementation.
var Filters = exec.FilterSet{
	"type_debug": filterTypeDebug,
}

func filterTypeDebug(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	return exec.AsValue(in.Val.Type().String())
}
