package rat128_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/kbolino/rat128"
)

type GCDCase struct {
	M, N, D int64
}

var GCDCases = []GCDCase{
	{1, 1, 1},
	{1, 2, 1},
	{2, 2, 2},
	{2, 3, 1},
	{2, 4, 2},
	{2, 6, 2},
	{3, 6, 3},
	{4, 6, 2},
	{6, 6, 6},
	{6, 8, 2},
	{6, 9, 3},
	{24, 120, 24},
	{36, 120, 12},
	{7, 360, 1},
	{7, 14, 7},
	{7, 21, 7},
	{360, 92821, 1},
	{360, 92822, 2},
	{3600, 216000, 3600},
	{123456789, 987654321, 9},
	{P1 * P2 * P3, P2 * P3 * P4, P2 * P3},
	{
		2 * 3 * 5 * 7 * 11 * 13 * 17 * 19 * 23 * 29 * 31 * 37 * 41 * 43 * 47,
		2 * 3 * 5 * 7 * 11 * 13 * 17 * 19 * 23 * 29 * 31 * 37 * 41 * 43 * 53,
		2 * 3 * 5 * 7 * 11 * 13 * 17 * 19 * 23 * 29 * 31 * 37 * 41 * 43,
	},
	{math.MaxInt64 - 1, math.MaxInt64, 1},
}

var SymGCDCases []GCDCase

func init() {
	SymGCDCases = append(SymGCDCases, GCDCases...)
	for _, c := range GCDCases {
		if c.M == c.N {
			continue
		}
		SymGCDCases = append(SymGCDCases, GCDCase{c.N, c.M, c.D})
	}
}

func TestExtGCD(t *testing.T) {
	for _, c := range SymGCDCases {
		t.Run(fmt.Sprintf("ExtGCD(%d,%d)", c.M, c.N), func(t *testing.T) {
			a, b, d := rat128.ExtGCD(c.M, c.N)
			if d != c.D {
				t.Errorf("_, _, d := ExtGCD(%d, %d); d == %d != %d", c.M, c.N, d, c.D)
			}
			if a*c.M+b*c.N != d {
				t.Errorf("a, b, _ := ExtGCD(%d, %d); a*%d+b*%d == %d != %d", c.M, c.N, c.M, c.N, a*c.M+b*c.N, d)
			}
		})
	}
}
