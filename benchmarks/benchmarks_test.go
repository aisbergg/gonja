package benchmarks

import (
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
	"github.com/aisbergg/gonja/pkg/gonja"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
)

const tplPath = "benchdata/complex.tpl"

var env = gonja.NewEnvironment(
	gonja.OptLoader(gonja.MustFileSystemLoader("")),
	gonja.OptUndefined(gonja.StrictUndefined),
)

func BenchmarkParse(b *testing.B) {
	var (
		tpl = &exec.Template{}
		err error
	)
	for i := 0; i < b.N; i++ {
		tpl, err = env.FromFile(tplPath)
		if err != nil {
			b.Fatal(err)
		}
	}
	_ = tpl
}

func BenchmarkExecute(b *testing.B) {
	tpl, err := env.FromFile(tplPath)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = tpl.Execute(testutils.Fixtures)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelExecute(b *testing.B) {
	tpl, err := env.FromFile(tplPath)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tpl.Execute(testutils.Fixtures)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
