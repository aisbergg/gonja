package gonja_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aisbergg/gonja/internal/diff"
	gonja "github.com/aisbergg/gonja/pkg/gonja"

	tu "github.com/aisbergg/gonja/internal/testutils"
)

var testCases = []struct {
	name                string
	trimBlocks          bool
	lstripBlocks        bool
	keepTrailingNewline bool
}{
	{"default", false, false, false},
	{"trim_blocks", true, false, false},
	{"lstrip_blocks", false, true, false},
	{"keep_trailing_newline", false, false, true},
	{"all", true, true, true},
}

const (
	source = "testdata/whitespaces/source.tpl"
	result = "testdata/whitespaces/%s.out"
)

func TestWhiteSpace(t *testing.T) {
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			options := []gonja.Option{
				gonja.OptLoader(gonja.MustFileSystemLoader("")),
			}
			if test.trimBlocks {
				options = append(options, gonja.OptTrimBlocks())
			}
			if test.lstripBlocks {
				options = append(options, gonja.OptLstripBlocks())
			}
			if test.keepTrailingNewline {
				options = append(options, gonja.OptKeepTrailingNewline())
			}
			env := gonja.NewEnvironment(options...)

			tpl, err := env.FromFile(source)
			if err != nil {
				t.Fatalf("Error on FromFile('%s'): %s", source, err.Error())
			}
			output := fmt.Sprintf(result, test.name)
			expected, rerr := ioutil.ReadFile(output)
			if rerr != nil {
				t.Fatalf("Error on ReadFile('%s'): %s", output, rerr.Error())
			}
			rendered, err := tpl.ExecuteBytes(tu.Fixtures)
			if err != nil {
				t.Fatalf("Error on Execute('%s'): %s", source, err.Error())
			}
			// rendered = testTemplateFixes.fixIfNeeded(match, rendered)
			if !bytes.Equal(expected, rendered) {
				d, err := diff.Diff([]byte(expected), []byte(rendered))
				if err != nil {
					t.Fatalf("failed to compute diff for %s:\n%s", source, err.Error())
				}
				fmt.Println(string(expected))
				fmt.Println("-----------------")
				fmt.Println(string(rendered))
				t.Errorf("%s rendered with diff:\n%s", source, string(d))
			}
		})
	}
}
