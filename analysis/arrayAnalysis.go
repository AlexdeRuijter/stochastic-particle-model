package analysis

import "math"

// Calculate the mean of a slice of float64's
func CalculateMean(list []float64) float64 {
	var mean float64

	total := len(list)

	for _, v := range list {
		mean += v
	}
	return mean / float64(total)
}

// Calculate the variation of a slice of float64's
func CalculateVariation(list []float64, mean float64) float64 {
	var variation float64

	total := len(list)

	for _, v := range list {
		variation = (v - mean) * (v - mean)
	}

	variation = math.Sqrt(variation) / float64(total)

	return variation
}

// Calculate the mean of a chan of float64's with an unknown total of elements
func CalculateMeanChannel(c chan float64) float64 {
	var mean float64
	var total int

	for v := range c {
		total += 1
		mean = mean * float64(total-1)
		mean += v
		mean /= float64(total)
	}

	return mean
}

// Calculate the mean and variation of a list-slice of float64s
func CalculateMeanAndVariation(list []float64) [2]float64 {
	var result [2]float64 // Create a placeholder variable

	result[0] = CalculateMean(list)                 // Calculate and store the mean
	result[1] = CalculateVariation(list, result[0]) // Calculate and store the variation

	return result
}
