package reedSolomon

import (
	"errors"
	"fmt"
)

var exponents [510]int // # anti-log (exponential) table. The first two elements will always be [GF256int(1), generator]
var logs [256]int      // log table, log[0] is impossible and thus unused
var fcr int            // first consecutive root

// ==========================================
//             Exported Methods
// ==========================================

// InitGaloisFields precomputes the logarithm and anti-log tables for faster computation later, using the provided primitive polynomial.
// prim is the primitive (binary) polynomial. Since it's a polynomial in the binary sense,
// it's only in fact a single galois field value between 0 and 255, and not a list of gf values.
func InitGaloisFields(prim int, firstConsecutiveRoot int) error {
	// fcr
	fcr = firstConsecutiveRoot

	// For each possible value in the galois field 2^8, we will pre-compute the logarithm and anti-logarithm (exponential) of this value
	x := 1
	for i := 0; i < 255; i++ {

		exponents[i] = x // compute exponents for this value and store it in a table
		logs[x] = i      // compute log at the same time

		// TODO: if generator=2 use current method (fastest) if not require fast or slow defined in inputs
		// Slow: Standard carry-less multiplication + modular reduction using an irreducible prime polynomial.
		// Fast: Russian Peasant Multiplication algorithm

		x <<= 1           // Bitwise multiply by 2 (change 1 by another number y to multiply by a power of 2^y)
		if x&0x100 != 0 { // similar to x >= 256, but a lot faster (because 0x100 == 256)
			// Rolls over the value from 256 back to 0 and then up again
			x ^= prim // substract the primary polynomial to the current value (instead of 255, so that we get a unique set made of coprime numbers), this is the core of the tables generation
		}
	}

	// Double the size of the anti-log table so that we don't need to mod 255 later
	copy(exponents[255:510], exponents[0:255]) // optimized (vs for loop)
	// TODO: confirm that this is correct... For loop below is based on reference code
	//for i := 255; i < 512; i++ {
	//	exponents[i] = exponents[i-255]
	//}

	return nil
}

// Decode takes a Reed-Solomon encdoded msg (as an int slice) and corrects the errors and erasures returning the correct string
// Errors cost 2 each: therefore you can correct half the errors as the amount of error correction symbols appended to the message
// IE: if the message has 8 ECC symbols than up to 4 errors can be corrected
// Erasures cost 1 each: therefore you can correct as many erasure as the amount of error correction symbols appended to the message
// IE: if the message has 8 ECC symbols than up to 8 erasure as long as the erased position is provided
func Decode(msg []int, numberEccSymbols int, erasedIndices []int) ([]int, []int, error) {
	// Reed-Solomon main decoding function

	if len(msg) > 255 { // can't decode, message is too big
		return []int{}, []int{}, fmt.Errorf("Message is too long (%d when max is 255)", len(msg))
	}

	msgOut := msg // copy

	// erasures: set them to null bytes for easier decoding (but this is not necessary, they will be corrected anyway,
	// but debugging will be easier with null bytes because the error locator polynomial values will
	// only depend on the errors locations, not their values)

	if len(erasedIndices) == 0 {
		erasedIndices = []int{}
	} else {
		for _, ePos := range erasedIndices {
			msgOut[ePos] = 0
		}
	}

	// check if there are too many erasures to correct (beyond the Singleton bound)
	if len(erasedIndices) > numberEccSymbols {
		return []int{}, []int{}, errors.New("Too many erasures to correct")
	}

	// prepare the syndrome polynomial using only errors (ie: errors = characters that were either replaced by null byte
	// or changed to another character, but we don't know their positions)
	synd := calculateSyndromes(msgOut, numberEccSymbols)

	// check if there's any error/erasure in the input codeword.
	// If not (all syndromes coefficients are 0), then just return the message as-is.
	if isSyndromeClean(synd) {
		m := len(msgOut) - numberEccSymbols
		return msgOut[:m], msgOut[m:], nil // no errors
	}

	// compute the Forney syndromes, which hide the erasures from the original syndrome (so that BM will just have to deal with errors, not erasures)
	fsynd := calcForneySyndromes(synd, erasedIndices, len(msgOut))

	// compute the error locator polynomial using Berlekamp-Massey
	// NOTE: when using forney syndromes DO NOT pass the erasure positions
	errLoc, err := unknownErrorLocator(fsynd, []int{}, numberEccSymbols, len(erasedIndices))
	if err != nil {
		return []int{}, []int{}, err
	}

	// locate the message errors using Chien search (or brute-force search)
	errPos, err := findErrors(sliceIntReverse(errLoc), len(msgOut))

	if err != nil {
		return []int{}, []int{}, err // error location failed
	}
	if len(errPos) == 0 && len(erasedIndices) == 0 {
		return []int{}, []int{}, errors.New("Could not calculate error positions") // error location failed
	}

	// Find errors values and apply them to correct the message
	// compute errata evaluator and errata magnitude polynomials, then correct errors and erasures
	msgOut = correctErrors(msgOut, synd, append(erasedIndices, errPos...)) // note that we here use the original syndrome, not the forney syndrome

	// (because we will correct both errors and erasures, so we need the full syndrome)
	// check if the final message is fully repaired
	synd = calculateSyndromes(msgOut, numberEccSymbols)

	if !isSyndromeClean(synd) {
		return []int{}, []int{}, errors.New("Could not correct message") // message could not be repaired
	}

	// return the successfully decoded message
	m := len(msgOut) - numberEccSymbols
	return msgOut[:m], msgOut[m:], nil // also return the corrected ecc block so that the user can check()
}

// ==========================================
//             Unexported Methods
// ==========================================

// Given the received codeword msg and the number of error correcting symbols (nsym), this computes the syndromes polynomial.
// Mathematically, it's essentially equivalent to a Fourrier Transform (Chien search being the inverse).
func calculateSyndromes(msg []int, nsym int) []int {

	synd := make([]int, nsym) //  Make an empty syndrome slice <--- this is c
	//msg = append([]int{0}, msg...)
	for i := 0; i < nsym; i++ {
		synd[i] = gfPolynomialEval(msg, gfPower(2, i+fcr)) // TODO: +1 is first consecutive root? might need to change this value for some generators
	}

	// Here we append a 0 coefficient for the lowest degree (the constant). This effectively shifts the
	// syndrome, and will shift every computations depending on the syndromes (such as the errors locator polynomial,
	// errors evaluator polynomial, etc. but not the errors positions).

	// This is not necessary, you can adapt subsequent computations to start from 0 instead of skipping the first
	// iteration (ie, the often seen range(1, n-k+1))
	synd = append([]int{0}, synd...)
	return synd
}

// Check if the syndrom is all 0's or if there are corrections to be made
func isSyndromeClean(synd []int) bool {
	for _, v := range synd {
		if v > 0 {
			return false
		}
	}
	return true
}

// Compute the erasures/errors/errata locator polynomial from the erasures/errors/errata positions
// (the positions must be relative to the x coefficient, eg: "hello worldxxxxxxxxx" is tampered to "h_ll_ worldxxxxxxxxx"
// with xxxxxxxxx being the ecc of length n-k=9, here the string positions are [1, 4], but the coefficients are reversed
// since the ecc characters are placed as the first coefficients of the polynomial, thus the coefficients of the
// erased characters are n-1 - [1, 4] = [18, 15] = erasures_loc to be specified as an argument.
func calcErrorLocatorPolynomial(errorPositions []int) []int {

	errorLocatorPolynomial := []int{1} // just to init because we will multiply, so it must be 1 so that the multiplication starts correctly without nulling any term
	// erasures_loc = product(1 - x*alpha**i) for i in erasures_pos and where alpha is the alpha chosen to evaluate polynomials.

	for _, p := range errorPositions {
		errorLocatorPolynomial = gfPolynomialMultiplication(errorLocatorPolynomial, gfPolynomialAddition([]int{1}, []int{gfPower(2, p), 0}))
	}
	return errorLocatorPolynomial
}

// Compute the error (or erasures if you supply sigma=erasures locator polynomial, or errata) evaluator polynomial Omega
// from the syndrome and the error/erasures/errata locator Sigma.
func calcErrorPolynomial(synd, errorLocatorPolynomial []int, nsym int) []int {

	// Omega(x) = [ Synd(x) * Error_loc(x) ] mod x^(n-k+1)
	placeholder := make([]int, nsym+1)
	placeholder = append([]int{1}, placeholder...)

	_, remainder := gfPolynomialDivision(gfPolynomialMultiplication(synd, errorLocatorPolynomial), placeholder) // first multiply syndromes * errata_locator, then do a polynomial division to truncate the polynomial to the required length

	//remainder := gfPolynomialMultiplication(synd, errorLocatorPolynomial) // first multiply the syndromes with the errata locator polynomial
	// remainder = remainder[len(remainder)-(nsym+1):]                  // then slice the list to truncate it (which represents the polynomial), which

	return remainder
}

// Find error/errata locator and evaluator polynomials with Berlekamp-Massey algorithm
// NOTE: If forney syndromes are provided then use and emtpy erasureLoc slice as the erasures are not part of the forney syndrome
func unknownErrorLocator(synd, erasureLoc []int, nsym, erasureCount int) ([]int, error) {

	// The idea is that BM will iteratively estimate the error locator polynomial.
	// To do this, it will compute a Discrepancy term called Delta, which will tell us if the error locator polynomial needs an update or not
	// (hence why it's called discrepancy: it tells us when we are getting off board from the correct value).
	errLoc := []int{1} // This is the main variable we want to fill, also called Sigma in other notations or more formally the errors/errata locator polynomial.
	oldLoc := []int{1} // BM is an iterative algorithm, and we need the errata locator polynomial of the previous iteration in order to update other necessary variables.

	// Init the polynomials
	if len(erasureLoc) > 0 { // if the erasure locator polynomial is supplied, we start with its value, so that we include erasures in the final locator polynomial
		errLoc = erasureLoc
		oldLoc = erasureLoc
	}

	// Fix the syndrome shifting: when computing the syndrome, some implementations may prepend a 0 coefficient for the lowest degree term (the constant). This is a case of syndrome shifting, thus the syndrome will be bigger than the number of ecc symbols (I don't know what purpose serves this shifting). If that's the case, then we need to account for the syndrome shifting when we use the syndrome such as inside BM, by skipping those prepended coefficients.
	// Another way to detect the shifting is to detect the 0 coefficients: by definition, a syndrome does not contain any 0 coefficient (except if there are no errors/erasures, in this case they are all 0). This however doesn't work with the modified Forney syndrome, which set to 0 the coefficients corresponding to erasures, leaving only the coefficients corresponding to errors.
	syndShift := 0
	if len(synd) > nsym {
		syndShift = len(synd) - nsym
	}

	for i := 0; i < nsym-erasureCount; i++ { // generally: nsym-erase_count == len(synd), except when you input a partial erase_loc and using the full syndrome instead of the Forney syndrome, in which case nsym-erase_count is more correct (len(synd) will fail badly with IndexError).
		var K int
		if len(erasureLoc) > 0 { // if an erasures locator polynomial was provided to init the errors locator polynomial, then we must skip the FIRST erase_count iterations (not the last iterations, this is very important!)
			K = erasureCount + i + syndShift
		} else { // if erasures locator is not provided, then either there's no erasures to account or we use the Forney syndromes, so we don't need to use erase_count nor erase_loc (the erasures have been trimmed out of the Forney syndromes).
			K = i + syndShift
		}

		// Compute the discrepancy Delta
		// Here is the close-to-the-books operation to compute the discrepancy Delta: it's a simple polynomial multiplication of error locator with the syndromes, and then we get the Kth element.
		// delta = gf_poly_mul(err_loc[::-1], synd)[K] // theoretically it should be gfPolynomialAdd(synd[::-1], [1])[::-1] instead of just synd, but it seems it's not absolutely necessary to correctly decode.
		// But this can be optimized: since we only need the Kth element, we don't need to compute the polynomial multiplication for any other element but the Kth. Thus to optimize, we compute the polymul only at the item we need, skipping the rest (avoiding a nested loop, thus we are linear time instead of quadratic).
		// This optimization is actually described in several figures of the book "Algebraic codes for data transmission", Blahut, Richard E., 2003, Cambridge university press.
		delta := synd[K]

		for j := 1; j < len(errLoc); j++ {
			delta ^= gfMultiplication(errLoc[len(errLoc)-(j+1)], synd[K-j]) // delta is also called discrepancy. Here we do a partial polynomial multiplication (ie, we compute the polynomial multiplication only for the term of degree K). Should be equivalent to brownanrs.polynomial.mul_at().
		}

		// Shift polynomials to compute the next degree
		oldLoc = append(oldLoc, 0)

		// Iteratively estimate the errata locator and evaluator polynomials
		if delta != 0 { // Update only if there's a discrepancy
			if len(oldLoc) > len(errLoc) { // Rule B (rule A is implicitly defined because rule A just says that we skip any modification for this iteration)
				// Computing errata locator polynomial Sigma
				newLoc := gfPolynomialScale(oldLoc, delta)
				oldLoc = gfPolynomialScale(errLoc, gfInverse(delta)) // effectively we are doing err_loc * 1/delta = err_loc // delta
				errLoc = newLoc
			}

			// Update with the discrepancy
			errLoc = gfPolynomialAddition(errLoc, gfPolynomialScale(oldLoc, delta))
		}
	}

	for len(errLoc) > 0 && errLoc[0] == 0 {
		errLoc = errLoc[1:] // drop leading 0s, else errs will not be of the correct size
	}

	// Check if the result is correct, that there's not too many errors to correct
	errs := len(errLoc) - 1

	// Max error score is equal to the number of ECC symbols (nsym)
	// Errors cost 2 each
	// Erasures (with provided location) cost 1 each
	// If non forney syndromes are used the erasures will be treated as errors and
	// And will cost 2.
	// TODO: better way to set this up to that erasure counts are sent for forney syndromes
	if ((errs-len(erasureLoc))*2 + (erasureCount - len(erasureLoc))) > nsym {
		return []int{}, fmt.Errorf("Too many errors to correct: %d of max %d (Found at least %d errors and %d erasures)", (errs)*2+erasureCount, nsym, errs, erasureCount) // too many errors to correct
	}

	return errLoc, nil
}

func correctErrors(msgIn, synd, errPos []int) []int {
	// errPos is a list of the positions of the errors/erasures/errata
	// Forney algorithm, computes the values (error magnitude) to correct the input message.

	// calculate errata locator polynomial to correct both errors and erasures (by combining the errors positions given by the error locator polynomial found by BM with the erasures positions given by caller)
	coefPos := make([]int, len(errPos))

	for i, p := range errPos {
		// need to convert the positions to coefficients degrees for the errata locator algo to work
		//(eg: instead of [0, 1, 2] it will become [len(msg)-1, len(msg)-2, len(msg) -3])
		coefPos[i] = len(msgIn) - 1 - p
	}

	errorLocatorPolynomial := calcErrorLocatorPolynomial(coefPos)
	// calculate errata evaluator polynomial (often called Omega or Gamma in academic papers)

	errorPolynomial := calcErrorPolynomial(sliceIntReverse(synd), errorLocatorPolynomial, len(errorLocatorPolynomial)-1)
	//errorPolynomial = sliceIntReverse(errorPolynomial) // reverse the order

	// Second part of Chien search to get the error location polynomial X from the error positions in errPos (the roots of the error locator polynomial, ie, where it evaluates to 0)
	locationPolynomial := []int{} // will store the position of the errors
	for i := 0; i < len(coefPos); i++ {
		l := 255 - coefPos[i]
		locationPolynomial = append(locationPolynomial, gfPower(2, -l))
	}

	// Forney algorithm: compute the magnitudes
	E := forney(msgIn, errorPolynomial, locationPolynomial, errPos, fcr)

	// Apply the correction of values to get our message corrected! (note that the ecc bytes also gets corrected!)
	// (this isn't the Forney algorithm, we just apply the result of decoding here)
	msgIn = gfPolynomialAddition(msgIn, E) // equivalent to Ci = Ri - Ei where Ci is the correct message, Ri the received (senseword) message, and Ei the errata magnitudes (minus is replaced by XOR since it's equivalent in GF(2^p)). So in fact here we substract from the received message the errors magnitude, which logically corrects the value to what it should be.

	return msgIn
}

// Find the roots (ie, where evaluation = zero) of error polynomial by brute-force trial, this is a sort of Chien's search
// (but less efficient, Chien's search is a way to evaluate the polynomial such that each evaluation only takes constant time).
func findErrors(errLoc []int, msgLen int) ([]int, error) {

	// Find the roots (ie, where evaluation = zero) of error polynomial by brute-force trial, this is a sort of Chien's search
	// (but less efficient, Chien's search is a way to evaluate the polynomial such that each evaluation only takes constant time).
	errs := len(errLoc) - 1
	errPos := []int{}

	for i := 0; i < msgLen; i++ { // normally we should try all 2^8 possible values, but here we optimize to just check the interesting symbols
		if gfPolynomialEval(errLoc, gfPower(2, i)) == 0 { // It's a 0? Bingo, it's a root of the error locator polynomial,
			// in other terms this is the location of an error
			errPos = append(errPos, msgLen-1-i)
		}
	}

	// Sanity check: the number of errors/errata positions found should be exactly the same as the length of the errata locator polynomial
	if len(errPos) != errs {
		// couldn't find error locations
		return []int{}, errors.New("too many (or few) errors found by Chien Search for the errata locator polynomial")
	}

	return errPos, nil
}

func calcForneySyndromes(synd, pos []int, msgLen int) []int {
	// Compute Forney syndromes, which computes a modified syndromes to compute only errors (erasures are trimmed out).
	// Do not confuse this with Forney algorithm, which allows to correct the message based on the location of errors.

	// prepare the coefficient degree positions (instead of the erasures positions)
	erasePosReversed := make([]int, len(pos))
	for i, p := range pos {
		// need to convert the positions to coefficients degrees for the errata locator algo to work (eg: instead of [0, 1, 2] it will become [len(msg)-1, len(msg)-2, len(msg) -3])
		erasePosReversed[i] = msgLen - 1 - p
	}

	// Optimized method, all operations are inlined
	fsynd := make([]int, len(synd)-1)
	copy(fsynd[:], synd[1:]) // make a copy and trim the first coefficient which is always 0 by definition

	for i := 0; i < len(pos); i++ {
		x := gfPower(2, erasePosReversed[i])
		for j := 0; j < len(fsynd)-1; j++ {
			fsynd[j] = gfMultiplication(fsynd[j], x) ^ fsynd[j+1]
		}
	}

	// Equivalent, theoretical way of computing the modified Forney syndromes: fsynd = (erase_loc * synd) % x^(n-k)
	// See Shao, H. M., Truong, T. K., Deutsch, L. J., & Reed, I. S. (1986, April). A single chip VLSI Reed-Solomon decoder. In Acoustics, Speech, and Signal Processing, IEEE International Conference on ICASSP'86. (Vol. 11, pp. 2151-2154). IEEE.ISO 690
	//erase_loc = calcErrorLocatorPolynomial(erase_pos_reversed, generator=generator) // computing the erasures locator polynomial
	//fsynd = gf_poly_mul(erase_loc[::-1], synd[1:]) // then multiply with the syndrome to get the untrimmed forney syndrome
	//fsynd = fsynd[len(pos):] // then trim the first erase_pos coefficients which are useless. Seems to be not necessary, but this reduces the computation time later in BM (thus it's an optimization).

	return fsynd
}
