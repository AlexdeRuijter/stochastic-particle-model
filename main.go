package main

import (
	"fmt"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/brownian"
)

func main() {
	b := brownian.BrownianState{
		Seed: time.Now().UnixNano(),
		T:    0.0,
		W:    0.0}
	dt := 0.01
	dW := brownian.Timestep(dt, b)
	fmt.Println("Increment is: ", dW, "\nFor timestep: ", dt)
}
