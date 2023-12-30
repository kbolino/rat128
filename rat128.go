// Package rat128 provides fixed-precision rational numbers.
// See the N type and New function for details.
package rat128

import (
	"errors"
	"fmt"
	"math"
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
	ErrFmtInvalid  = errors.New("invalid rational number format")
)

// N is a rational number with 64-bit numerator and denominator.
//
// One bit of the numerator is used for the sign and the denominator must be
// positive, so only 63 bits of precision are actually available in each.
// Internally, the denominator is biased by 1, which means the zero value is
// equivalent to 0/1 and thus valid and equal to 0.
//
// Valid values are obtained in the following ways:
//   - the zero value of the type N
//   - returned by the New function
//   - returned by arithmetic on any valid values
//   - copied from a valid value
//
// N has proper value semantics and its values can be freely copied.
// Two valid values of N can be compared using the == and != operators.
type N struct {
	m int64
	n int64
}

// Try creates a new rational number with the given numerator and denominator.
// Try returns an error if the denominator is not positive.
func Try(num, den int64) (N, error) {
	if den <= 0 {
		return N{}, ErrDenInvalid
	}
	return N{num, den - 1}.reduce(), nil
}

// New is like Try but panics if the denominator is not positive.
func New(num, den int64) N {
	n, err := Try(num, den)
	if err != nil {
		panic(err)
	}
	return n
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
		return New(s*(m<<e), 1), nil
	}
	// else, v is not an integer

	if e <= -63 {
		// the denominator of v needs more bits than we have
		return N{}, ErrDenOverflow
	}
	return New(s*m, 1<<-e), nil
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
	return x.n >= 0 && x.n != math.MaxInt64 && x.reduce() == x
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

// Inv returns the inverse of x, 1/x.
func (x N) Inv() N {
	if x.m == 0 {
		panic(ErrDivByZero)
	}
	sgn := int64(x.Sign())
	return New(sgn*x.Den(), abs64(x.Num()))
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

// Add adds x and y and returns the result.
// Add panics if the result would overflow.
func (x N) Add(y N) N {
	m1, n1 := x.Num(), x.Den()
	m2, n2 := y.Num(), y.Den()
	// TODO detect overflow more precisely
	if abs64(m1) < math.MaxInt32 && abs64(m2) < math.MaxInt32 && n1 < math.MaxInt32 && n2 < math.MaxInt32 {
		return New(m1*n2+m2*n1, n1*n2)
	}
	s1, s2 := sgn64(m1), sgn64(m2)
	if s1 == 0 {
		return y
	} else if s2 == 0 {
		return x
	}
	mh1, ml1 := bits.Mul64(uint64(abs64(m1)), uint64(n2))
	mh2, ml2 := bits.Mul64(uint64(abs64(m2)), uint64(n1))
	nh, nl := bits.Mul64(uint64(n1), uint64(n2))
	var ml, mh uint64
	sgn := int64(1)
	if s1 == s2 {
		if s1 < 0 {
			sgn = -1
		}
		var mlc, mhc uint64
		ml, mlc = bits.Add64(ml1, ml2, 0)
		mh, mhc = bits.Add64(mh1, mh2, mlc)
		if mhc != 0 {
			panic(ErrNumOverflow)
		}
	} else {
		// m1 < m2
		if s2 > 0 {
			sgn = -sgn
		}
		// |m1| < |m2|
		if mh2 > mh1 || (mh2 == mh1 && ml2 > ml1) {
			mh1, mh2 = mh2, mh1
			ml1, ml2 = ml2, ml1
			sgn = -sgn
		}
		var mlb, mhb uint64
		ml, mlb = bits.Sub64(ml1, ml2, 0)
		mh, mhb = bits.Sub64(mh1, mh2, mlb)
		if mhb != 0 {
			panic(ErrNumOverflow)
		}
	}
	d := GCD(n1, n2)
	n, _ := bits.Div64(nh, nl, uint64(d))
	if n > math.MaxInt64 {
		panic(ErrDenOverflow)
	}
	m, _ := bits.Div64(mh, ml, uint64(d))
	if m > math.MaxInt64 {
		panic(ErrNumOverflow)
	}
	return New(sgn*int64(m), int64(n))
}

// Sub subtracts y from x and returns the result.
// The following are equivalent in outcome and behavior:
//
//	x.Sub(y) == x.Add(y.Neg())
func (x N) Sub(y N) N {
	return x.Add(y.Neg())
}

// Mul multiplies x and y and returns the result.
// Mul panics if the result would overflow.
func (x N) Mul(y N) N {
	sgn := int64(x.Sign() * y.Sign())
	if sgn == 0 {
		return N{}
	}
	m1, n1 := abs64(x.Num()), x.Den()
	m2, n2 := abs64(y.Num()), y.Den()
	d12, d21 := GCD(m1, n2), GCD(m2, n1)
	if d12 != 1 {
		m1, n2 = m1/d12, n2/d12
	}
	if d21 != 1 {
		m2, n1 = m2/d21, n1/d21
	}
	if m1 < math.MaxInt32 && m2 < math.MaxInt32 && n1 < math.MaxInt32 && n2 < math.MaxInt32 {
		return New(sgn*m1*m2, n1*n2)
	}
	mh, ml := bits.Mul64(uint64(m1), uint64(m2))
	if mh > 0 || ml > math.MaxInt64 {
		panic(ErrNumOverflow)
	}
	nh, nl := bits.Mul64(uint64(n1), uint64(n2))
	if nh > 0 || nl > math.MaxInt64 {
		panic(ErrDenOverflow)
	}
	return New(sgn*int64(ml), int64(nl))
}

// Div divides x by y and returns the result.
// The following are equivalent in outcome and behavior:
//
//	x.Div(y) == x.Mul(y.Inv())
func (x N) Div(y N) N {
	return x.Mul(y.Inv())
}

// String returns a string representation of x, as m/n.
func (x N) String() string {
	return fmt.Sprintf("%d/%d", x.Num(), x.Den())
}

// Float64 returns the floating-point equivalent of x. If exact is true, then
// value is exactly equal to x; otherwise, it is the closest approximation.
func (x N) Float64() (v float64, exact bool) {
	m, n := x.Num(), x.Den()
	if m == 0 {
		return 0, true
	}
	prec := bits.Len64(uint64(abs64(m)))
	if n == 1 {
		return float64(m), prec <= 53
	}
	nIsPow2 := bits.OnesCount64(uint64(n)) == 1
	return float64(m) / float64(n), prec <= 53 && nIsPow2
}

func (x N) reduce() N {
	if x.m == 0 {
		return N{}
	}
	sgn := int64(x.Sign())
	m, n := abs64(x.Num()), x.Den()
	d := GCD(m, n)
	return N{sgn * (m / d), (n / d) - 1}
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func sgn64(x int64) int64 {
	if x == 0 {
		return 0
	}
	if x < 0 {
		return -1
	}
	return 1
}
