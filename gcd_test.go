package rat128_test

import (
	"fmt"
	"testing"

	"github.com/kbolino/rat128"
)

type gcdCase struct {
	m, n, d int64
}

var gcdCases = []gcdCase{
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
	{p1 * p2 * p3, p2 * p3 * p4, p2 * p3},
}

var symCases []gcdCase

func init() {
	symCases = append(symCases, gcdCases...)
	for _, c := range gcdCases {
		if c.m == c.n {
			continue
		}
		symCases = append(symCases, gcdCase{c.n, c.m, c.d})
	}
}

func BenchmarkExtGCD(b *testing.B) {
	for _, c := range gcdCases {
		b.Run(fmt.Sprintf("ExtGCD(%d,%d)", c.m, c.n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				rat128.ExtGCD(123456789, 987654321)
			}
		})
	}
}

func TestExtGCD(t *testing.T) {
	for _, c := range symCases {
		t.Run(fmt.Sprintf("ExtGCD(%d,%d)", c.m, c.n), func(t *testing.T) {
			a, b, d := rat128.ExtGCD(c.m, c.n)
			if d != c.d {
				t.Errorf("_, _, d := ExtGCD(%d, %d); d == %d != %d", c.m, c.n, d, c.d)
			}
			if a*c.m+b*c.n != d {
				t.Errorf("a, b, _ := ExtGCD(%d, %d); a*%d+b*%d == %d != %d", c.m, c.n, c.m, c.n, a*c.m+b*c.n, d)
			}
		})
	}
}
