package time_test

import (
	"testing"

	arrow "github.com/bmuller/arrow/lib"

	"github.com/aisbergg/gonja/pkg/gonja"
	"github.com/aisbergg/gonja/pkg/gonja/ext/time"
	tu "github.com/aisbergg/gonja/pkg/gonja/testutils"
)

func Env(root string) *gonja.Environment {
	env := tu.TestEnv(root)
	env.Statements.Update(time.Statements)
	cfg := time.NewConfig()
	parsed, _ := arrow.CParse("%Y-%m-%d %H:%M:%S", "1984-06-07 16:40:00")
	cfg.Now = &parsed
	env.ExtensionConfig["time"] = cfg
	return env
}

func TestTimeStatement(t *testing.T) {
	root := "./testdata"
	env := Env(root)
	tu.GlobTemplateTests(t, root, env)
}
