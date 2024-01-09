# rat128

Fixed-precision rational numbers for Go. This module has no dependencies and
is fairly straightforward to use:

- Construct new values with `var x rat128.N`, `x := rat128.New(m, n)`, or
  `x, err := rat128.Try(m, n)`.
- Retrieve the numerator and denominator with `x.Num()` and `x.Den()`,
  respectively.
- Perform arithmetic with panicking `x.Add(y)`, `x.Mul(y)` or error-returning
  `x.TryAdd(y)`, `x.TryMul(y)`, etc.
- Convert from/to `float64` with `FromFloat64` and `x.Float64()` respectively.
- Convert from/to decimal strings (`"12.34"`) with `ParseDecimalString` and
  `x.DecimalString(digits)` respectively.

## Design Goals

- First and foremost: value semantics.
  Unlike `big.Rat` in the standard library, `rat128.N` is a value type like
  `int64` or `float32` and is safe to copy and compare with `==` and `!=`.
  Methods return values instead of mutating their receiver.
- Mathematical operations should return correct results. If they don't, the
  incorrect behavior is a bug and fixing this bug will be treated as a
  non-breaking change. Do not rely on incorrect results.
- Widened integer arithmetic is used where possible to avoid overflow of
  intermediate values used in basic arithmetic operations (add, subtract,
  multiply, and divide). The finite precision is more limiting than `big.Rat`,
  however, and panics from overflow of the numerator or denominator in the
  final result are possible.
- Valid values are always in reduced form, to simplify operations and reduce
  the risk of overflow.
- Converting to and from floating-point values should be exact wherever it is
  possible to do so, and approximation should be explicit.
- There are panicking and panic-free versions of most operations. This means
  there are actually three ways to use this library:
  - Use the panicking operations (New, Add, etc.)
  - Use the panic-free operations (Try, TryAdd, etc.) and check for errors
  - Use the panic-free operations and ignore the errors; this is not
    recommended but will be fastest

## Reporting issues

File bug reports, feature requests, etc. through GitHub Issues on this
repository.

The following are always considered issues:

- Incorrect mathematical results
- Violation of the design goals; though fixing these may require breaking
  changes and thus a new major version
- Unexpected behavior; for example, operations on otherwise valid values
  returning invalid results instead of valid results, a panic, or an error
- Panics in `Try*` functions/methods; these return errors, so ordinary
  arithmetic exceptions should not cause panics

The following will be evaluated on a case-by-case basis:

- Optimization; in particular, micro-optimizations that reduce readability
  of the code for little measurable benefit may be rejected
- (Un-)marshaling/(de-)serializing in various formats
- Undefined behavior; for example, the results of operations on invalid
  values created through unsafe code, reflection, unsychronized sharing
  across goroutines, cgo hacks, memory corruption, etc.
- Converting to formats outside of the core language and standard library
- Performance regressions; in general, this library should outperform `big.Rat`
  on an apples-to-apples basis on 64-bit machines, but the exact performance
  characteristics are not guaranteed

## To Do

- Improved overflow detection to reduce usage of widened arithmetic
- Other optimizations where feasible/practical (e.g. faster `Cmp`)
- More test coverage
- Convenience methods as practical usage demonstrates their value
