package schemes

import "github.com/AlexdeRuijter/stochastic-particle-model/brownian"

type NumericScheme interface {
	Update(float64) float64
}

type ForwardEuler interface {
	Update(float64) float64
	NewForwardEuler(int64,
		[2]float64,
		func([2]float64) float64,
		func([2]float64) float64,
	) forwardEuler
	GetPosition() [2]float64
	GetRandomState() brownian.BrownianState
}
type forwardEuler struct {
	w        brownian.BrownianState
	position [2]float64
	f        func([2]float64) float64
	g        func([2]float64) float64
}

func NewForwardEuler(seed int64,
	position [2]float64,
	f func([2]float64) float64,
	g func([2]float64) float64) *forwardEuler {
	b := brownian.NewBrownianState(seed)
	return &forwardEuler{
		w:        b,
		position: position,
		f:        f,
		g:        g,
	}
}
