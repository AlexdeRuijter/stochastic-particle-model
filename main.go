package main

import (
	"fmt"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/analysis"
	"github.com/AlexdeRuijter/stochastic-particle-model/schemes"
)

func f(position [2]float64) [2]float64 {
	return position
}

func g(position [2]float64) [2]float64 {
	return position
}

func test(c chan [2]float64) {
	var initial_position [2]float64
	initial_position = [2]float64{.5, .5}
	scheme := schemes.NewForwardEuler2D(time.Now().UnixNano(), initial_position, f, g)
	for i := 0; i < 1000; i++ {
		scheme.Update(0.001)
	}
	initial_position = scheme.GetPosition()
	c <- initial_position
}

func main() {
	a := [4]float64{0, 10, 3}
	mean := analysis.CalculateMean(a[:])
	variation := analysis.CalculateVariation(a[:], mean)

	fmt.Println("Mean:", mean)
	fmt.Println("Variation:", variation)
}
