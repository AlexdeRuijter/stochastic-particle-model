package main

import (
	"fmt"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/schemes"
)

func f(position [2]float64) [2]float64 {
	return position
}

func g(position [2]float64) [2]float64 {
	return position
}

func main() {
	var initial_position [2]float64
	initial_position = [2]float64{.5, .5}
	scheme := schemes.NewForwardEuler2D(time.Now().UnixNano(), initial_position, f, g)
	scheme.Update(0.001)
	initial_position = scheme.GetPosition()
	fmt.Println(initial_position)
}
