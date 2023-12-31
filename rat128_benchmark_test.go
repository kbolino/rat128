package rat128_test

import (
	"math/big"
	"testing"

	"github.com/kbolino/rat128"
)

var BenchCases = map[string]struct {
	X, Y rat128.N
}{
	"Small":   {New(7, 11*13), New(11, 7*13)},
	"WideAdd": {New(P1, P2*P3), New(P2, P1*P3)},
	"WideMul": {New(P1*P2, P3), New(P3, P4)},
}

func BenchmarkRat128_Add(b *testing.B) {
	for name, c := range BenchCases {
		x, y := c.X, c.Y
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				x.Add(y)
			}
		})
	}
}

func BenchmarkRat128_Mul(b *testing.B) {
	for name, c := range BenchCases {
		x, y := c.X, c.Y
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				x.Mul(y)
			}
		})
	}
}

func BenchmarkBigRat_Add(b *testing.B) {
	z := new(big.Rat)
	for name, c := range BenchCases {
		x, y := c.X.BigRat(), c.Y.BigRat()
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				z.Add(x, y)
			}
		})
	}
}

func BenchmarkBigRat_Mul(b *testing.B) {
	z := new(big.Rat)
	for name, c := range BenchCases {
		x, y := c.X.BigRat(), c.Y.BigRat()
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				z.Mul(x, y)
			}
		})
	}
}
