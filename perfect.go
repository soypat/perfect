// Package perfect finds perfect hash functions for a given set of strings.
// It searches coefficient space to find a hash that maps all inputs to unique values.
package perfect

import (
	"errors"
	"fmt"
	"slices"
)

// Hash represents a hash function that can be incremented to try new coefficients.
type Hash interface {
	Hash(dataToHash string) uint
	Increment() (done bool)
}

// HashFinder searches for perfect hash coefficients.
type HashFinder struct {
	hashmap []uint
}

// HashSequential computes: h = len(s)*LenCoef + op(s[i])*Coefs[i] for each coefficient.
type HashSequential struct {
	LenCoef Coef
	Coefs   []Coef
}

// ConfigCoefs initializes all coefficients and sets MaxValue to defaultMax where unset.
func (hs *HashSequential) ConfigCoefs(defaultMax uint) error {
	coefs := hs.Coefs
	for i := range coefs {
		err := coefs[i].config(defaultMax)
		if err != nil {
			return err
		}
	}
	return hs.LenCoef.config(defaultMax)
}

func (hs *HashSequential) String() string {
	s := fmt.Sprintf("h := uint(len(s))*%d\n", hs.LenCoef.Value)
	for _, c := range hs.Coefs {
		pfx := ""
		if c.IndexApplied < 0 {
			pfx = "len(s)"
		}
		s += fmt.Sprintf("h %s= uint(s[%s%d])*%d\n", c.Op.String(), pfx, c.IndexApplied, c.Value)
	}
	return s
}

// Hash computes the hash value for the given string.
func (hs *HashSequential) Hash(dataToHash string) uint {
	h := uint(len(dataToHash)) * hs.LenCoef.Value
	for i := range hs.Coefs {
		h = hs.Coefs[i].Apply(h, dataToHash)
	}
	return h
}

// Increment advances coefficients to try the next hash function. Returns true when exhausted.
func (hs *HashSequential) Increment() (done bool) {
	coefs := hs.Coefs
	coefs[0].Increment()
	for i := 0; coefs[i].Saturated() && i < len(coefs)-1; i++ {
		coefs[i].init()
		coefs[i+1].Increment()
	}
	lastIsSaturated := coefs[len(coefs)-1].Saturated()
	if lastIsSaturated {
		coefs[len(coefs)-1].init()
		hs.LenCoef.Increment()
	}
	return hs.LenCoef.Value > hs.LenCoef.MaxValue // Check for super saturation of last coefficient which is never reset (length coefficient)
}

// Coef is a single coefficient in the hash function.
type Coef struct {
	IndexApplied int  // Byte index to use. Negative indexes from end.
	Value        uint // Current coefficient value.
	MaxValue     uint
	StartValue   uint
	OnlyPow2     bool
	Op           Op
}

// ErrNoCoefficientsFound is returned when no perfect hash exists in the search space.
var ErrNoCoefficientsFound = errors.New("no coefficients found")

func (c *Coef) init() {
	if c.StartValue == 0 {
		c.Value = 1
	} else {
		c.Value = c.StartValue
	}
	if c.Op == 0 {
		c.Op = OpAdd
	}
}

// Increment advances the coefficient value.
func (c *Coef) Increment() {
	if c.OnlyPow2 {
		c.Value *= 2
	} else {
		c.Value++
	}
}

// Saturated returns true when the coefficient has reached its maximum value.
func (c *Coef) Saturated() bool { return c.Value >= c.MaxValue }

// Search finds coefficients that produce unique hashes for all inputs.
// Returns the number of attempts and an error if no perfect hash was found.
func (phf *HashFinder) Search(hasher Hash, tableSizeBits int, inputs []string) (int, error) {
	if tableSizeBits <= 0 || tableSizeBits > 32 {
		return 0, errors.New("zero/negative bits for table size or too large")
	} else if len(inputs) == 0 {
		return 0, errors.New("zero inputs")
	}
	tblsz := 1 << tableSizeBits
	phf.hashmap = slices.Grow(phf.hashmap[:0], tblsz)[:tblsz]
	hashmap := phf.hashmap
	mask := uint(tblsz) - 1
	currentAttempt := 0
	for {
		currentAttempt++
		attemptSuccess := true
		clear(hashmap)
		for _, kw := range inputs {
			h := hasher.Hash(kw) & mask
			tok := hashmap[h]
			if tok != 0 {
				attemptSuccess = false
				break
			}
			hashmap[h] = 1
		}
		if attemptSuccess {
			return currentAttempt, nil
		}
		cannotContinue := hasher.Increment()
		if cannotContinue {
			break
		}
	}
	return currentAttempt, ErrNoCoefficientsFound
}

// Apply combines the byte at IndexApplied with h using the coefficient's operation.
func (coef *Coef) Apply(h uint, kw string) uint {
	idx := coef.IndexApplied
	var a uint
	if idx < 0 && -idx <= len(kw) {
		a = uint(kw[len(kw)+idx]) * coef.Value
	} else if idx >= 0 && idx < len(kw) {
		a = uint(kw[idx]) * coef.Value
	}
	switch coef.Op {
	case OpAdd:
		h += a
	case OpXor:
		h ^= a
	case OpMul:
		h *= a
	default:
		panic("unsupported operation")
	}
	return h
}

func (coef *Coef) config(defaultMax uint) error {
	coef.init()
	if coef.MaxValue == 0 {
		if defaultMax <= 0 {
			return errors.New("default max coefficient need be set and positive for input")
		}
		coef.MaxValue = defaultMax
	}
	return nil
}

// Op defines the arithmetic operation used to combine a byte with the hash.
type Op int

const (
	opUndefined Op = iota
	OpAdd          // Addition
	OpXor          // XOR
	OpMul          // Multiplication
)

func (op Op) String() (s string) {
	switch op {
	case OpAdd:
		s = "+"
	case OpXor:
		s = "^"
	case OpMul:
		s = "*"
	default:
		s = "<unknownop>"
	}
	return s
}
