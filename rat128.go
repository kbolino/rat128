// Package rat128 provides fixed-precision rational numbers.
// See the N type and New function for details.
package rat128

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
)

// Common errors returned by functions in this package.
var (
	ErrDenInvalid  = errors.New("denominator is not positive")
	ErrDenOverflow = errors.New("denominator overflow")
	ErrNumOverflow = errors.New("numerator overflow")
	ErrDivByZero   = errors.New("division by zero")
	ErrFmtInvalid  = errors.New("invalid number format")
)

// N is a rational number with 64-bit numerator and denominator.
//
// One bit of the numerator is used for the sign and the denominator must be
// positive, so only 63 bits of precision are actually available in each.
// Internally, the denominator is biased by 1, which means the zero value is
// equivalent to 0/1 and thus valid and equal to 0. Due to the asymmetry of
// the int64 type (|math.MinInt64| > math.MaxInt64), math.MinInt64 is not a
// valid numerator in reduced form.
//
// Valid values are obtained in the following ways:
//   - the zero value of the type N
//   - returned by the New or Try functions (with non-nil error)
//   - returned by arithmetic on any valid values (with non-nil error)
//   - copied from a valid value
//
// N has proper value semantics and its values can be freely copied.
// Two valid values of N can be compared using the == and != operators.
type N struct {
	m int64
	n int64
}

// Try creates a new rational number with the given numerator and denominator.
// Try returns an error if the reduced numerator is math.MinInt64 or if the
// denominator is not positive.
func Try(num, den int64) (N, error) {
	if den <= 0 {
		return N{}, ErrDenInvalid
	}
	return N{num, den - 1}.reduce()
}

// New is like Try but panics instead of returning an error.
func New(num, den int64) N {
	n, err := Try(num, den)
	if err != nil {
		panic(err)
	}
	return n
}

// tryAlreadyReduced is like Try but assumes the numerator and denominator are
// already in reduced form.
func tryAlreadyReduced(num, den int64) (N, error) {
	if den <= 0 {
		return N{}, ErrDenInvalid
	} else if num == math.MinInt64 {
		return N{}, ErrNumOverflow
	}
	return N{num, den - 1}, nil
}

// ParseRationalString parses a string representation of a rational number.
// The string must be in the form "m/n", where m and n are integers in base 10,
// n is not zero, and only m may be negative (indicated with leading hyphen).
// It is not necessary for m/n to be in lowest terms, but the result will be.
// Also, m and n cannot overflow int64.
func ParseRationalString(s string) (N, error) {
	parts := strings.SplitN(s, "/", 3)
	if len(parts) != 2 {
		return N{}, ErrFmtInvalid
	}
	num, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return N{}, fmt.Errorf("parsing numerator: %w", err)
	}
	den, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return N{}, fmt.Errorf("parsing denominator: %w", err)
	}
	return Try(num, den)
}

// ParseDecimalString parses a string representation of a decimal number as a
// rational number. The string must be in the form "A", "A.B", or ".B" where
// A is an integer that may have leading zeroes and may be negative (indicated
// with leading hyphen) and B is an integer that may have trailing zeroes.
// The concatenation of A without leading zeroes and B without trailing zeroes
// must not overflow int64.
func ParseDecimalString(s string) (N, error) {
	neg := false
	firstNonzeroIndex := -1
	lastNonzeroIndex := -1
	dotIndex := -1
	digits := 0
	for i, r := range s {
		switch r {
		case '-':
			if i != 0 {
				return N{}, ErrFmtInvalid
			}
			neg = true
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if firstNonzeroIndex < 0 {
				firstNonzeroIndex = i
			}
			lastNonzeroIndex = i
			fallthrough
		case '0':
			digits++
		case '.':
			if dotIndex >= 0 {
				return N{}, ErrFmtInvalid
			}
			dotIndex = i
			if firstNonzeroIndex < 0 {
				firstNonzeroIndex = i
			}
		default:
			return N{}, ErrFmtInvalid
		}
	}
	if digits == 0 {
		return N{}, ErrFmtInvalid
	}
	if firstNonzeroIndex < 0 {
		return N{}, nil
	}
	if dotIndex >= 0 {
		lastNonzeroIndex = max(lastNonzeroIndex, dotIndex-1)
	} else {
		lastNonzeroIndex = max(lastNonzeroIndex, len(s)-1)
	}
	pow10 := 0
	if dotIndex < 0 {
		pow10 = lastNonzeroIndex - firstNonzeroIndex
	} else if firstNonzeroIndex < dotIndex {
		pow10 = dotIndex - firstNonzeroIndex - 1
	}
	place := New(1, 1)
	ten := New(10, 1)
	for i := 0; i < pow10; i++ {
		var err error
		place, err = place.TryMul(ten)
		if err != nil {
			return N{}, fmt.Errorf("computing pow10(%d): %w", i+1, err)
		}
	}
	var result N
	first := true
	for i := firstNonzeroIndex; i <= lastNonzeroIndex; i++ {
		if i == dotIndex {
			first = false
			continue
		}
		if first {
			first = false
		} else {
			var err error
			place, err = place.TryDiv(ten)
			if err != nil {
				return N{}, fmt.Errorf("updating place for digit at index %d: %w", i, err)
			}
		}
		digit := New(int64(s[i]-'0'), 1)
		placed, err := digit.TryMul(place)
		if err != nil {
			return N{}, fmt.Errorf("placing digit at index %d: %w", i, err)
		}
		result, err = result.TryAdd(placed)
		if err != nil {
			return N{}, fmt.Errorf("adding digit at index %d: %w", i, err)
		}
	}
	if neg {
		result = result.Neg()
	}
	return result, nil
}

// FromFloat64 extracts a rational number from a float64. The result will be
// exactly equal to v, or else an error will be returned.
func FromFloat64(v float64) (N, error) {
	if v == 0 {
		return N{}, nil
	}

	// decompose v such that v = f*2^e with abs(f) in [0.5, 1)
	f, e := math.Frexp(v)

	// convert f to an integer in [2^52, 2^53); m is this integer and
	// s is its original sign
	s := int64(1)
	if f < 0 {
		s = -1
		f = -f
	}
	m := int64(f * 0x1p53)
	e -= 53

	// remove trailing zeros from m and compute its precision (significant
	// figures in base 2)
	tz := bits.TrailingZeros64(uint64(m))
	m >>= tz
	e += tz
	prec := bits.Len64(uint64(m))

	// at this point we have v = m*2^e with m an integer w/o trailing zeroes,
	// so whether v is an integer or not is simply down to e
	if e >= 0 {
		// v is an integer
		if prec+e > 63 {
			// v needs more bits than we have
			return N{}, ErrNumOverflow
		}
		return Try(s*(m<<e), 1)
	}
	// else, v is not an integer

	if e <= -63 {
		// the denominator of v needs more bits than we have
		return N{}, ErrDenOverflow
	}
	return Try(s*m, 1<<-e)
}

// FromBigRat converts a big.Rat to N, if it is possible to do so.
func FromBigRat(r *big.Rat) (N, error) {
	num, den := r.Num(), r.Denom()
	if !num.IsInt64() {
		return N{}, ErrNumOverflow
	} else if !den.IsInt64() {
		return N{}, ErrDenOverflow
	}
	return Try(num.Int64(), den.Int64())
}

// Num returns the numerator of x.
func (x N) Num() int64 {
	return x.m
}

// Den returns the denominator of x.
func (x N) Den() int64 {
	return x.n + 1
}

// IsValid returns true if x is a valid rational number.
// Invalid numbers do not arise under normal circumstances, but may occur if
// a value is constructed or manipulated using unsafe operations.
func (x N) IsValid() bool {
	if x.n < 0 || x.n == math.MaxInt64 {
		return false
	}
	if r, err := x.reduce(); err != nil || x != r {
		return false
	}
	return true
}

// IsZero returns true if x is equal to 0.
func (x N) IsZero() bool {
	return x.m == 0
}

// Sign returns the sign of x: -1 if x < 0, 0 if x == 0, and 1 if x > 0.
func (x N) Sign() int {
	if x.m == 0 {
		return 0
	}
	if x.m < 0 {
		return -1
	}
	return 1
}

// Neg returns the negation of x, -x.
func (x N) Neg() N {
	return N{-x.m, x.n}
}

// TryInv returns the inverse of x, 1/x.
// If x.Num() == 0, TryInv returns (0, ErrDivByZero).
func (x N) TryInv() (N, error) {
	if x.m == 0 {
		return N{}, ErrDivByZero
	}
	sgn := int64(x.Sign())
	return tryAlreadyReduced(sgn*x.Den(), abs64(x.Num()))
}

// Inv is like TryInv but panics instead of returning an error.
func (x N) Inv() N {
	y, err := x.TryInv()
	if err != nil {
		panic(err)
	}
	return y
}

// Abs returns the absolute value of x, |x|.
func (x N) Abs() N {
	return N{abs64(x.m), x.n}
}

// Cmp returns -1 if x < y, 0 if x == y, and 1 if x > y.
func (x N) Cmp(y N) int {
	if x == y {
		return 0
	}
	return x.Sub(y).Sign()
}

// TryAdd adds x and y and returns the result.
// TryAdd returns 0 and a non-nil error if the result would overflow.
func (x N) TryAdd(y N) (N, error) {
	mx, nx := x.Num(), x.Den()
	my, ny := y.Num(), y.Den()

	// Use naive arithmetic if we can.
	// TODO improve overflow check if possible, e.g. using bit-length sums?
	if abs64(mx) < math.MaxInt32 && abs64(my) < math.MaxInt32 && nx < math.MaxInt32 && ny < math.MaxInt32 {
		// Overflow analysis:
		//
		// Define len(x) as the number of bits used to represent abs(x); we
		// can ignore the sign here because it always takes up 1 bit in the
		// result regardless of the operation or the size of the operands.
		//
		// Next, the if statement guarantees us that len(abs(mx)) <= 31,
		// len(abs(my)) <= 31, len(nx) <= 31, and len(ny) <= 31.
		//
		// Therefore, len(mx*ny) <= 62 and len(my*nx) <= 62 since the product
		// of two n-bit numbers takes at most 2*n bits.
		//
		// Finally, len(mx*ny+my*nx) <= 63 since the sum of two n-bit numbers
		// takes at most n+1 bits. Thus, the numerator cannot overflow.
		//
		// It also follows that the denominator cannot overflow, since
		// len(nx*ny) <= 62.
		return Try(mx*ny+my*nx, nx*ny)
	}

	// We can't use simple arithmetic if we've made it this far, because the
	// intermediate values might overflow. Instead, we will use wider ops.
	// But first, let's check the signs to skip unnecessary work.
	s1, s2 := sgn64(mx), sgn64(my)
	if s1 == 0 {
		return y, nil
	} else if s2 == 0 {
		return x, nil
	}

	// Multiply the mx*ny, my*nx, and nx*ny terms with 128-bit precision.
	// From here on out, h is for "high bits" and l is for "low bits".
	m1h, m1l := bits.Mul64(uint64(abs64(mx)), uint64(ny))
	m2h, m2l := bits.Mul64(uint64(abs64(my)), uint64(nx))
	nh, nl := bits.Mul64(uint64(nx), uint64(ny))

	// Compute the full numerator m (mh:ml) with wide arithmetic.
	//
	// There are six cases to consider with respect to the signs and sizes of
	// m1 (m1h:m1l) and m2 (m2h:m2l):
	//
	// - the signs are the same and positive; then       m =   |m1| + |m2|
	// - the signs are the same and negative; then       m = -(|m1| + |m2|)
	// - the signs differ, m1 > 0, and |m1| > |m2|; then m =   |m1| - |m2|
	// - the signs differ, m1 > 0, and |m1| < |m2|; then m = -(|m2| - |m1|)
	// - the signs differ, m1 < 0, and |m1| > |m2|; then m = -(|m1| - |m2|)
	// - the signs differ, m1 < 0, and |m1| < |m2|; then m =   |m2| - |m1|
	var ml, mh uint64
	sgn := int64(1)
	if s1 == s2 {
		if s1 < 0 {
			sgn = -1
		}
		var mlc, mhc uint64 // c is for "carry"
		ml, mlc = bits.Add64(m1l, m2l, 0)
		mh, mhc = bits.Add64(m1h, m2h, mlc)
		if mhc != 0 {
			return N{}, ErrNumOverflow
		}
	} else {
		// m1 < m2
		if s2 > 0 {
			sgn = -sgn
		}
		// |m1| < |m2|
		if m2h > m1h || (m2h == m1h && m2l > m1l) {
			m1h, m2h = m2h, m1h
			m1l, m2l = m2l, m1l
			sgn = -sgn
		}
		var mlb, mhb uint64 // b is for "borrow"
		ml, mlb = bits.Sub64(m1l, m2l, 0)
		mh, mhb = bits.Sub64(m1h, m2h, mlb)
		if mhb != 0 {
			return N{}, ErrNumOverflow
		}
	}

	// Finally, find the GCD of the numerator and denominator and divide it out
	// to reduce the result before the final overflow checks.
	d := uint64(GCD(nx, ny))
	if d <= mh {
		return N{}, ErrNumOverflow
	}
	m, _ := bits.Div64(mh, ml, uint64(d))
	if m > math.MaxInt64 {
		return N{}, ErrNumOverflow
	}
	if d <= nh {
		return N{}, ErrDenOverflow
	}
	n, _ := bits.Div64(nh, nl, uint64(d))
	if n > math.MaxInt64 {
		return N{}, ErrDenOverflow
	}
	return Try(sgn*int64(m), int64(n))
}

// Add adds x and y and returns the result.
// Add panics if the result would overflow.
func (x N) Add(y N) N {
	z, err := x.TryAdd(y)
	if err != nil {
		panic(err)
	}
	return z
}

// TrySub subtracts y from x and returns the result.
// TrySub returns 0 and a non-nil error if the result would overflow.
func (x N) TrySub(y N) (N, error) {
	return x.TryAdd(y.Neg())
}

// Sub subtracts y from x and returns the result.
// The following are equivalent in outcome and behavior:
//
//	x.Sub(y) == x.Add(y.Neg())
func (x N) Sub(y N) N {
	return x.Add(y.Neg())
}

// TryMul multiplies x and y and returns the result.
// TryMul returns 0 and a non-nil error if the result would overflow.
func (x N) TryMul(y N) (N, error) {
	// Compute the sign of the result.
	sgn := int64(x.Sign() * y.Sign())
	if sgn == 0 {
		return N{}, nil
	}
	// We can ignore the operand signs now that we know the result sign, so we
	// work only with absolute values for simplicity.
	mx, nx := abs64(x.Num()), x.Den()
	my, ny := abs64(y.Num()), y.Den()

	// Next, we reduce the fractions by their cross-GCDs to avoid overflow.
	// Even though x and y are already reduced, their product may introduce
	// factors from each that aren't present in the other.
	// Since the result is going to be (mx*my)/(nx*ny), we can divide out
	// GCD(mx, ny) and GCD(my, nx) without changing the value.
	if d := GCD(mx, ny); d != 1 {
		mx, ny = mx/d, ny/d
	}
	if d := GCD(my, nx); d != 1 {
		my, nx = my/d, nx/d
	}

	// Use naive multiplication if we can.
	// TODO improve the overflow check if possible, e.g. len(mx)+len(my)<=63?
	if mx < math.MaxInt32 && my < math.MaxInt32 && nx < math.MaxInt32 && ny < math.MaxInt32 {
		// See Add for a detailed overflow analysis; suffice it to say that
		// the above if statement protects us from overflow here.
		return tryAlreadyReduced(sgn*mx*my, nx*ny)
	}

	// At this point, we can't trust naive multiplication to not overflow, so
	// we use wide arithmetic to check for overflow.
	mh, ml := bits.Mul64(uint64(mx), uint64(my))
	if mh > 0 || ml > math.MaxInt64 {
		return N{}, ErrNumOverflow
	}
	nh, nl := bits.Mul64(uint64(nx), uint64(ny))
	if nh > 0 || nl > math.MaxInt64 {
		return N{}, ErrDenOverflow
	}
	return tryAlreadyReduced(sgn*int64(ml), int64(nl))
}

// Mul multiplies x and y and returns the result.
// Mul panics if the result would overflow.
func (x N) Mul(y N) N {
	z, err := x.TryMul(y)
	if err != nil {
		panic(err)
	}
	return z
}

// TryDiv divides x by y and returns the result.
// TryDiv returns 0 and a non-nil error for division by zero or if the result
// would overflow.
func (x N) TryDiv(y N) (N, error) {
	if inv, err := y.TryInv(); err != nil {
		return N{}, err
	} else {
		return x.TryMul(inv)
	}
}

// Div divides x by y and returns the result.
// The following are equivalent in outcome and behavior:
//
//	x.Div(y) == x.Mul(y.Inv())
func (x N) Div(y N) N {
	return x.Mul(y.Inv())
}

// RationalString returns a string representation of x, as m+sep+n.
// For example, x.String() is equivalent to x.RationalString("/").
func (x N) RationalString(sep string) string {
	return fmt.Sprintf("%d%s%d", x.Num(), sep, x.Den())
}

// String returns a string representation of x, as m/n.
func (x N) String() string {
	return x.RationalString("/")
}

// DecimalString returns a string representation of x, as a decimal number
// to the given number of digits after the decimal point.
// The last digit is rounded to nearest, with ties rounded away from zero.
// If prec <= 0, the decimal point is omitted from the string.
// If the result of rounding is zero but x is negative, the string will still
// include a negative sign.
//
// The following relation should hold for all valid values of x:
//
//	x.DecimalString(prec) == x.BigRat().FloatString(prec)
func (x N) DecimalString(prec int) string {
	if prec < 0 {
		prec = 0
	}
	var buf strings.Builder
	m, n := x.Num(), x.Den()
	// write the negative sign if needed then ensure m is in absolute value
	if m < 0 {
		buf.WriteByte('-')
		m = -m
	}
	// although we have a string builder already, we need a mutable slice to
	// hold the digits, because rounding is done with schoolbook arithmetic
	// and carry over may change every single digit and even prepend a 1;
	// thus we start with a leading zero to make room for it
	digits := []byte{'0'}
	// we start by dividing m over n with remainder; the quotient will be the
	// integer part of the number and the remainder will be the fractional part
	q, r := m/n, m%n
	// we append the integer part and then we will append the decimal digits,
	// one by one without the decimal point; we will put it in later
	digits = strconv.AppendInt(digits, q, 10)
	// going out to prec+1 gives us an extra digit for rounding
	for i := 0; i < prec+1; i++ {
		if r == 0 {
			digits = append(digits, '0')
			continue
		}
		// now we multiply the remainder by 10 to extract another decimal
		// digit, then re-divide by 10 to get a new quotient and remainder for
		// the next iteration
		if r < math.MaxInt64/10 {
			// use ordinary arithmetic if we can
			r *= 10
			q, r = r/n, r%n
		} else {
			// r is too large so we have to use wide arithmetic to avoid
			// overflow; this gives us (rh:rl) <= MaxInt64*10, which is
			// (4:18446744073709551606) according to big.Int
			rh, rl := bits.Mul64(uint64(r), 10)
			// we know that we got here because r >= MaxInt64/10 and moreover
			// that r is a remainder of division by n, so n > r, thus
			// n > MaxInt64/10 > rh and therefore Div64 won't panic
			quo, rem := bits.Div64(rh, rl, uint64(n))
			// quo < 10 and rem < n <= MaxInt64 so int64 cast is safe
			q, r = int64(quo), int64(rem)
		}
		digits = append(digits, byte(q)+'0')
	}
	// use digit in last position to round
	if k := len(digits) - 1; digits[k] >= '5' {
		digits[k-1]++
		for i := k - 1; i >= 0; i-- {
			if digits[i] <= '9' {
				break
			}
			digits[i] = '0'
			digits[i-1]++
		}
	}
	// prepare the digit string to be written out
	start := 0
	end := len(digits) - 1
	if digits[0] == '0' {
		// skip the leading zero if we didn't use it
		start = 1
	}
	if prec > 0 {
		// shift the digits after the decimal point by 1 to make room for it;
		// we have space for it because we kept an extra digit for rounding
		dotIndex := len(digits) - prec - 1
		for i := len(digits) - 1; i > dotIndex; i-- {
			digits[i] = digits[i-1]
		}
		digits[dotIndex] = '.'
		end = len(digits)
	}
	buf.Write(digits[start:end])
	// this may return "-0" etc. which could be filtered out but agrees with
	// the output of big.Rat.FloatString
	return buf.String()
}

// Float64 returns the floating-point equivalent of x. If exact is true, then
// v is exactly equal to x; otherwise, it is the closest approximation.
func (x N) Float64() (v float64, exact bool) {
	m, n := x.Num(), x.Den()

	// check for zero, trivial case
	if m == 0 {
		return 0, true
	}

	// integers are exact as long as they fit in the mantissa
	prec := bits.Len64(uint64(abs64(m)))
	if n == 1 {
		return float64(m), prec <= 53
	}

	// non-integers are exact as long as the numerator fits in the mantissa
	// and the denominator is a power of two
	nIsPow2 := bits.OnesCount64(uint64(n)) == 1
	return float64(m) / float64(n), prec <= 53 && nIsPow2
}

// BigRat converts x to a new big.Rat.
func (x N) BigRat() *big.Rat {
	return big.NewRat(x.Num(), x.Den())
}

// reduce returns x in lowest terms.
func (x N) reduce() (N, error) {
	if x.m == 0 {
		return N{}, nil
	}
	sgn := int64(x.Sign())
	m, n := abs64(x.Num()), x.Den()
	d := GCD(m, n)
	num := sgn * (m / d)
	if num == math.MinInt64 {
		return N{}, ErrNumOverflow
	}
	den := (n / d) - 1
	return N{num, den}, nil
}

// abs64 returns the absolute value of x.
// WARNING: abs64(math.MinInt64) == math.MinInt64 < 0.
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// sgn64 returns -1 if x < 0, 0 if x == 0, and 1 if x > 0.
func sgn64(x int64) int64 {
	if x == 0 {
		return 0
	}
	if x < 0 {
		return -1
	}
	return 1
}
