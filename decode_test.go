package reedSolomon

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	InitGaloisFields(301) // test are written to work with 301
	// 301 for datamatrix
	// 285 for qr codes

	os.Exit(m.Run())
}

var encodedMsg = []int{68, 90, 46, 145, 46, 131, 153, 53, 32, 43, 239, 193, 240, 155, 85, 215, 63, 202}

func TestInitGaloisFields(t *testing.T) {
	t.Log("Testing precomputing galois field tables with primative value 301")

	// Test exponents table
	if len(exponents) != 510 {
		t.Errorf("Expected exponents to be of length 510, but it was %d instead.", len(exponents))
	}
	if exponents[10] != 180 {
		t.Errorf("Expected exponent at index 10 to be 180, but it was %d instead.", exponents[10])
	}

	// Test logs table
	if len(logs) != 256 {
		t.Errorf("Expected logs to be of length 255, but it was %d instead.", len(logs))
	}
	if logs[10] != 226 {
		t.Errorf("Expected exponent at index 10 to be 226, but it was %d instead.", logs[10])
	}
}

func TestCalculateSyndromesCorrect(t *testing.T) {
	t.Log("Testing calculate syndroms")

	// Test a fully correct msg
	msg := encodedMsg
	nsym := 8

	expected := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	resp := calculateSyndromes(msg, nsym)

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

// Test a message that has a correctable amount of errors
func TestCalculateSyndromesErrors(t *testing.T) {
	t.Log("Testing calculate syndroms with correctable errors")

	// Test a correctable msg
	msg := encodedMsg

	//include errors
	msg[6] = 11
	msg[8] = 12
	msg[13] = 125

	nsym := 8

	expected := []int{0, 251, 151, 100, 105, 36, 252, 184, 46, 161, 32}
	resp := calculateSyndromes(msg, nsym)

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

func TestIsSyndromeClean(t *testing.T) {
	t.Log("Testing if syndrome is clean (all 0's)")

	// True
	synd := []int{0, 0, 0, 0, 0}
	resp := isSyndromeClean(synd)

	if !resp {
		t.Error("Syndrome is clean but returned false")
	}

	// False
	synd = []int{0, 0, 0, 151, 0}
	resp = isSyndromeClean(synd)

	if resp {
		t.Error("Syndrome is not clean but returned true")
	}
}

func TestCalcErrorLocatorPolynomial(t *testing.T) {
	t.Log("Test calculating error polynomial from error position data")

	errorPositions := []int{5, 10, 12}

	expected := []int{157, 152, 30, 1}
	resp := calcErrorLocatorPolynomial(errorPositions)

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

func TestCalcErrorPolynomial(t *testing.T) {
	t.Log("Test computing the error polynomial")

	synd := []int{0, 251, 151, 100, 105, 36, 252, 184, 46, 161}
	errataPolynomial := []int{157, 152, 30, 1}

	expected := []int{1, 161, 220, 174, 205, 73, 201, 199, 93, 1, 161}
	resp := calcErrorPolynomial(synd, errataPolynomial, 10)

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

func TestUnknownErrorLocatorWithEarasures(t *testing.T) {
	t.Log("Test locating errors with erasures")

	// Message with 2 errors
	// msgIn := encodedMsg
	// msgIn[4] = 10
	// msgIn[7] = 10
	// synd := calculateSyndromes(msgIn, 8)

	synd := []int{0, 211, 85, 186, 161, 174, 206, 94, 238}
	erasurePositions := []int{5, 10}
	nsym := 8
	erasureCount := len(erasurePositions)

	expected := []int{49, 209, 106, 71, 10}
	resp, err := unknownErrorLocator(synd, erasurePositions, nsym, erasureCount)

	if err != nil {
		t.Error(err)
	}

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

func TestUnknownErrorLocatorWithoutEarasures(t *testing.T) {
	t.Log("Test locating errors without erasures")

	synd := []int{0, 251, 151, 100, 105, 36, 252, 184, 46, 161}
	errorPositions := []int{}
	nsym := 8
	erasureCount := 0

	expected := []int{54, 38, 15, 1}
	resp, err := unknownErrorLocator(synd, errorPositions, nsym, erasureCount)

	if err != nil {
		t.Error(err)
	}

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

func TestUnknownErrorLocatorTooManyErrors(t *testing.T) {
	t.Log("Test locating errors with too many errors to correct")

	// Message with 2 errors
	// msgIn := encodedMsg
	// msgIn[4] = 10 //cost 2 each
	// msgIn[7] = 10
	// msgIn[9] = 10
	// synd := calculateSyndromes(msgIn, 8)
	// fsynd := calcForneySyndromes(synd, []int{0}, len(msgIn))
	// log.Println(fsynd)

	fsynd := []int{169, 221, 32, 205, 193, 85, 88, 249}
	errasureLocator := []int{2, 5, 10} //cost 1 each (because we are using forney syndrome)
	nsym := 8
	erasureCount := len(errasureLocator)

	_, err := unknownErrorLocator(fsynd, []int{}, nsym, erasureCount)

	if err.Error() != "Too many errors to correct: 9 of max 8 (Found at least 3 errors and 3 erasures)" {
		t.Error("Should fail with too many errors to correct")
	}
}

func TestUnknownErrorLocatorLeadingZero(t *testing.T) {
	t.Log("Test locating errors with too many errors to correct")

	// Message with 2 errors
	// msgIn := encodedMsg
	// msgIn[4] = 10 //cost 2 each
	// msgIn[7] = 10
	// msgIn[9] = 10
	// synd := calculateSyndromes(msgIn, 8)
	// fsynd := calcForneySyndromes(synd, []int{0}, len(msgIn))
	// log.Println(fsynd)

	fsynd := []int{0, 0, 0, 0, 0, 0, 0, 0}
	errasureLocator := []int{} //cost 1 each (because we are using forney syndrome)
	nsym := 8
	erasureCount := len(errasureLocator)

	unknownErrorLocator(fsynd, []int{}, nsym, erasureCount)

}

func TestCorrectErrors(t *testing.T) {
	t.Log("Test locating errors of msgIn")

	msgIn := []int{68, 90, 46, 145, 46, 131, 11, 53, 12, 43, 239, 193, 240, 125, 85, 215, 63, 202}
	synd := []int{0, 193, 251, 151, 100, 105, 36, 252, 184, 46, 161} // confirmed in python
	errPos := []int{6, 8, 13}

	expected := []int{68, 90, 46, 145, 46, 131, 211, 53, 90, 43, 239, 193, 240, 139, 85, 215, 63, 202}
	resp := correctErrors(msgIn, synd, errPos)

	for i, r := range resp {
		if r != expected[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expected[i], r)
		}
	}
}

func TestCorrectMessage(t *testing.T) {
	t.Log("Test Correcting Message")

	msgIn := []int{68, 90, 46, 145, 46, 131, 153, 53, 32, 43, 239, 193, 240, 155, 85, 215, 63, 202}

	expectedCorrectedMsg := []int{68, 90, 46, 145, 46, 131, 153, 53, 32, 43}
	expectedCorrectedEcc := []int{239, 193, 240, 155, 85, 215, 63, 202}
	correctedMsg, correctedEcc, err := Decode(msgIn, 8, []int{})

	if err != nil {
		t.Error(err)
	}

	// Message
	for i, r := range correctedMsg {
		if r != expectedCorrectedMsg[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expectedCorrectedMsg[i], r)
		}
	}

	// ECC
	for i, r := range correctedEcc {
		if r != expectedCorrectedEcc[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expectedCorrectedEcc[i], r)
		}
	}
}

func TestCorrectMessageWithErrors(t *testing.T) {
	t.Log("Test Correcting Message with errors")

	msgIn := []int{68, 90, 46, 145, 46, 131, 11, 53, 12, 43, 239, 193, 240, 125, 85, 215, 63, 202}

	expectedCorrectedMsg := []int{68, 90, 46, 145, 46, 131, 153, 53, 32, 43}
	expectedCorrectedEcc := []int{239, 193, 240, 155, 85, 215, 63, 202}
	correctedMsg, correctedEcc, err := Decode(msgIn, 8, []int{})

	if err != nil {
		t.Error(err)
	}

	// Message
	for i, r := range correctedMsg {
		if r != expectedCorrectedMsg[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expectedCorrectedMsg[i], r)
		}
	}

	// ECC
	for i, r := range correctedEcc {
		if r != expectedCorrectedEcc[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expectedCorrectedEcc[i], r)
		}
	}
}

func TestCorrectMessageWithTooManyErrors(t *testing.T) {
	t.Log("Test Correcting Message With Too Many Errors")

	// Max errors: nsym/2 = 8/2 = 4
	// number of errors: 5
	// number of erasures = 0
	msgIn := []int{68, 90, 25, 145, 46, 131, 11, 53, 12, 43, 239, 11, 240, 125, 85, 215, 63, 202}
	erasurePos := []int{}

	_, _, err := Decode(msgIn, 8, erasurePos)

	if err.Error() != "too many (or few) errors found by Chien Search for the errata locator polynomial" {
		t.Error("Should have stated too many errors")
	}
}

func TestCorrectMessageGibberish(t *testing.T) {
	t.Log("Test Correcting Message With Too Many Errors")

	// Max errors: nsym/2 = 8/2 = 4
	msgIn := encodedMsg
	msgIn[2] = 10
	msgIn[5] = 10
	msgIn[6] = 10
	msgIn[11] = 10
	//msgIn[14] = 10

	erasurePos := []int{0}

	_, _, err := Decode(msgIn, 8, erasurePos)

	if err.Error() != "Too many errors to correct: 9 of max 8 (Found at least 4 errors and 1 erasures)" {
		t.Error("Should have stated too many errors")
	}
}

func TestCorrectMessageWithErasure(t *testing.T) {
	t.Log("Test Correcting Message With Erasure")

	// Max Erasure = 8 (same as nsym)
	// number of errors = 0
	// number of erasures = 3
	msgIn := []int{68, 90, 46, 145, 46, 131, 153, 53, 32, 43, 239, 193, 240, 155, 85, 215, 63, 202}
	erasurePos := []int{4, 6, 8}

	expectedCorrectedMsg := []int{68, 90, 46, 145, 46, 131, 153, 53, 32, 43}
	expectedCorrectedEcc := []int{239, 193, 240, 155, 85, 215, 63, 202}
	correctedMsg, correctedEcc, err := Decode(msgIn, 8, erasurePos)

	if err != nil {
		t.Error(err)
	}

	// Message
	for i, r := range correctedMsg {
		if r != expectedCorrectedMsg[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expectedCorrectedMsg[i], r)
		}
	}

	// ECC
	for i, r := range correctedEcc {
		if r != expectedCorrectedEcc[i] {
			t.Errorf("Response at index %d was expected to be %d, but it was %d instead.", i, expectedCorrectedEcc[i], r)
		}
	}
}

func TestCorrectMessageTooLong(t *testing.T) {
	t.Log("Test Correcting Message that is longer then the max 255")

	msgIn := make([]int, 300)

	_, _, err := Decode(msgIn, 8, []int{})

	if err.Error() != "Message is too long (300 when max is 255)" {
		t.Error("Should have stated that the message was too long")
	}
}

func TestCorrectTooManyErasures(t *testing.T) {
	t.Log("Test Correcting Message with too many erasures")

	// max erasures should be 8
	// we will test with 9
	msgIn := []int{68, 90, 46, 145, 46, 131, 153, 53, 32, 43, 239, 193, 240, 155, 85, 215, 63, 202}
	erasurePos := []int{4, 6, 8, 9, 10, 11, 12, 13, 14}

	_, _, err := Decode(msgIn, 8, erasurePos)

	if err.Error() != "Too many erasures to correct" {
		t.Error("Should have stated: Too many erasures to correct")
	}
}
