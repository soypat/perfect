package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/soypat/perfect"
)

func main() {
	// Populate intrinsics, keywords.
	var keywords, intrinsics, vendored []string
	for kw := keywordBeg + 1; kw < keywordEnd; kw++ {
		s := kw.String()
		if strings.ToUpper(s) != s {
			continue // is a separator token.
		}
		keywords = append(keywords, s)
	}
	for intr := intrinsicUndefined + 1; intr < fortran77End; intr++ {
		s := intr.String()
		if strings.ToUpper(s) != s {
			continue // is a separator token.
		}
		intrinsics = append(intrinsics, s)
	}
	for vndr := vendorUndefined + 1; vndr < vendorPGIEnd; vndr++ {
		if !vndr.IsValid() {
			continue
		}
		vendored = append(vendored, vndr.String())
	}
	const maxCoef = 64
	hasher := &perfect.HashSequential{
		LenCoef: perfect.Coef{MaxValue: maxCoef},
		Coefs:   make([]perfect.Coef, 4),
	}
	for i := range hasher.Coefs {
		if i < 2 {
			hasher.Coefs[i].IndexApplied = i
		} else {
			hasher.Coefs[i].IndexApplied = i - len(hasher.Coefs)
		}
	}
	err := hasher.ConfigCoefs(maxCoef)
	if err != nil {
		log.Fatal(err)
	}
	var phf perfect.HashFinder

	// SEARCH INTRINSICS.

	log.Printf("intrinsics: Searching perfect hash for %d intrinsics with %d coefficients", len(intrinsics), len(hasher.Coefs)+1)
	randomRetries := 1000
	attempts := 0
	found := false
	tm := timer("intrinsic search")
	for _, tbits := range []int{10, 11} {
		a, err := randomSearch(hasher, &phf, intrinsics, tbits, randomRetries)
		attempts += a
		if err == nil {
			found = true
			break
		} else if !errors.Is(err, perfect.ErrNoCoefficientsFound) {
			log.Fatal(err)
		}
	}
	if found {
		tm()
		log.Printf("intrinsics: perfect hash found after %d attempts:\n%s", attempts, hasher.String())
	} else {
		log.Printf("intrinsics: no perfect hash found after %d attempts", attempts)
	}

	// SEARCH KEYWORDS.

	found = false
	attempts = 0
	log.Printf("keywords: Searching perfect hash for %d keywords with %d coefficients", len(keywords), len(hasher.Coefs)+1)
	tm = timer("keyword search")
	for _, tbits := range []int{10, 11} {
		a, err := randomSearch(hasher, &phf, keywords, tbits, randomRetries)
		attempts += a
		if err == nil {
			found = true
			break
		} else if !errors.Is(err, perfect.ErrNoCoefficientsFound) {
			log.Fatal(err)
		}
	}
	if found {
		tm()
		log.Printf("keywords: perfect hash found after %d attempts:\n%s", attempts, hasher.String())
	} else {
		log.Printf("keywords: no perfect hash found after %d attempts", attempts)
	}

	// SEARCH VENDORED INTRINSICS.

	// Require more search space for this one, add two coefficients.
	hasher.Coefs = append(hasher.Coefs, perfect.Coef{IndexApplied: 2}, perfect.Coef{IndexApplied: -3})
	hasher.ConfigCoefs(maxCoef)
	found = false
	attempts = 0
	log.Printf("vendored: Searching perfect hash for %d intrinsics(vendored) with %d coefficients", len(vendored), len(hasher.Coefs)+1)
	tm = timer("vendored intrinsic search")
	for _, tbits := range []int{10, 11} {
		a, err := randomSearch(hasher, &phf, vendored, tbits, randomRetries)
		attempts += a
		if err == nil {
			found = true
			break
		} else if !errors.Is(err, perfect.ErrNoCoefficientsFound) {
			log.Fatal(err)
		}
	}
	if found {
		tm()
		log.Printf("vendored: perfect hash found after %d attempts:\n%s", attempts, hasher.String())
	} else {
		log.Printf("vendored: no perfect hash found after %d attempts", attempts)
	}
}

func randomSearch(hasher *perfect.HashSequential, phf *perfect.HashFinder, words []string, tableBits int, randomRetries int) (attempts int, err error) {
	const neighborhoodSearchSpace = 10
	rng := rand.New(rand.NewPCG(1, 1))
	for range randomRetries {
		for i := range hasher.Coefs {
			randomizeCoef(&hasher.Coefs[i], rng, neighborhoodSearchSpace)
		}
		currentAttempts, err := phf.Search(hasher, tableBits, words)
		attempts += currentAttempts
		if err == nil {
			return attempts, nil
		} else if !errors.Is(err, perfect.ErrNoCoefficientsFound) {
			return attempts, err
		}
	}
	return attempts, perfect.ErrNoCoefficientsFound
}

var ops = []perfect.Op{perfect.OpAdd, perfect.OpXor, perfect.OpMul}

func randomizeCoef(c *perfect.Coef, rng *rand.Rand, searchSpace int) {
	start := rng.IntN(int(c.MaxValue))
	end := min(start+searchSpace, int(c.MaxValue))
	*c = perfect.Coef{
		IndexApplied: c.IndexApplied, // Keep indexing, user should provide intelligence here on best indexing.
		Value:        uint(start),
		StartValue:   uint(start),
		MaxValue:     uint(end),
		OnlyPow2:     false,
		Op:           ops[rng.IntN(len(ops))],
	}

}

func timer(context string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		fmt.Printf("[%s] %s\n", elapsed.Round(time.Microsecond), context)
	}
}

// Install stringer tool:
//  go install golang.org/x/tools/cmd/stringer@latest

//go:generate stringer -type=Token,Intrinsic,VendorIntrinsic -linecomment -output stringers.go .

type Token int

const (
	// Not to be used in code. Is to catch uninitialized tokens.
	Undefined Token = iota // <undefined>

	// ==================== KEYWORDS ====================

	// Type declaration keywords
	keywordBeg      // invalid
	INTEGER         // INTEGER
	REAL            // REAL
	COMPLEX         // COMPLEX
	LOGICAL         // LOGICAL
	CHARACTER       // CHARACTER
	DOUBLE          // DOUBLE
	PRECISION       // PRECISION
	DOUBLECOMPLEX   // DOUBLECOMPLEX
	DOUBLEPRECISION // DOUBLEPRECISION

	// Program structure keywords
	PROGRAM       // PROGRAM
	END           // END
	ENDPROGRAM    // ENDPROGRAM
	SUBROUTINE    // SUBROUTINE
	ENDSUBROUTINE // ENDSUBROUTINE
	FUNCTION      // FUNCTION
	ENDFUNCTION   // ENDFUNCTION
	MODULE        // MODULE
	ENDMODULE     // ENDMODULE
	CONTAINS      // CONTAINS
	ENTRY         // ENTRY
	BLOCK         // BLOCK

	// Control flow keywords
	IF        // IF
	THEN      // THEN
	ELSE      // ELSE
	ELSEIF    // ELSEIF
	ENDIF     // ENDIF
	DO        // DO
	ENDDO     // ENDDO
	WHILE     // WHILE
	SELECT    // SELECT
	CASE      // CASE
	DEFAULT   // DEFAULT
	ENDSELECT // ENDSELECT
	CYCLE     // CYCLE
	EXIT      // EXIT
	GOTO      // GOTO
	GO        // GO
	TO        // TO
	CONTINUE  // CONTINUE
	RETURN    // RETURN
	STOP      // STOP

	// I/O keywords
	READ      // READ
	WRITE     // WRITE
	PRINT     // PRINT
	OPEN      // OPEN
	CLOSE     // CLOSE
	INQUIRE   // INQUIRE
	FILE      // FILE
	BACKSPACE // BACKSPACE
	REWIND    // REWIND
	ENDFILE   // ENDFILE
	FORMAT    // FORMAT
	NAMELIST  // NAMELIST

	// Declaration and specification keywords
	IMPLICIT    // IMPLICIT
	DATA        // DATA
	EQUIVALENCE // EQUIVALENCE
	COMMON      // COMMON
	EXTERNAL    // EXTERNAL
	INTRINSIC   // INTRINSIC
	SEQUENCE    // SEQUENCE

	// Interface and type keywords (F90)
	INTERFACE    // INTERFACE
	ENDINTERFACE // ENDINTERFACE
	TYPE         // TYPE
	ENDTYPE      // ENDTYPE

	// Module and visibility keywords (F90)
	USE  // USE
	ONLY // ONLY

	// Miscellaneous keywords
	CALL    // CALL
	ASSIGN  // ASSIGN
	INCLUDE // INCLUDE
	DEFINE  // DEFINE

	// Array operations (F90)
	WHERE     // WHERE
	ELSEWHERE // ELSEWHERE
	ENDWHERE  // ENDWHERE

	// ==================== ATTRIBUTES (F90) ====================
	attrStart

	SAVE        // SAVE
	PRIVATE     // PRIVATE
	PUBLIC      // PUBLIC
	PARAMETER   // PARAMETER
	DIMENSION   // DIMENSION
	INTENT      // INTENT
	IN          // IN
	OUT         // OUT
	INOUT       // INOUT
	OPTIONAL    // OPTIONAL
	POINTER     // POINTER
	TARGET      // TARGET
	ALLOCATABLE // ALLOCATABLE
	ALLOCATE    // ALLOCATE
	DEALLOCATE  // DEALLOCATE
	NULLIFY     // NULLIFY
	RECURSIVE   // RECURSIVE
	ELEMENTAL   // ELEMENTAL
	PURE        // PURE
	RESULT      // RESULT
	KIND        // KIND
	LEN         // LEN

	keywordEnd // invalid
)

type Intrinsic uint

const (
	intrinsicUndefined Intrinsic = 0 // undefined
)

// Fortran 66 function Intrinsics (FORTRAN IV compatible)
const (
	fortran66Start Intrinsic = iota + intrinsicUndefined + 1
	// Mathematical functions
	IntrinsicABS   // ABS
	IntrinsicMOD   // MOD
	IntrinsicSIGN  // SIGN
	IntrinsicDIM   // DIM
	IntrinsicDPROD // DPROD
	// Trigonometric
	IntrinsicSIN   // SIN
	IntrinsicCOS   // COS
	IntrinsicTAN   // TAN
	IntrinsicASIN  // ASIN
	IntrinsicACOS  // ACOS
	IntrinsicATAN  // ATAN
	IntrinsicATAN2 // ATAN2
	// Hyperbolic
	IntrinsicSINH // SINH
	IntrinsicCOSH // COSH
	IntrinsicTANH // TANH
	// Exponential and logarithmic
	IntrinsicEXP   // EXP
	IntrinsicLOG   // LOG
	IntrinsicLOG10 // LOG10
	IntrinsicSQRT  // SQRT
	// Type conversion
	IntrinsicINT   // INT
	IntrinsicREAL  // REAL
	IntrinsicDBLE  // DBLE
	IntrinsicCMPLX // CMPLX
	IntrinsicFLOAT // FLOAT
	IntrinsicIFIX  // IFIX
	IntrinsicSNGL  // SNGL
	// Truncation and rounding
	IntrinsicAINT  // AINT
	IntrinsicANINT // ANINT
	IntrinsicNINT  // NINT
	// Min/Max
	IntrinsicMAX   // MAX
	IntrinsicMIN   // MIN
	IntrinsicMAX0  // MAX0
	IntrinsicMAX1  // MAX1
	IntrinsicMIN0  // MIN0
	IntrinsicMIN1  // MIN1
	IntrinsicAMAX0 // AMAX0
	IntrinsicAMAX1 // AMAX1
	IntrinsicAMIN0 // AMIN0
	IntrinsicAMIN1 // AMIN1
	IntrinsicDMAX1 // DMAX1
	IntrinsicDMIN1 // DMIN1
	// Complex conjugate and imaginary part
	IntrinsicCONJG // CONJG
	IntrinsicAIMAG // AIMAG
	// Specific names for type variants
	IntrinsicIABS   // IABS
	IntrinsicDABS   // DABS
	IntrinsicCABS   // CABS
	IntrinsicDSIN   // DSIN
	IntrinsicDCOS   // DCOS
	IntrinsicDTAN   // DTAN
	IntrinsicDASIN  // DASIN
	IntrinsicDACOS  // DACOS
	IntrinsicDATAN  // DATAN
	IntrinsicDATAN2 // DATAN2
	IntrinsicDSINH  // DSINH
	IntrinsicDCOSH  // DCOSH
	IntrinsicDTANH  // DTANH
	IntrinsicDEXP   // DEXP
	IntrinsicDLOG   // DLOG
	IntrinsicDLOG10 // DLOG10
	IntrinsicDSQRT  // DSQRT
	IntrinsicCSIN   // CSIN
	IntrinsicCCOS   // CCOS
	IntrinsicCEXP   // CEXP
	IntrinsicCLOG   // CLOG
	IntrinsicCSQRT  // CSQRT
	IntrinsicIDIM   // IDIM
	IntrinsicDDIM   // DDIM
	IntrinsicIDINT  // IDINT
	IntrinsicISIGN  // ISIGN
	IntrinsicDSIGN  // DSIGN
	IntrinsicDNINT  // DNINT
	IntrinsicIDNINT // IDNINT
	IntrinsicDIMAG  // DIMAG
	IntrinsicDCONJG // DCONJG
	fortran66End
)

// Fortran 77 function Intrinsics (added string handling)
const (
	fortran77Start Intrinsic = iota + fortran66End
	// Character functions
	IntrinsicCHAR  // CHAR
	IntrinsicICHAR // ICHAR
	IntrinsicLEN   // LEN
	IntrinsicINDEX // INDEX
	// Lexical comparison
	IntrinsicLGE // LGE
	IntrinsicLGT // LGT
	IntrinsicLLE // LLE
	IntrinsicLLT // LLT
	fortran77End
)

// Fortran 90 function Intrinsics
const (
	fortran90Start Intrinsic = iota + fortran77End
	// Array reduction functions
	IntrinsicSUM     // SUM
	IntrinsicPRODUCT // PRODUCT
	IntrinsicMAXVAL  // MAXVAL
	IntrinsicMINVAL  // MINVAL
	IntrinsicALL     // ALL
	IntrinsicANY     // ANY
	IntrinsicCOUNT   // COUNT
	// Array inquiry functions
	IntrinsicSIZE      // SIZE
	IntrinsicSHAPE     // SHAPE
	IntrinsicLBOUND    // LBOUND
	IntrinsicUBOUND    // UBOUND
	IntrinsicALLOCATED // ALLOCATED
	// Array construction functions
	IntrinsicRESHAPE // RESHAPE
	IntrinsicSPREAD  // SPREAD
	IntrinsicPACK    // PACK
	IntrinsicUNPACK  // UNPACK
	IntrinsicMERGE   // MERGE
	// Array manipulation functions
	IntrinsicTRANSPOSE // TRANSPOSE
	IntrinsicCSHIFT    // CSHIFT
	IntrinsicEOSHIFT   // EOSHIFT
	// Array location functions
	IntrinsicMAXLOC // MAXLOC
	IntrinsicMINLOC // MINLOC
	// Matrix functions
	IntrinsicMATMUL      // MATMUL
	IntrinsicDOT_PRODUCT // DOT_PRODUCT
	// Bit manipulation functions
	IntrinsicIAND   // IAND
	IntrinsicIOR    // IOR
	IntrinsicIEOR   // IEOR
	IntrinsicNOT    // NOT
	IntrinsicBTEST  // BTEST
	IntrinsicIBSET  // IBSET
	IntrinsicIBCLR  // IBCLR
	IntrinsicIBITS  // IBITS
	IntrinsicISHFT  // ISHFT
	IntrinsicISHFTC // ISHFTC
	IntrinsicMVBITS // MVBITS
	// Floating point inquiry
	IntrinsicHUGE         // HUGE
	IntrinsicTINY         // TINY
	IntrinsicEPSILON      // EPSILON
	IntrinsicPRECISION    // PRECISION
	IntrinsicRANGE        // RANGE
	IntrinsicRADIX        // RADIX
	IntrinsicDIGITS       // DIGITS
	IntrinsicBIT_SIZE     // BIT_SIZE
	IntrinsicEXPONENT     // EXPONENT
	IntrinsicFRACTION     // FRACTION
	IntrinsicNEAREST      // NEAREST
	IntrinsicRRSPACING    // RRSPACING
	IntrinsicSPACING      // SPACING
	IntrinsicSCALE        // SCALE
	IntrinsicSET_EXPONENT // SET_EXPONENT
	// Kind functions
	IntrinsicKIND               // KIND
	IntrinsicSELECTED_INT_KIND  // SELECTED_INT_KIND
	IntrinsicSELECTED_REAL_KIND // SELECTED_REAL_KIND
	// String functions
	IntrinsicLEN_TRIM // LEN_TRIM
	IntrinsicTRIM     // TRIM
	IntrinsicADJUSTL  // ADJUSTL
	IntrinsicADJUSTR  // ADJUSTR
	IntrinsicREPEAT   // REPEAT
	IntrinsicSCAN     // SCAN
	IntrinsicVERIFY   // VERIFY
	// Pointer inquiry
	IntrinsicASSOCIATED // ASSOCIATED
	// Argument presence
	IntrinsicPRESENT // PRESENT
	// Transfer and conversion
	IntrinsicTRANSFER // TRANSFER
	IntrinsicLOGICAL  // LOGICAL
	// Miscellaneous
	IntrinsicCEILING // CEILING
	IntrinsicFLOOR   // FLOOR
	IntrinsicMODULO  // MODULO
	IntrinsicNULL    // NULL
	fortran90End
)

// Fortran 95 Intrinsics
const (
	fortran95Start    Intrinsic = iota + fortran90End
	IntrinsicCPU_TIME           // CPU_TIME
	fortran95End
)

// Fortran 2003 Intrinsics
const (
	fortran2003Start                  Intrinsic = iota + fortran95End
	IntrinsicMOVE_ALLOC                         // MOVE_ALLOC
	IntrinsicIS_IOSTAT_END                      // IS_IOSTAT_END
	IntrinsicIS_IOSTAT_EOR                      // IS_IOSTAT_EOR
	IntrinsicNEW_LINE                           // NEW_LINE
	IntrinsicCOMMAND_ARGUMENT_COUNT             // COMMAND_ARGUMENT_COUNT
	IntrinsicGET_COMMAND                        // GET_COMMAND
	IntrinsicGET_COMMAND_ARGUMENT               // GET_COMMAND_ARGUMENT
	IntrinsicGET_ENVIRONMENT_VARIABLE           // GET_ENVIRONMENT_VARIABLE
	fortran2003End
)

// Fortran 2008 Intrinsics
const (
	fortran2008Start      Intrinsic = iota + fortran2003End
	IntrinsicACOSH                  // ACOSH
	IntrinsicASINH                  // ASINH
	IntrinsicATANH                  // ATANH
	IntrinsicBESSEL_J0              // BESSEL_J0
	IntrinsicBESSEL_J1              // BESSEL_J1
	IntrinsicBESSEL_JN              // BESSEL_JN
	IntrinsicBESSEL_Y0              // BESSEL_Y0
	IntrinsicBESSEL_Y1              // BESSEL_Y1
	IntrinsicBESSEL_YN              // BESSEL_YN
	IntrinsicERF                    // ERF
	IntrinsicERFC                   // ERFC
	IntrinsicERFC_SCALED            // ERFC_SCALED
	IntrinsicGAMMA                  // GAMMA
	IntrinsicLOG_GAMMA              // LOG_GAMMA
	IntrinsicHYPOT                  // HYPOT
	IntrinsicNORM2                  // NORM2
	IntrinsicPARITY                 // PARITY
	IntrinsicFINDLOC                // FINDLOC
	IntrinsicSTORAGE_SIZE           // STORAGE_SIZE
	fortran2008End
)

func (intr Intrinsic) Version() (year int) {
	switch {
	case intr > fortran2008Start:
		year = 2008
	case intr > fortran2003Start:
		year = 2003
	case intr > fortran95Start:
		year = 95
	case intr > fortran77Start:
		year = 77
	case intr > fortran66Start:
		year = 66
	default:
		year = -1
	}
	return year
}

func (intr Intrinsic) IsValid() bool {
	return intr > fortran66Start && intr < fortran2008End && intr != fortran77Start &&
		intr != fortran90Start && intr != fortran95Start && intr != fortran2003Start && intr != fortran2008Start
}

type VendorIntrinsic int

const (
	vendorUndefined VendorIntrinsic = iota // undefined

	// Common vendor intrinsics (shared across multiple compilers)
	vendorCommonStart
	VendorSYSTEM // SYSTEM
	VendorMALLOC // MALLOC
	VendorFREE   // FREE
	VendorLOC    // LOC
	VendorISNAN  // ISNAN
	VendorSIZEOF // SIZEOF
	// Command-line argument handling (legacy)
	VendorIARGC  // IARGC
	VendorGETARG // GETARG
	VendorNARGS  // NARGS
	// Type conversions
	VendorDFLOAT // DFLOAT
	VendorDCMPLX // DCMPLX
	VendorDREAL  // DREAL
	// Random numbers
	VendorRAND  // RAND
	VendorSRAND // SRAND
	VendorIRAND // IRAND
	// Bit operations (legacy names)
	VendorLSHIFT // LSHIFT
	VendorRSHIFT // RSHIFT
	// Degree-based trigonometric
	VendorSIND   // SIND
	VendorCOSD   // COSD
	VendorTAND   // TAND
	VendorASIND  // ASIND
	VendorACOSD  // ACOSD
	VendorATAND  // ATAND
	VendorATAN2D // ATAN2D
	vendorCommonEnd

	// Intel-specific intrinsics
	vendorIntelStart VendorIntrinsic = iota + vendorCommonEnd
	VendorQCMPLX                     // QCMPLX
	VendorQEXT                       // QEXT
	VendorQFLOAT                     // QFLOAT
	VendorQREAL                      // QREAL
	// String-to-value scanning
	VendorDNUM // DNUM
	VendorINUM // INUM
	VendorJNUM // JNUM
	VendorKNUM // KNUM
	VendorQNUM // QNUM
	VendorRNUM // RNUM
	// Bit manipulation
	VendorIBCHNG // IBCHNG
	VendorISHA   // ISHA
	VendorISHC   // ISHC
	VendorISHL   // ISHL
	VendorIXOR   // IXOR
	// Query functions
	VendorILEN         // ILEN
	VendorMCLOCK       // MCLOCK
	VendorSECNDS       // SECNDS
	VendorCACHESIZE    // CACHESIZE
	VendorEOF          // EOF
	VendorFP_CLASS     // FP_CLASS
	VendorINT_PTR_KIND // INT_PTR_KIND
	// Address functions
	VendorBADDRESS // BADDRESS
	VendorIADDR    // IADDR
	// Random
	VendorRAN  // RAN
	VendorRANF // RANF
	// Cotan
	VendorCOTAND // COTAND
	vendorIntelEnd

	// GNU-specific intrinsics
	vendorGNUStart VendorIntrinsic = iota + vendorIntelEnd
	// Time functions
	VendorETIME   // ETIME
	VendorDTIME   // DTIME
	VendorDSECNDS // DSECNDS
	VendorTIME    // TIME
	VendorTIME8   // TIME8
	VendorCTIME   // CTIME
	VendorFDATE   // FDATE
	VendorGMTIME  // GMTIME
	VendorLTIME   // LTIME
	VendorIDATE   // IDATE
	VendorITIME   // ITIME
	// File operations
	VendorGETCWD // GETCWD
	VendorCHDIR  // CHDIR
	VendorRENAME // RENAME
	VendorUNLINK // UNLINK
	VendorLINK   // LINK
	VendorSYMLNK // SYMLNK
	VendorACCESS // ACCESS
	VendorCHMOD  // CHMOD
	VendorSTAT   // STAT
	VendorLSTAT  // LSTAT
	VendorFSTAT  // FSTAT
	// File I/O
	VendorFSEEK // FSEEK
	VendorFTELL // FTELL
	VendorFNUM  // FNUM
	VendorFGET  // FGET
	VendorFGETC // FGETC
	VendorFPUT  // FPUT
	VendorFPUTC // FPUTC
	VendorFLUSH // FLUSH
	// System information
	VendorHOSTNM // HOSTNM
	VendorGETLOG // GETLOG
	VendorGETPID // GETPID
	VendorGETUID // GETUID
	VendorGETGID // GETGID
	VendorGETENV // GETENV
	VendorPUTENV // PUTENV
	VendorISATTY // ISATTY
	VendorTTYNAM // TTYNAM
	// Process control
	VendorALARM     // ALARM
	VendorSIGNAL    // SIGNAL
	VendorKILL      // KILL
	VendorSLEEP     // SLEEP
	VendorABORT     // ABORT
	VendorEXIT      // EXIT
	VendorBACKTRACE // BACKTRACE
	// Error handling
	VendorPERROR // PERROR
	VendorIERRNO // IERRNO
	VendorGERROR // GERROR
	// Miscellaneous
	VendorUMASK  // UMASK
	VendorLNBLNK // LNBLNK
	VendorCOTAN  // COTAN
	VendorQSORT  // QSORT
	vendorGNUEnd

	// PGI/NVIDIA-specific intrinsics
	vendorPGIStart VendorIntrinsic = iota + vendorGNUEnd
	// Bitwise operations (function form)
	VendorAND   // AND
	VendorOR    // OR
	VendorXOR   // XOR
	VendorCOMPL // COMPL
	VendorEQV   // EQV
	VendorNEQV  // NEQV
	// Type conversions
	VendorZEXT  // ZEXT
	VendorIZEXT // IZEXT
	VendorINT8  // INT8
	VendorJINT  // JINT
	VendorJNINT // JNINT
	VendorKNINT // KNINT
	// Shift
	VendorSHIFT // SHIFT
	vendorPGIEnd
)

// Vendor returns the vendor/origin of the intrinsic.
func (vi VendorIntrinsic) Vendor() string {
	switch {
	case vi > vendorPGIStart && vi < vendorPGIEnd:
		return "PGI"
	case vi > vendorGNUStart && vi < vendorGNUEnd:
		return "GNU"
	case vi > vendorIntelStart && vi < vendorIntelEnd:
		return "Intel"
	case vi > vendorCommonStart && vi < vendorCommonEnd:
		return "Common"
	default:
		return ""
	}
}

// IsValid returns true if the VendorIntrinsic is a valid vendor intrinsic.
func (vi VendorIntrinsic) IsValid() bool {
	return (vi > vendorCommonStart && vi < vendorCommonEnd) ||
		(vi > vendorIntelStart && vi < vendorIntelEnd) ||
		(vi > vendorGNUStart && vi < vendorGNUEnd) ||
		(vi > vendorPGIStart && vi < vendorPGIEnd)
}
