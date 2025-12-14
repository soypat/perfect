# perfect

[![Go Reference](https://pkg.go.dev/badge/github.com/soypat/perfect.svg)](https://pkg.go.dev/github.com/soypat/perfect)

A Go library for finding perfect hash functions for static string sets.

## What is a Perfect Hash?

A perfect hash function maps a set of keys to unique integers with no collisions. This is useful for:
- Fast keyword/token lookup in lexers and parsers
- Efficient symbol tables with known keys
- Compile-time lookup tables

## Installation

```bash
go get github.com/soypat/perfect
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

## License

See [LICENSE](LICENSE).
