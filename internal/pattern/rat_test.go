package pattern

import "testing"

func TestNewRatNormalizes(t *testing.T) {
	cases := []struct {
		n, d int64
		want Rat
	}{
		{2, 4, Rat{1, 2}},
		{2, -4, Rat{-1, 2}},
		{0, 5, Rat{0, 1}},
		{3, 1, Rat{3, 1}},
	}
	for _, c := range cases {
		if got := NewRat(c.n, c.d); got != c.want {
			t.Errorf("NewRat(%d, %d) = %v, want %v", c.n, c.d, got, c.want)
		}
	}
}
