package matrixplots

import (
	"fmt"

	"github.com/pkg/errors"
)

func Histogram2D(x []float64, y []float64, nBins int, xmin float64, xmax float64, ymin float64, ymax float64) [][]int64 {
	if len(x) != len(y) {
		panic(errors.New("Histogram2D: different number of elements for x and y"))
	}

	xdif := (xmax - xmin) / float64(nBins)
	ydif := (ymax - ymin) / float64(nBins)

	// Create the matrix and initialize it with zeros
	mat := make([][]int64, nBins)
	for i := 0; i < nBins; i++ {
		mat[i] = make([]int64, nBins)
	}

	for k, xk := range x {
		yk := y[k]

		i := int((xk - xmin) / xdif)
		j := int((yk - ymin) / ydif)

		if i < 0 || i > nBins-1 || j < 0 || j > nBins-1 {
			fmt.Println("Following datapoint was outside of limits: ")
			fmt.Println("x: ", xk, ", y: ", yk)
		} else {
			mat[i][j] += 1
		}
	}

	return mat
}
