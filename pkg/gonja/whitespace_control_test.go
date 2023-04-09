package gonja_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	gonja "github.com/aisbergg/gonja/pkg/gonja"
	"github.com/pmezard/go-difflib/difflib"

	tu "github.com/aisbergg/gonja/pkg/gonja/testutils"
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

const source = "testdata/whitespaces/source.tpl"
const result = "testdata/whitespaces/%s.out"

func TestWhiteSpace(t *testing.T) {
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			env := gonja.NewEnvironment(
				gonja.DefaultLoader,
				gonja.TrimBlocks(),
				gonja.LstripBlocks(),
				gonja.KeepTrailingNewline(),
			)

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
				diff := difflib.UnifiedDiff{
					A:        difflib.SplitLines(string(expected)),
					B:        difflib.SplitLines(string(rendered)),
					FromFile: "Expected",
					ToFile:   "Rendered",
					Context:  2,
					Eol:      "\n",
				}
				result, _ := difflib.GetUnifiedDiffString(diff)
				t.Errorf("%s rendered with diff:\n%v", source, result)
			}
		})
	}
}
