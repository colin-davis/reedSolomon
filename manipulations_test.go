package reedSolomon

import (
	"testing"
)

func TestGfAddition(t *testing.T) {
	t.Log("Testing Galois Field Addition")
	expected := 12

	resp := gfAddition(2, 14)

	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}
}

func TestGfSubtraction(t *testing.T) {
	t.Log("Testing Galois Field Subtraction")
	expected := 12

	resp := gfSubtraction(2, 14)

	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}
}

func TestGfMultiplication(t *testing.T) {
	t.Log("Testing Galois Field Multiplication")

	expected := 28
	resp := gfMultiplication(2, 14)

	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}

	expected = 0
	resp = gfMultiplication(0, 14)

	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}
}

func TestGfDivision(t *testing.T) {
	t.Log("Testing Galois Field Division")

	expected := 197
	resp, err := gfDivision(2, 14)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}

	// divide by 0
	expected = -1
	resp, err = gfDivision(2, 0)

	if err == nil {
		t.Error("Should return divide by 0 error")
	}
	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}

	// numerator 0
	expected = 0
	resp, err = gfDivision(0, 14)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}
}

func TestGfPower(t *testing.T) {
	t.Log("Testing Galois Field Power")

	expected := 17
	resp := gfPower(5, 2)

	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}

	expected = 52
	resp = gfPower(15, -1)

	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}
}

func TestGfInverse(t *testing.T) {
	t.Log("Testing Galois Field Inverse")

	expected := 150
	resp := gfInverse(2)

	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}
}

// ***************************
// Polynomial Manipulations
// ***************************

func TestGfPolynomialScale(t *testing.T) {
	t.Log("Testing Galois Field Polynomial Scale")

	expected := []int{6, 17, 41, 5}
	resp := gfPolynomialScale([]int{2, 15, 252, 3}, 3)

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

func TestGfPolynomialDivision(t *testing.T) {
	t.Log("Testing Galois Field Polynomial Division")

	eQuotient := []int{5, 56, 231, 148}
	eRemainder := []int{172}
	quotient, remainder := gfPolynomialDivision([]int{5, 11, 156, 163}, []int{2, 15, 252, 3})

	// Quotients
	for i, q := range quotient {
		if q != eQuotient[i] {
			t.Errorf("Quotient response at index %d was expected to be %d, but it was %d instead.", i, eQuotient[i], q)
		}
	}

	// Remainders
	for i, r := range remainder {
		if r != eRemainder[i] {
			t.Errorf("Remainder response at index %d was expected to be %d, but it was %d instead.", i, eRemainder[i], r)
		}
	}
}

func TestGfPolynomialEval(t *testing.T) {
	t.Log("Testing Galois Field Polynomial Eval")

	expected := 7
	resp := gfPolynomialEval([]int{2, 15, 252, 3}, 3)

	if resp != expected {
		t.Errorf("Expected respsonse to be %d, but it was %d instead.", expected, resp)
	}
}

func TestGfPolynomialMultiplication(t *testing.T) {
	t.Log("Testing Galois Field Polynomial Multiplication")

	expected := []int{10, 37, 7, 153, 10, 70, 200}
	resp := gfPolynomialMultiplication([]int{5, 11, 156, 163}, []int{2, 15, 252, 3})

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

func TestGfPolynomialAddition(t *testing.T) {
	t.Log("Testing Galois Field Polynomial Addition")

	expected := []int{7, 4, 96, 160}
	resp := gfPolynomialAddition([]int{5, 11, 156, 163}, []int{2, 15, 252, 3})

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}

	// Test with mismatched sized slices
	expected = []int{5, 9, 147, 95, 196}
	resp = gfPolynomialAddition([]int{5, 11, 156, 163, 199}, []int{2, 15, 252, 3})

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}
