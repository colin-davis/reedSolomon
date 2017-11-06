package reedSolomon

import (
	"testing"
)

func TestForney(t *testing.T) {
	t.Log("Test Forney Algorithm")

	msgIn := []int{68, 90, 46, 145, 46, 131, 11, 53, 12, 43, 239, 193, 240, 125, 85, 215, 63, 202}
	errorLocatorPolynomial := []int{157, 152, 30}
	errorPolynomial := []int{245, 112, 220, 174, 205, 73, 201, 199, 93, 1, 161}
	errPos := []int{6, 8, 13}
	fcr = 1

	expected := []int{0, 0, 0, 0, 0, 0, 244, 0, 223, 0, 0, 0, 0, 16, 0, 0, 0, 0}
	resp := forney(msgIn, errorPolynomial, errorLocatorPolynomial, errPos, fcr)

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}
