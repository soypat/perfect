package perfect

import (
	"fmt"
	"go/token"
	"log"
)

func ExampleHashFinder_goKeywords() {
	var keywords []string
	for tok := token.Token(0); tok < 256; tok++ {
		if tok.IsKeyword() {
			keywords = append(keywords, tok.String())
		}
	}
	var phf HashFinder
	hasher := &HashSequential{
		LenCoef: Coef{IndexApplied: 0, OnlyPow2: true, Op: OpAdd},
		Coefs: []Coef{
			{IndexApplied: 0, OnlyPow2: true, Op: OpXor},
			{IndexApplied: 1, OnlyPow2: true, Op: OpXor},
		},
	}
	err := hasher.ConfigCoefs(16)
	if err != nil {
		log.Fatalln(err)
	}
	const tablesizebits = 6
	log.Printf("exhaustive search for perfect hash for Go's %d keywords, table size of %d", len(keywords), 1<<tablesizebits)
	attempts, err := phf.Search(hasher, tablesizebits, keywords)
	if err != nil {
		log.Fatalln(err, "after", attempts, "attempts")
	}
	fmt.Print(hasher.String())
	// Output:
	// h := uint(len(s))*8
	// h ^= uint(s[0])*1
	// h ^= uint(s[1])*8
}
