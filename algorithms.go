package reedSolomon

func forney(msgIn, errorPolynomial, locationPolynomial, errPos []int, fcr int) []int {
	E := make([]int, len(msgIn)) // will store the values that need to be corrected (subtracted) to the message containing errors. This is sometimes called the error magnitude polynomial.

	for i, location := range locationPolynomial {

		locationInverse := gfInverse(location)

		// Compute the formal derivative of the error locator polynomial (see Blahut, Algebraic codes for data transmission, pp 196-197).
		// the formal derivative of the errata locator is used as the denominator of the Forney Algorithm, which simply says that the ith error value is
		//given by error_evaluator(gfInverse(location)) / error_locator_derivative(gfInverse(location)). See Blahut, Algebraic codes for data transmission, pp 196-197.
		errorLocatorPrimeTemp := []int{}

		for j := 0; j < len(locationPolynomial); j++ {
			if j != i {
				errorLocatorPrimeTemp = append(errorLocatorPrimeTemp, gfSubtraction(1, gfMultiplication(locationInverse, locationPolynomial[j])))
			}
		}

		// compute the product, which is the denominator of the Forney algorithm (errata locator derivative)
		errorLocatorPrime := 1

		for _, coef := range errorLocatorPrimeTemp {
			errorLocatorPrime = gfMultiplication(errorLocatorPrime, coef)
		}

		// Compute y (evaluation of the errata evaluator polynomial)
		// This is a more faithful translation of the theoretical equation contrary to the old forney method. Here it is an exact reproduction:
		// Yl = omega(Xl.inverse()) / prod(1 - Xj*Xl.inverse()) for j in len(X)
		y := gfPolynomialEval(errorPolynomial, locationInverse) // numerator of the Forney algorithm (errata evaluator evaluated)
		y = gfMultiplication(gfPower(location, 1-fcr), y)       // TODO: adjust to fcr parameter -1 (currently hard coded to 1)

		// Compute the magnitude
		magnitude, _ := gfDivision(y, errorLocatorPrime) // magnitude value of the error, calculated by the Forney algorithm (an equation in fact): dividing the errata evaluator with the errata locator derivative gives us the errata magnitude (ie, value to repair) the ith symbol
		E[errPos[i]] = magnitude                         // store the magnitude for this error into the magnitude polynomial
	}

	return E
}
