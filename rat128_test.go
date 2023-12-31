package rat128_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/kbolino/rat128"
)

// some distinct primes satisfying both pM*pN > 2^32 and pK*pM*pN < 2^64,
// for all K, M, N
const (
	p1 = 92821
	p2 = 92831
	p3 = 92849
	p4 = 92857
)

type r128 [2]int64

func (r r128) Rat128() rat128.N {
	return rat128.New(r[0], r[1])
}

// x, y, z := arithCase[0], arithCase[1], arithCase[2]
// x OP y == z
type arithCase [3]r128

func (c arithCase) x() rat128.N {
	return c[0].Rat128()
}

func (c arithCase) y() rat128.N {
	return c[1].Rat128()
}

func (c arithCase) z() rat128.N {
	return c[2].Rat128()
}

func TestN_Add(t *testing.T) {
	cases := []arithCase{
		{{1, 1}, {1, 1}, {2, 1}},
		{{-1, 1}, {1, 1}, {0, 1}},
		{{1, 1}, {-1, 1}, {0, 1}},
		{{-1, 1}, {-1, 1}, {-2, 1}},
		{{1, 2}, {1, 2}, {1, 1}},
		{{-1, 2}, {1, 2}, {0, 1}},
		{{1, 2}, {-1, 2}, {0, 1}},
		{{-1, 2}, {-1, 2}, {-1, 1}},
		{{1, 2}, {1, 4}, {3, 4}},
		{{-1, 2}, {1, 4}, {-1, 4}},
		{{7, 11 * 13}, {11, 7 * 13}, {7*7 + 11*11, 7 * 11 * 13}},
		{{p1, p2 * p3}, {p2, p1 * p3}, {p1*p1 + p2*p2, p1 * p2 * p3}},
		{{-p1, p2 * p3}, {p2, p1 * p3}, {p2*p2 - p1*p1, p1 * p2 * p3}},
		{{p1, p2 * p3}, {-p2, p1 * p3}, {p1*p1 - p2*p2, p1 * p2 * p3}},
		{{-p1, p2 * p3}, {-p2, p1 * p3}, {-(p1*p1 + p2*p2), p1 * p2 * p3}},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("(%s)+(%s)", c.x(), c.y()), func(t *testing.T) {
			z := c.x().Add(c.y())
			if z != c.z() {
				t.Errorf("got %v, want %v", z, c.z())
			}
		})
	}
}

func TestN_Mul(t *testing.T) {
	cases := []arithCase{
		{{1, 1}, {1, 1}, {1, 1}},
		{{-1, 1}, {1, 1}, {-1, 1}},
		{{1, 1}, {-1, 1}, {-1, 1}},
		{{-1, 1}, {-1, 1}, {1, 1}},
		{{1, 2}, {1, 2}, {1, 4}},
		{{-1, 2}, {1, 2}, {-1, 4}},
		{{1, 2}, {-1, 2}, {-1, 4}},
		{{-1, 2}, {-1, 2}, {1, 4}},
		{{1, 2}, {1, 4}, {1, 8}},
		{{7, 11 * 13}, {11, 7 * 13}, {1, 13 * 13}},
		{{p1, p2 * p3}, {p2, p1 * p3}, {1, p3 * p3}},
		{{-p1, p2 * p3}, {p2, p1 * p3}, {-1, p3 * p3}},
		{{p1, p2 * p3}, {-p2, p1 * p3}, {-1, p3 * p3}},
		{{-p1, p2 * p3}, {-p2, p1 * p3}, {1, p3 * p3}},
		{{p1 * p2, p3}, {p3, p4}, {p1 * p2, p4}},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("(%s)*(%s)", c.x(), c.y()), func(t *testing.T) {
			z := c.x().Mul(c.y())
			if z != c.z() {
				t.Errorf("got %v, want %v", z, c.z())
			}
		})
	}
}

func TestFromFloat64(t *testing.T) {
	cases := []struct {
		f float64
		r r128
		e error
	}{
		{0, r128{0, 1}, nil},
		{1, r128{1, 1}, nil},
		{-1, r128{-1, 1}, nil},
		{0.5, r128{1, 2}, nil},
		{0.25, r128{1, 4}, nil},
		{0.375, r128{3, 8}, nil},
		{12.375, r128{99, 8}, nil},
		{123, r128{123, 1}, nil},
		{0x1p53, r128{1 << 53, 1}, nil},
		{0x1p-53, r128{1, 1 << 53}, nil},
		{0x1.Fp53, r128{0x1F << 49, 1}, nil},
		{0x1.Fp0, r128{0x1F, 1 << 4}, nil},
		{0x1.FEDC_BA98_7654_3p52, r128{0x1F_EDCB_A987_6543, 1}, nil},
		{0x1.FEDC_BA98_7654_3p62, r128{0x1F_EDCB_A987_6543 << 10, 1}, nil},
		{0x1.FEDC_BA98_7654_3p63, r128{}, rat128.ErrNumOverflow},
		{0x1.FEDC_BA98_7654_3p-10, r128{0x1F_EDCB_A987_6543, 1 << 62}, nil},
		{0x1.FEDC_BA98_7654_3p-11, r128{}, rat128.ErrDenOverflow},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%g", c.f), func(t *testing.T) {
			r, err := rat128.FromFloat64(c.f)
			if err != c.e {
				t.Fatalf("got error %v, want %v", err, c.e)
			}
			if c.e == nil && r != c.r.Rat128() {
				t.Errorf("got value %s, want %s", r, c.r.Rat128())
			}
		})
	}
}

func TestN_Float64(t *testing.T) {
	cases := []struct {
		r     r128
		f     float64
		exact bool
	}{
		{r128{0, 1}, 0, true},
		{r128{1, 1}, 1, true},
		{r128{-1, 1}, -1, true},
		{r128{1, 2}, 0.5, true},
		{r128{-1, 2}, -0.5, true},
		{r128{1, 5}, 0.2, false},
		{r128{-1, 5}, -0.2, false},
		{r128{1, 9}, 0.111_111_111_111_111_111, false},
		{r128{2, 3}, 0.666_666_666_666_666_666, false},
		{r128{-2, 3}, -0.666_666_666_666_666_666, false},
		{r128{1, 7}, 0.142_857_142_857_142_857, false},
		{r128{1<<63 - 1, 1}, 9.223_372_036_854_775_807e18, false},
	}
	for _, c := range cases {
		t.Run(c.r.Rat128().String(), func(t *testing.T) {
			f, exact := c.r.Rat128().Float64()
			if exact != c.exact {
				t.Errorf("got exact=%v, want %v", exact, c.exact)
			}
			if exact {
				if f != c.f {
					t.Errorf("got %g, want %g", f, c.f)
				}
			} else {
				next := math.Nextafter(c.f, math.Inf(1))
				prev := math.Nextafter(c.f, math.Inf(-1))
				if f > next || f < prev {
					t.Errorf("got %g, want ~%g", f, c.f)
				}
			}
		})
	}
}

func TestParseDecimalString(t *testing.T) {
	cases := []struct {
		s string
		r r128
		e bool
	}{
		{"0", r128{0, 1}, false},
		{"1.0", r128{1, 1}, false},
		{"1.23", r128{123, 100}, false},
		{"123.0", r128{123, 1}, false},
		{"1234567890.", r128{1234567890, 1}, false},
		{"123456789.0", r128{123456789, 1}, false},
		{"12345678.90", r128{123456789, 10}, false},
		{"1234567.890", r128{123456789, 100}, false},
		{"123456.7890", r128{123456789, 1000}, false},
		{"12345.67890", r128{123456789, 10_000}, false},
		{"1234.567890", r128{123456789, 100_000}, false},
		{"123.4567890", r128{123456789, 1_000_000}, false},
		{"12.34567890", r128{123456789, 10_000_000}, false},
		{"1.234567890", r128{123456789, 100_000_000}, false},
		{".1234567890", r128{123456789, 1_000_000_000}, false},
		{".01234567890", r128{123456789, 10_000_000_000}, false},
		{".001234567890", r128{123456789, 100_000_000_000}, false},
		{"", r128{}, true},
		{" ", r128{}, true},
		{".", r128{}, true},
		{"-", r128{}, true},
		{"a", r128{}, true},
		{"1234567890123456789012345678901234567890", r128{}, true},
		{"1234567890123456789.012345678901234567890", r128{}, true},
		{".1234567890123456789012345678901234567890", r128{}, true},
		{"0000000000000000000000000000000000000000", r128{0, 1}, false},
		{"0000000000000000000000000000000000000001", r128{1, 1}, false},
		{"1.000000000000000000000000000000000000000", r128{1, 1}, false},
		{"1000000000000000000000000000000000000001", r128{}, true},
		{"1.000000000000000000000000000000000000001", r128{}, true},
		{"000000000000000000000000000000000000000101", r128{101, 1}, false},
		{"1.010000000000000000000000000000000000000", r128{101, 100}, false},
		{"0.000001010000000000000000000000000000000", r128{101, 100_000_000}, false},
	}
	for _, c := range cases {
		t.Run(c.s, func(t *testing.T) {
			r, err := rat128.ParseDecimalString(c.s)
			if !c.e {
				if err != nil {
					t.Fatalf("got unexpected error %v", err)
				}
				if r != c.r.Rat128() {
					t.Errorf("got value %s, want %s", r, c.r.Rat128())
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
		r r128
		p int
		s string
	}{
		{r128{0, 1}, -1, "0"},
		{r128{0, 1}, 0, "0"},
		{r128{0, 1}, 1, "0.0"},
		{r128{0, 1}, 2, "0.00"},
		{r128{0, 1}, 3, "0.000"},
		{r128{1, 1}, -1, "1"},
		{r128{1, 1}, 0, "1"},
		{r128{1, 1}, 1, "1.0"},
		{r128{1, 1}, 2, "1.00"},
		{r128{1, 1}, 3, "1.000"},
		{r128{-1, 1}, 0, "-1"},
		{r128{-1, 1}, 1, "-1.0"},
		{r128{-1, 1}, 2, "-1.00"},
		{r128{-1, 1}, 3, "-1.000"},
		{r128{1, 2}, -1, "1"},
		{r128{1, 2}, 0, "1"},
		{r128{1, 2}, 1, "0.5"},
		{r128{1, 2}, 2, "0.50"},
		{r128{1, 2}, 3, "0.500"},
		{r128{-1, 2}, 0, "-1"},
		{r128{-1, 2}, 1, "-0.5"},
		{r128{-1, 2}, 2, "-0.50"},
		{r128{-1, 2}, 3, "-0.500"},
		{r128{3, 4}, -1, "1"},
		{r128{3, 4}, 0, "1"},
		{r128{3, 4}, 1, "0.8"},
		{r128{3, 4}, 2, "0.75"},
		{r128{3, 4}, 3, "0.750"},
		{r128{1, 3}, 0, "0"},
		{r128{1, 3}, 1, "0.3"},
		{r128{1, 3}, 2, "0.33"},
		{r128{1, 3}, 3, "0.333"},
		{r128{-1, 3}, 0, "-0"}, // matches big.Rat.FloatString
		{r128{-1, 3}, 1, "-0.3"},
		{r128{-1, 3}, 2, "-0.33"},
		{r128{-1, 3}, 3, "-0.333"},
		{r128{2, 3}, 0, "1"},
		{r128{2, 3}, 1, "0.7"},
		{r128{2, 3}, 2, "0.67"},
		{r128{2, 3}, 3, "0.667"},
		{r128{-2, 3}, 0, "-1"},
		{r128{-2, 3}, 1, "-0.7"},
		{r128{-2, 3}, 2, "-0.67"},
		{r128{-2, 3}, 3, "-0.667"},
		{r128{4, 3}, 0, "1"},
		{r128{4, 3}, 1, "1.3"},
		{r128{4, 3}, 2, "1.33"},
		{r128{4, 3}, 3, "1.333"},
		{r128{-4, 3}, 0, "-1"},
		{r128{-4, 3}, 1, "-1.3"},
		{r128{-4, 3}, 2, "-1.33"},
		{r128{-4, 3}, 3, "-1.333"},
		{r128{76, 7}, 0, "11"},
		{r128{76, 7}, 1, "10.9"},
		{r128{76, 7}, 2, "10.86"},
		{r128{76, 7}, 3, "10.857"},
		{r128{76, 7}, 4, "10.8571"},
		{r128{76, 7}, 5, "10.85714"},
		{r128{1<<63 - 1, 2}, 1, "4611686018427387903.5"},

		// the following reference values were obtained from big.Rat.FloatString
		{r128{1<<63 - 2, 1<<63 - 1}, 18, "1.000000000000000000"},
		{r128{1<<63 - 2, 1<<63 - 1}, 19, "0.9999999999999999999"},
		{r128{1<<63 - 2, 1<<63 - 1}, 20, "0.99999999999999999989"},
		{r128{1<<63 - 2, 1<<63 - 1}, 21, "0.999999999999999999892"},
		{r128{1<<63 - 2, 1<<63 - 1}, 22, "0.9999999999999999998916"},
		{r128{1<<63 - 2, 1<<63 - 1}, 23, "0.99999999999999999989158"},
		{r128{1<<63 - 2, 1<<63 - 1}, 24, "0.999999999999999999891580"},
		{r128{1<<63 - 2, 1<<63 - 1}, 25, "0.9999999999999999998915798"},
	}
	for _, c := range cases {
		r := c.r.Rat128()
		t.Run(fmt.Sprintf("(%s):%d", r, c.p), func(t *testing.T) {
			s := r.DecimalString(c.p)
			if s != c.s {
				t.Errorf("got %s, want %s", s, c.s)
			}
		})
	}
}
