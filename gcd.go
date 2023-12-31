package rat128

// GCD returns the greatest common denominator (GCD) of m and n.
// The GCD is the largest integer that divides both m and n.
func GCD(m, n int64) int64 {
	// there are other algorithms, but ExtGCD took 2 to 11 ns/op for a wide
	// range of m and n on an AMD Ryzen 5600X so it is probably fast enough
	_, _, d := ExtGCD(m, n)
	return d
}

// ExtGCD returns the GCD of m and n along with the Bézout coefficients.
// That is, it returns a, b, d such that:
//
//	a*m + b*n == d == GCD(m, n)
func ExtGCD(m, n int64) (a, b, d int64) {
	// per Donald Knuth, TAOCP Vol 1 (3e), pp 13-14, Algorithm E
	var a0, b0 int64
	a0, a = 1, 0
	b0, b = 0, 1
	c := m
	d = n
	for {
		q, r := c/d, c%d
		if r == 0 {
			return a, b, d
		}
		c = d
		d = r
		t := a0
		a0 = a
		a = t - q*a
		t = b0
		b0 = b
		b = t - q*b
	}
}
