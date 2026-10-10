package pattern

// Rat is an exact rational number, so subdivisions like thirds never drift.
// Always construct with NewRat so values stay normalized (D > 0, lowest terms).
type Rat struct {
	N, D int64
}

// NewRat returns n/d in lowest terms with a positive denominator.
func NewRat(n, d int64) Rat {
	if d == 0 {
		panic("pattern: zero denominator")
	}
	if d < 0 {
		n, d = -n, -d
	}
	if g := gcd(abs(n), d); g > 1 {
		n, d = n/g, d/g
	}
	return Rat{N: n, D: d}
}

// TODO(M3): Add, Sub, Mul, Div, Cmp, Floor, Ceil, Float64.

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
