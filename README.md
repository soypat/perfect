# perfect

[![Go Reference](https://pkg.go.dev/badge/github.com/soypat/perfect.svg)](https://pkg.go.dev/github.com/soypat/perfect)

A Go library for finding perfect hash functions for static string sets.

```
exhaustive search for perfect hash for Go's 25 keywords, table size of 64 (98.86% collision free probability)
```

See working example for Go's keywords [`example_test.go`](./example_test.go).
## What is a Perfect Hash?

A perfect hash function maps a set of keys to unique integers with no collisions. This is useful for:
- Fast keyword/token lookup in lexers and parsers
- Efficient symbol tables with known keys
- Compile-time lookup tables

## Installation

```bash
go get github.com/soypat/perfect@latest
```

## Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/soypat/perfect"
)

func main() {
    keywords := []string{"if", "else", "for", "return", "func", "var", "const"}

    hasher := &perfect.HashSequential{
        LenCoef: perfect.Coef{MaxValue: 32},
        Coefs: []perfect.Coef{
            {IndexApplied: 0},  // first byte
            {IndexApplied: 1},  // second byte
            {IndexApplied: -1}, // last byte
        },
    }
    hasher.ConfigCoefs(32)

    var finder perfect.HashFinder
    attempts, err := finder.Search(hasher, 4, keywords) // 2^4 = 16 slots
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found after %d attempts:\n%s", attempts, hasher)
}
```

## How It Works

The library searches for coefficients that produce a perfect hash of the form:

```
h = len(s) * lenCoef
h op= s[i] * coef[i]  // for each coefficient
h &= mask             // mask to table size
```

Where `op` can be `+`, `^` (XOR), or `*`.

### Search Strategy

1. Configure coefficients with index positions and value ranges
2. Call `Search()` which iterates through coefficient combinations
3. For each combination, test if all inputs hash to unique values
4. Returns when a perfect hash is found or search space is exhausted

For larger key sets, use randomized search by calling `Search()` multiple times with randomized coefficient starting points.

## Examples

See [`examples/`](examples/) for complete examples:
- [`examples/fortran/`](examples/fortran/) - Perfect hash for Fortran intrinsics

```
go run ./examples/fortran
2025/12/16 12:25:36 intrinsics: Searching perfect hash for 82 intrinsics with 5 coefficients
[59µs] intrinsic search
2025/12/16 12:25:36 intrinsics: perfect hash found after 185 attempts:
h := uint(len(s))*1
h += uint(s[0])*56
h *= uint(s[1])*47
h *= uint(s[len(s)-2])*7
h += uint(s[len(s)-1])*34
2025/12/16 12:25:36 keywords: Searching perfect hash for 95 keywords with 5 coefficients
[34µs] keyword search
2025/12/16 12:25:36 keywords: perfect hash found after 105 attempts:
h := uint(len(s))*1
h += uint(s[0])*56
h *= uint(s[1])*51
h *= uint(s[len(s)-2])*5
h += uint(s[len(s)-1])*35
2025/12/16 12:25:36 vendored: Searching perfect hash for 117 intrinsics(vendored) with 7 coefficients
[49.605ms] vendored intrinsic search
2025/12/16 12:25:36 vendored: perfect hash found after 160986 attempts:
h := uint(len(s))*1
h += uint(s[0])*57
h *= uint(s[1])*49
h *= uint(s[len(s)-2])*8
h += uint(s[len(s)-1])*43
h += uint(s[2])*58
h *= uint(s[len(s)-3])*47
```

## License

See [LICENSE](LICENSE).
