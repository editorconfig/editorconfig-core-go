package editorconfig //nolint:testpackage

import (
	"testing"
)

func TestTranslate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern  string
		expected string
	}{
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
		{"{a,{b,c}}", `(?:a|(?:b|c))`},
		{"{{a,b},c}", `(?:(?:a|b)|c)`},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			t.Parallel()

			result := translate(test.pattern)
			if result != test.expected {
				t.Errorf("%s != %s", test.expected, result)
			}
		})
	}
}
