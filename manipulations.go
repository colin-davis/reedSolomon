package reedSolomon

import (
	"errors"
)

// ***************************
// Single value Manipulations
// ***************************

// GF sub and add are the same (both are XOR functions)
func gfAddition(x, y int) int {
	return x ^ y
}
func gfSubtraction(x, y int) int {
	return gfAddition(x, y)
}

// Use the exponents table to lookup multiplication value
func gfMultiplication(x, y int) int {
	if x == 0 || y == 0 {
		return 0
	}
	return exponents[logs[x]+logs[y]]
}

func gfDivision(x, y int) (int, error) {
	if y == 0 {
		return -1, errors.New("Zero Division Error")
	}
	if x == 0 {
		return 0, nil
	}
	return exponents[(logs[x]+255-logs[y])%255], nil
}

func gfPower(x, power int) int {

	index := (logs[x] * power) % 255

	// If the index is positive get it
	if index >= 0 {
		return exponents[index]
	}
	// If the index is negative simulate a rollover in the LUT
	return exponents[len(exponents)+index]
}

func gfInverse(x int) int {
	return exponents[255-logs[x]] // gfInverse(x) == gfDivision(1, x)
}

// ***************************
// Polynomial Manipulations
// ***************************

func gfPolynomialScale(p []int, x int) []int {
	// TODO: manipulate and return p?
	r := make([]int, len(p)) // make a destination array

	for i := 0; i < len(p); i++ {
		r[i] = gfMultiplication(p[i], x)
	}
	return r
}

func gfPolynomialDivision(dividend, divisor []int) ([]int, []int) {
	// Fast polynomial division by using Extended Synthetic Division and optimized for GF(2^p) computations
	// (doesn't work with standard polynomials outside of this galois field, see the Wikipedia article for generic algorithm).
	// CAUTION: this function expects polynomials to follow the opposite convention at decoding:
	// the terms must go from the biggest to lowest degree (while most other functions here expect
	// a list from lowest to biggest degree). eg: 1 + 2x + 5x^2 = [5, 2, 1], NOT [1, 2, 5]

	msgOut := make([]int, len(dividend))
	copy(msgOut, dividend) // Copy the dividend

	//normalizer = divisor[0] // precomputing for performance
	for i := 0; i < len(dividend)-(len(divisor)-1); i++ {
		// msg_out[i] /= normalizer // for general polynomial division (when polynomials are non-monic), the usual way of using
		// synthetic division is to divide the divisor g(x) with its leading coefficient, but not needed here.

		coef := msgOut[i] // precaching
		if coef != 0 {    // log(0) is undefined, so we need to avoid that case explicitly (and it's also a good optimization).
			for j := 1; j < len(divisor); j++ { // in synthetic division, we always skip the first coefficient of the divisior,
				// because it's only used to normalize the dividend coefficient
				if divisor[j] != 0 { // log(0) is undefined
					msgOut[i+j] ^= gfMultiplication(divisor[j], coef) // equivalent to the more mathematically correct
					// (but xoring directly is faster): msg_out[i + j] += -divisor[j] * coef
				}
			}
		}
	}

	// The resulting msg_out contains both the quotient and the remainder, the remainder being the size of the divisor
	// (the remainder has necessarily the same degree as the divisor -- not length but degree == length-1 -- since it's
	// what we couldn't divide from the dividend), so we compute the index where this separation is, and return the quotient and remainder.
	separator := len(divisor) - 1
	return msgOut[:separator], msgOut[separator:] // return quotient, remainder.
}

func gfPolynomialEval(poly []int, x int) int {
	// Evaluates a polynomial in GF(2^p) given the value for x .This is based on Horner's scheme for maximum efficiency.
	y := poly[0]

	for i := 1; i < len(poly); i++ {
		y = gfMultiplication(y, x) ^ poly[i]
	}

	return y
}

func gfPolynomialMultiplication(p, q []int) []int {
	// Multiply two polynomials, inside Galois Field
	// Pre-allocate the result array
	r := make([]int, len(p)+len(q)-1)
	// Compute the polynomial multiplication (just like the outer product of two vectors,
	// we multiply each coefficients of p with all coefficients of q)
	for j := 0; j < len(q); j++ {
		for i := 0; i < len(p); i++ {
			r[i+j] ^= gfMultiplication(p[i], q[j]) // equivalent to: r[i + j] = gfAddition(r[i+j], gfMultiplication(p[i], q[j]))
		}
	}
	// -- you can see it's your usual polynomial multiplication
	return r
}

func gfPolynomialAddition(p, q []int) []int {

	var r []int

	if len(p) > len(q) {
		r = make([]int, len(p))
	} else {
		r = make([]int, len(q))
	}

	for i := 0; i < len(p); i++ {
		r[i+len(r)-len(p)] = p[i]
	}
	for i := 0; i < len(q); i++ {
		r[i+len(r)-len(q)] ^= q[i]
	}
	return r
}
