package editorconfig_test

import (
	"fmt"
	"testing"

	"github.com/editorconfig/editorconfig-core-go/v2"
)

func TestTranslate(t *testing.T) {
	t.Parallel()

	var tests = [][2]string{
		{"a*e.c", `a[^/]*e\.c`},
		{"a**z.c", `a.*z\.c`},
		{"d/**/z.c", `d(?:/|/.*/)z\.c`},
		{"som?.c", `som[^/]\.c`},
		{"[\\]ab].g", `[\]ab]\.g`},
		{"[ab]].g", `[ab]]\.g`},
		{"ab[/c", `ab\[/c`},
		{"*.{py,js,html}", `[^/]*\.(?:py|js|html)`},
		{"{single}.b", `\{single\}\.b`},
		{"{{,b,c{d}.i", `\{\{,b,c\{d\}\.i`},
		{"{a\\,b,cd}", `(?:a,b|cd)`},
		{"{e,\\},f}", `(?:e|}|f)`},
	}

	for i, test := range tests {
		title := fmt.Sprintf("%d) %s => %s", i, test[0], test[1])
		t.Run(title, func(t *testing.T) {
			t.Parallel()
			result := editorconfig.Translate(test[0]) // nolint: scopelint
			if result != test[1] {                    // nolint: scopelint
				t.Errorf("%s != %s", test[1], result) // nolint: scopelint
			}
		})
	}
}
