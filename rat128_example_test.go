package rat128_test

import (
	"fmt"

	"github.com/kbolino/rat128"
)

func ExampleNew() {
	n := rat128.New(1, 2)
	fmt.Println(n)
	// Output: 1/2
}

func ExampleTry() {
	n, err := rat128.Try(1, 2)
	if err != nil {
		panic(err)
	}
	fmt.Println(n)
	// Output: 1/2
}

func ExampleTry_denomZero() {
	_, err := rat128.Try(1, 0)
	fmt.Println(err)
	// Output: denominator is not positive
}

func ExampleParseRationalString_oneHalf() {
	n, err := rat128.ParseRationalString("1/2")
	if err != nil {
		panic(err)
	}
	fmt.Println(n)
	// Output: 1/2
}

func ExampleParseRationalString_negOneHalf() {
	n, err := rat128.ParseRationalString("-1/2")
	if err != nil {
		panic(err)
	}
	fmt.Println(n)
	// Output: -1/2
}

func ExampleParseRationalString_twoFourths() {
	n, err := rat128.ParseRationalString("2/4")
	if err != nil {
		panic(err)
	}
	fmt.Println(n)
	// Output: 1/2
}

func ExampleParseRationalString_denomZero() {
	_, err := rat128.ParseRationalString("1/0")
	fmt.Println(err)
	// Output: denominator is not positive
}

func ExampleN_Add() {
	x := rat128.New(1, 2)
	y := rat128.New(1, 3)
	z := x.Add(y)
	fmt.Println(z)
	// Output: 5/6
}

func ExampleN_Sub() {
	x := rat128.New(1, 2)
	y := rat128.New(1, 3)
	z := x.Sub(y)
	fmt.Println(z)
	// Output: 1/6
}

func ExampleN_Mul() {
	x := rat128.New(1, 2)
	y := rat128.New(2, 3)
	z := x.Mul(y)
	fmt.Println(z)
	// Output: 1/3
}

func ExampleN_Div() {
	x := rat128.New(1, 2)
	y := rat128.New(2, 3)
	z := x.Div(y)
	fmt.Println(z)
	// Output: 3/4
}
