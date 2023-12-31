package rat128_test

import (
	"math/big"
	"testing"
)

var benchCases = map[string]struct {
	x, y r128
}{
	"small":    {x: r128{7, 11 * 13}, y: r128{11, 7 * 13}},
	"wide_add": {x: r128{p1, p2 * p3}, y: r128{p2, p1 * p3}},
	"wide_mul": {x: r128{p1 * p2, p3}, y: r128{p3, p4}},
}

func BenchmarkRat128_Add(b *testing.B) {
	for name, c := range benchCases {
		x, y := c.x.Rat128(), c.y.Rat128()
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				x.Add(y)
			}
		})
	}
}

func BenchmarkRat128_Mul(b *testing.B) {
	for name, c := range benchCases {
		x, y := c.x.Rat128(), c.y.Rat128()
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				x.Mul(y)
			}
		})
	}
}

func BenchmarkBigRat_Add(b *testing.B) {
	z := new(big.Rat)
	for name, c := range benchCases {
		x, y := big.NewRat(c.x[0], c.x[1]), big.NewRat(c.y[0], c.y[1])
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				z.Add(x, y)
			}
		})
	}
}

func BenchmarkBigRat_Mul(b *testing.B) {
	z := new(big.Rat)
	for name, c := range benchCases {
		x, y := big.NewRat(c.x[0], c.x[1]), big.NewRat(c.y[0], c.y[1])
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				z.Mul(x, y)
			}
		})
	}
}
