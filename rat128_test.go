package rat128_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/kbolino/rat128"
)

// some distinct primes satisfying both P_M*P_N > 2^32 and P_K*P_M*P_N < 2^64,
// for all K, M, N
const (
	P1 = 92821
	P2 = 92831
	P3 = 92849
	P4 = 92857
)

var New = rat128.New
var Zero rat128.N

type ArithCase struct {
	X, Y, Z rat128.N
	Err     error
}

func TestN_TryAdd(t *testing.T) {
	cases := []ArithCase{
		{New(1, 1), New(1, 1), New(2, 1), nil},
		{New(-1, 1), New(1, 1), New(0, 1), nil},
		{New(1, 1), New(-1, 1), New(0, 1), nil},
		{New(-1, 1), New(-1, 1), New(-2, 1), nil},
		{New(1, 2), New(1, 2), New(1, 1), nil},
		{New(-1, 2), New(1, 2), New(0, 1), nil},
		{New(1, 2), New(-1, 2), New(0, 1), nil},
		{New(-1, 2), New(-1, 2), New(-1, 1), nil},
		{New(1, 2), New(1, 4), New(3, 4), nil},
		{New(-1, 2), New(1, 4), New(-1, 4), nil},
		{New(7, 11*13), New(11, 7*13), New(7*7+11*11, 7*11*13), nil},
		{New(P1, P2*P3), New(P2, P1*P3), New(P1*P1+P2*P2, P1*P2*P3), nil},
		{New(-P1, P2*P3), New(P2, P1*P3), New(P2*P2-P1*P1, P1*P2*P3), nil},
		{New(P1, P2*P3), New(-P2, P1*P3), New(P1*P1-P2*P2, P1*P2*P3), nil},
		{New(-P1, P2*P3), New(-P2, P1*P3), New(-(P1*P1 + P2*P2), P1*P2*P3), nil},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("(%s)+(%s)", c.X, c.Y), func(t *testing.T) {
			z, err := c.X.TryAdd(c.Y)
			if err != c.Err {
				t.Errorf("got error %v, want %v", err, c.Err)
			} else if c.Err == nil && z != c.Z {
				t.Errorf("got %v, want %v", z, c.Z)
			}
		})
	}
}

func TestN_TryMul(t *testing.T) {
	cases := []ArithCase{
		{New(1, 1), New(1, 1), New(1, 1), nil},
		{New(-1, 1), New(1, 1), New(-1, 1), nil},
		{New(1, 1), New(-1, 1), New(-1, 1), nil},
		{New(-1, 1), New(-1, 1), New(1, 1), nil},
		{New(1, 2), New(1, 2), New(1, 4), nil},
		{New(-1, 2), New(1, 2), New(-1, 4), nil},
		{New(1, 2), New(-1, 2), New(-1, 4), nil},
		{New(-1, 2), New(-1, 2), New(1, 4), nil},
		{New(1, 2), New(1, 4), New(1, 8), nil},
		{New(7, 11*13), New(11, 7*13), New(1, 13*13), nil},
		{New(P1, P2*P3), New(P2, P1*P3), New(1, P3*P3), nil},
		{New(-P1, P2*P3), New(P2, P1*P3), New(-1, P3*P3), nil},
		{New(P1, P2*P3), New(-P2, P1*P3), New(-1, P3*P3), nil},
		{New(-P1, P2*P3), New(-P2, P1*P3), New(1, P3*P3), nil},
		{New(P1*P2, P3), New(P3, P4), New(P1*P2, P4), nil},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("(%s)*(%s)", c.X, c.Y), func(t *testing.T) {
			z, err := c.X.TryMul(c.Y)
			if err != c.Err {
				t.Errorf("got error %v, want %v", err, c.Err)
			} else if c.Err == nil && z != c.Z {
				t.Errorf("got %v, want %v", z, c.Z)
			}
		})
	}
}

func TestFromFloat64(t *testing.T) {
	cases := []struct {
		Float float64
		Rat   rat128.N
		Err   error
	}{
		{0, New(0, 1), nil},
		{1, New(1, 1), nil},
		{-1, New(-1, 1), nil},
		{0.5, New(1, 2), nil},
		{0.25, New(1, 4), nil},
		{0.375, New(3, 8), nil},
		{12.375, New(99, 8), nil},
		{123, New(123, 1), nil},
		{0x1p53, New(1<<53, 1), nil},
		{0x1p-53, New(1, 1<<53), nil},
		{0x1.Fp53, New(0x1F<<49, 1), nil},
		{0x1.Fp0, New(0x1F, 1<<4), nil},
		{0x1.FEDC_BA98_7654_3p52, New(0x1F_EDCB_A987_6543, 1), nil},
		{0x1.FEDC_BA98_7654_3p62, New(0x1F_EDCB_A987_6543<<10, 1), nil},
		{0x1.FEDC_BA98_7654_3p63, Zero, rat128.ErrNumOverflow},
		{0x1.FEDC_BA98_7654_3p-10, New(0x1F_EDCB_A987_6543, 1<<62), nil},
		{0x1.FEDC_BA98_7654_3p-11, Zero, rat128.ErrDenOverflow},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%g", c.Float), func(t *testing.T) {
			r, err := rat128.FromFloat64(c.Float)
			if err != c.Err {
				t.Fatalf("got error %v, want %v", err, c.Err)
			}
			if c.Err == nil && r != c.Rat {
				t.Errorf("got value %s, want %s", r, c.Rat)
			}
		})
	}
}

func TestN_Float64(t *testing.T) {
	cases := []struct {
		Rat   rat128.N
		Float float64
		Exact bool
	}{
		{New(0, 1), 0, true},
		{New(1, 1), 1, true},
		{New(-1, 1), -1, true},
		{New(1, 2), 0.5, true},
		{New(-1, 2), -0.5, true},
		{New(1, 5), 0.2, false},
		{New(-1, 5), -0.2, false},
		{New(1, 9), 0.111_111_111_111_111_111, false},
		{New(2, 3), 0.666_666_666_666_666_666, false},
		{New(-2, 3), -0.666_666_666_666_666_666, false},
		{New(1, 7), 0.142_857_142_857_142_857, false},
		{New(1<<63-1, 1), 9.223_372_036_854_775_807e18, false},
	}
	for _, c := range cases {
		t.Run(c.Rat.String(), func(t *testing.T) {
			f, exact := c.Rat.Float64()
			if exact != c.Exact {
				t.Errorf("got exact=%v, want %v", exact, c.Exact)
			}
			if exact {
				if f != c.Float {
					t.Errorf("got %g, want %g", f, c.Float)
				}
			} else {
				next := math.Nextafter(c.Float, math.Inf(1))
				prev := math.Nextafter(c.Float, math.Inf(-1))
				if f > next || f < prev {
					t.Errorf("got %g, want ~%g", f, c.Float)
				}
			}
		})
	}
}

func TestParseDecimalString(t *testing.T) {
	cases := []struct {
		String string
		Rat    rat128.N
		IsErr  bool
	}{
		{"0", New(0, 1), false},
		{"1.0", New(1, 1), false},
		{"1.23", New(123, 100), false},
		{"123.0", New(123, 1), false},
		{"1234567890.", New(1234567890, 1), false},
		{"123456789.0", New(123456789, 1), false},
		{"12345678.90", New(123456789, 10), false},
		{"1234567.890", New(123456789, 100), false},
		{"123456.7890", New(123456789, 1000), false},
		{"12345.67890", New(123456789, 10_000), false},
		{"1234.567890", New(123456789, 100_000), false},
		{"123.4567890", New(123456789, 1_000_000), false},
		{"12.34567890", New(123456789, 10_000_000), false},
		{"1.234567890", New(123456789, 100_000_000), false},
		{".1234567890", New(123456789, 1_000_000_000), false},
		{".01234567890", New(123456789, 10_000_000_000), false},
		{".001234567890", New(123456789, 100_000_000_000), false},
		{"", Zero, true},
		{" ", Zero, true},
		{".", Zero, true},
		{"-", Zero, true},
		{"a", Zero, true},
		{"1234567890123456789012345678901234567890", Zero, true},
		{"1234567890123456789.012345678901234567890", Zero, true},
		{".1234567890123456789012345678901234567890", Zero, true},
		{"0000000000000000000000000000000000000000", New(0, 1), false},
		{"0000000000000000000000000000000000000001", New(1, 1), false},
		{"1.000000000000000000000000000000000000000", New(1, 1), false},
		{"1000000000000000000000000000000000000001", Zero, true},
		{"1.000000000000000000000000000000000000001", Zero, true},
		{"000000000000000000000000000000000000000101", New(101, 1), false},
		{"1.010000000000000000000000000000000000000", New(101, 100), false},
		{"0.000001010000000000000000000000000000000", New(101, 100_000_000), false},
	}
	for _, c := range cases {
		t.Run(c.String, func(t *testing.T) {
			r, err := rat128.ParseDecimalString(c.String)
			if !c.IsErr {
				if err != nil {
					t.Fatalf("got unexpected error %v", err)
				}
				if r != c.Rat {
					t.Errorf("got value %s, want %s", r, c.Rat)
				}
			} else {
				if err == nil {
					t.Fatalf("got no error, want one")
				}
			}
		})
	}
}

func TestN_DecimalString(t *testing.T) {
	cases := []struct {
		Rat    rat128.N
		Prec   int
		String string
	}{
		{New(0, 1), -1, "0"},
		{New(0, 1), 0, "0"},
		{New(0, 1), 1, "0.0"},
		{New(0, 1), 2, "0.00"},
		{New(0, 1), 3, "0.000"},
		{New(1, 1), -1, "1"},
		{New(1, 1), 0, "1"},
		{New(1, 1), 1, "1.0"},
		{New(1, 1), 2, "1.00"},
		{New(1, 1), 3, "1.000"},
		{New(-1, 1), 0, "-1"},
		{New(-1, 1), 1, "-1.0"},
		{New(-1, 1), 2, "-1.00"},
		{New(-1, 1), 3, "-1.000"},
		{New(1, 2), -1, "1"},
		{New(1, 2), 0, "1"},
		{New(1, 2), 1, "0.5"},
		{New(1, 2), 2, "0.50"},
		{New(1, 2), 3, "0.500"},
		{New(-1, 2), 0, "-1"},
		{New(-1, 2), 1, "-0.5"},
		{New(-1, 2), 2, "-0.50"},
		{New(-1, 2), 3, "-0.500"},
		{New(3, 4), -1, "1"},
		{New(3, 4), 0, "1"},
		{New(3, 4), 1, "0.8"},
		{New(3, 4), 2, "0.75"},
		{New(3, 4), 3, "0.750"},
		{New(1, 3), 0, "0"},
		{New(1, 3), 1, "0.3"},
		{New(1, 3), 2, "0.33"},
		{New(1, 3), 3, "0.333"},
		{New(-1, 3), 0, "-0"}, // matches big.Rat.FloatString
		{New(-1, 3), 1, "-0.3"},
		{New(-1, 3), 2, "-0.33"},
		{New(-1, 3), 3, "-0.333"},
		{New(2, 3), 0, "1"},
		{New(2, 3), 1, "0.7"},
		{New(2, 3), 2, "0.67"},
		{New(2, 3), 3, "0.667"},
		{New(-2, 3), 0, "-1"},
		{New(-2, 3), 1, "-0.7"},
		{New(-2, 3), 2, "-0.67"},
		{New(-2, 3), 3, "-0.667"},
		{New(4, 3), 0, "1"},
		{New(4, 3), 1, "1.3"},
		{New(4, 3), 2, "1.33"},
		{New(4, 3), 3, "1.333"},
		{New(-4, 3), 0, "-1"},
		{New(-4, 3), 1, "-1.3"},
		{New(-4, 3), 2, "-1.33"},
		{New(-4, 3), 3, "-1.333"},
		{New(76, 7), 0, "11"},
		{New(76, 7), 1, "10.9"},
		{New(76, 7), 2, "10.86"},
		{New(76, 7), 3, "10.857"},
		{New(76, 7), 4, "10.8571"},
		{New(76, 7), 5, "10.85714"},
		{New(1<<63-1, 2), 1, "4611686018427387903.5"},

		// the following reference values were obtained from big.Rat.FloatString
		{New(1<<63-2, 1<<63-1), 18, "1.000000000000000000"},
		{New(1<<63-2, 1<<63-1), 19, "0.9999999999999999999"},
		{New(1<<63-2, 1<<63-1), 20, "0.99999999999999999989"},
		{New(1<<63-2, 1<<63-1), 21, "0.999999999999999999892"},
		{New(1<<63-2, 1<<63-1), 22, "0.9999999999999999998916"},
		{New(1<<63-2, 1<<63-1), 23, "0.99999999999999999989158"},
		{New(1<<63-2, 1<<63-1), 24, "0.999999999999999999891580"},
		{New(1<<63-2, 1<<63-1), 25, "0.9999999999999999998915798"},
	}
	for _, c := range cases {
		r := c.Rat
		t.Run(fmt.Sprintf("(%s):%d", r, c.Prec), func(t *testing.T) {
			s := r.DecimalString(c.Prec)
			if s != c.String {
				t.Errorf("got %s, want %s", s, c.String)
			}
		})
	}
}
