package testutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/aisbergg/gonja/internal/diff"
	"github.com/aisbergg/gonja/pkg/gonja"
	"github.com/aisbergg/gonja/pkg/gonja/loaders"

	u "github.com/aisbergg/gonja/pkg/gonja/utils"
)

func TestEnv(root string) *gonja.Environment {
	env := gonja.NewEnvironment(
		gonja.OptLoader(loaders.MustNewFileSystemLoader(root)),
		gonja.OptKeepTrailingNewline(),
		gonja.OptAutoescape(),
		gonja.OptSetGlobal("lorem", u.LoremIpsum), // Predictable random content
	)
	return env
}

func GlobTemplateTests(t *testing.T, root string, env *gonja.Environment) {
	pattern := filepath.Join(root, `*.tpl`)
	matches, err := filepath.Glob(pattern)
	// env := TestEnv(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range matches {
		filename, err := filepath.Rel(root, match)
		if err != nil {
			t.Fatalf("unable to compute path from `%s`:\n%s", match, err.Error())
		}
		testName := strings.Replace(path.Base(match), ".tpl", "", 1)
		t.Run(testName, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()

			rand.Seed(42) // Make tests deterministic

			tpl, err := env.FromFile(filename)
			if err != nil {
				t.Fatalf("Error on FromFile('%s'):\n%s", filename, err.Error())
			}
			testFilename := fmt.Sprintf("%s.out", match)
			expected, rerr := ioutil.ReadFile(testFilename)
			if rerr != nil {
				t.Fatalf("Error on ReadFile('%s'):\n%s", testFilename, rerr.Error())
			}
			rendered, err := tpl.ExecuteBytes(Fixtures)
			if err != nil {
				t.Fatalf("Error on Execute('%s'):\n%s", filename, err.Error())
			}
			// rendered = testTemplateFixes.fixIfNeeded(filename, rendered)
			if !bytes.Equal(expected, rendered) {
				d, err := diff.Diff([]byte(expected), []byte(rendered))
				if err != nil {
					t.Fatalf("failed to compute diff for %s:\n%s", testFilename, err.Error())
				}
				t.Errorf("%s rendered with diff:\n%v", testFilename, d)
			}
		})
	}
}

func GlobErrorTests(t *testing.T, root string) {
	pattern := filepath.Join(root, `*.err`)
	matches, err := filepath.Glob(pattern)
	env := TestEnv(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range matches {
		testName := strings.Replace(path.Base(match), ".err", "", 1)
		t.Run(testName, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()

			testdata, _ := ioutil.ReadFile(match)
			tests := strings.Split(string(testdata), "\n")

			checkFilename := fmt.Sprintf("%s.out", match)
			checkData, err := ioutil.ReadFile(checkFilename)
			if err != nil {
				t.Fatalf("Error on ReadFile('%s'):\n%s", checkFilename, err.Error())
			}
			checks := strings.Split(string(checkData), "\n")

			if len(checks) != len(tests) {
				t.Fatal("Template lines != Checks lines")
			}

			for idx, test := range tests {
				if strings.TrimSpace(test) == "" {
					continue
				}
				if strings.TrimSpace(checks[idx]) == "" {
					t.Fatalf("[%s Line %d] Check is empty (must contain an regular expression).",
						match, idx+1)
				}

				_, err := env.FromString(test)
				if err != nil {
					t.Fatalf("Error on FromString('%s'):\n%s", test, err.Error())
				}

				tpl, err := env.FromBytes([]byte(test))
				if err != nil {
					t.Fatalf("Error on FromBytes('%s'):\n%s", test, err.Error())
				}

				_, err = tpl.ExecuteBytes(Fixtures)
				if err == nil {
					t.Fatalf("[%s Line %d] Expected error for (got none): %s",
						match, idx+1, tests[idx])
				}

				re := regexp.MustCompile(fmt.Sprintf("^%s$", checks[idx]))
				if !re.MatchString(err.Error()) {
					t.Fatalf("[%s Line %d] Error for '%s' (err = '%s') does not match the (regexp-)check: %s",
						match, idx+1, test, err.Error(), checks[idx])
				}
			}
		})
	}
}
