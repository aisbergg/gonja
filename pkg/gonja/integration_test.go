package gonja_test

import (
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
)

func TestTemplates(t *testing.T) {
	// Add a global to the default set
	root := "./testdata"
	env := testutils.TestEnv(root)
	env.Globals["this_is_a_global_variable"] = "this is a global text"
	testutils.GlobTemplateTests(t, root, env)
}

func TestExpressions(t *testing.T) {
	root := "./testdata/expressions"
	env := testutils.TestEnv(root)
	testutils.GlobTemplateTests(t, root, env)
}

func TestFilters(t *testing.T) {
	root := "./testdata/filters"
	env := testutils.TestEnv(root)
	testutils.GlobTemplateTests(t, root, env)
}

func TestFunctions(t *testing.T) {
	root := "./testdata/functions"
	env := testutils.TestEnv(root)
	testutils.GlobTemplateTests(t, root, env)
}

func TestTests(t *testing.T) {
	root := "./testdata/tests"
	env := testutils.TestEnv(root)
	testutils.GlobTemplateTests(t, root, env)
}

func TestStatements(t *testing.T) {
	root := "./testdata/statements"
	env := testutils.TestEnv(root)
	testutils.GlobTemplateTests(t, root, env)
}

func TestCompilationErrors(t *testing.T) {
	root := "./testdata/errors/compilation"
	env := testutils.TestEnv(root)
	testutils.GlobTemplateTests(t, root, env)
}

func TestExecutionErrors(t *testing.T) {
	root := "./testdata/errors/execution"
	env := testutils.TestEnv(root)
	testutils.GlobTemplateTests(t, root, env)
}
