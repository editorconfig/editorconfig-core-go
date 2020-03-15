package assert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Equal(t *testing.T, x, y interface{}) {
	r := DiffReporter{}
	if !cmp.Equal(x, y, cmp.Reporter(&r)) {
		t.Error(r.String())
	}
}

func Nil(t *testing.T, x interface{}) {
	Equal(t, x, nil)
}
