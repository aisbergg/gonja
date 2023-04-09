//go:build integration
// +build integration

package gonja_test

import (
	"testing"

	tu "github.com/aisbergg/gonja/pkg/gonja/testutils"
)

func TestTemplates(t *testing.T) {
	// Add a global to the default set
	root := "./testdata"
	env := tu.TestEnv(root)
	env.Globals["this_is_a_global_variable"] = "this is a global text"
	tu.GlobTemplateTests(t, root, env)
}

func TestExpressions(t *testing.T) {
	root := "./testdata/expressions"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

func TestFilters(t *testing.T) {
	root := "./testdata/filters"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

func TestFunctions(t *testing.T) {
	root := "./testdata/functions"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

func TestTests(t *testing.T) {
	root := "./testdata/tests"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

func TestStatements(t *testing.T) {
	root := "./testdata/statements"
	env := tu.TestEnv(root)
	tu.GlobTemplateTests(t, root, env)
}

// func TestCompilationErrors(t *testing.T) {
// 	tu.GlobErrorTests(t, "./testdata/errors/compilation")
// }

// func TestExecutionErrors(t *testing.T) {
// 	tu.GlobErrorTests(t, "./testdata/errors/execution")
// }
