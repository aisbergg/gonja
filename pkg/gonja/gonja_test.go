package gonja_test

import (
	"flag"
	"os"
	"testing"

	. "github.com/go-check/check"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

func TestMain(m *testing.M) {
	flag.Parse()
	code := m.Run()
	os.Exit(code)
}
