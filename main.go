package main

import (
	"fmt"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/brownian"
)

func main() {
	b := brownian.NewBrownianState(time.Now().UnixNano())
	dt := 0.01
	dW := b.Timestep(dt)
	fmt.Println("Increment is: ", dW, "\nFor timestep: ", dt)
}
