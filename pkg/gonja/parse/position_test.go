package parse_test

import (
	"fmt"
	"testing"

	"github.com/aisbergg/gonja/internal/testutils"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

const multilineSample = `Hello
{#
    Multiline comment
#}
World
`

var readablePositionsCases = []struct {
	name string
	pos  int
	line int
	col  int
	char byte
}{
	{"First char", 0, 1, 1, 'H'},
	{"Last char", len(multilineSample) - 1, 5, 6, '\n'},
	{"Anywhere", 13, 3, 5, 'M'},
}

func TestReadablePosition(t *testing.T) {
	for _, rp := range readablePositionsCases {
		test := rp
		t.Run(test.name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(test.char, multilineSample[test.pos],
				`Invalid test, expected "%#U" rune at pos %d, got "%#U"`,
				test.char, test.pos, multilineSample[test.pos])
			line, col := parse.ColumnRowFromPos(test.pos, multilineSample)
			fmt.Println("line", line, "col", col)
			assert.Equal(test.line, line, "Expected line %d, got %d", test.line, line)
			assert.Equal(test.col, col, "Expected col %d, got %d", test.col, col)
		})
	}
}
