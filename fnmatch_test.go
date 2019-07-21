package editorconfig

import (
	"fmt"
	"testing"
)

func TestTranslate(t *testing.T) {
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
			result, err := translate(test[0])
			if err != nil {
				t.Fatal(err)
			}
			if result != test[1] {
				t.Errorf("%s != %s", test[1], result)
			}
		})
	}
}
