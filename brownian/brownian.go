package brownian

import (
	"math"
	"math/rand"
)

type BrownianState struct {
	Seed int64
	T    float64
	W    float64
}

func Timestep(dt float64, state BrownianState) float64 {
	rand.Seed(state.Seed)
	dW := rand.NormFloat64() * math.Sqrt(dt)
	state.T += dt
	state.W += dW

	return dW
}
