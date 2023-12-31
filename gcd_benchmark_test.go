package rat128_test

import (
	"fmt"
	"testing"

	"github.com/kbolino/rat128"
)

func BenchmarkExtGCD(b *testing.B) {
	for _, c := range GCDCases {
		b.Run(fmt.Sprintf("ExtGCD(%d,%d)", c.M, c.N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				rat128.ExtGCD(c.M, c.N)
			}
		})
	}
}
