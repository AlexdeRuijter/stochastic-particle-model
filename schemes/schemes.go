package schemes

import "github.com/AlexdeRuijter/stochastic-particle-model/brownian"

type NumericSchemeer interface {
	Update(float64) float64
}

type ForwardEuler struct {
	W brownian.BrownianState
}
