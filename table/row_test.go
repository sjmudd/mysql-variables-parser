package table

import (
	"testing"
)

func TestEmpty(t *testing.T) {
	var got bool
	got = empty(" ")
	if got != true {
		t.Errorf("empty(' ') = %+v, want %+v", got, true)
	}
	got = empty("")
	if got != true {
		t.Errorf("empty('') = %+v, want %+v", got, true)
	}
}
