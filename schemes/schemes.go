package schemes

import "github.com/AlexdeRuijter/stochastic-particle-model/brownian"

type NumericScheme interface {
	Update(float64)
}

type ForwardEuler2D interface {
	NumericScheme
	NewForwardEuler(int64,
		[2]float64,
		func([2]float64) [2]float64,
		func([2]float64) [2]float64,
	) forwardEuler2D
	GetPosition() [2]float64
	GetRandomState() brownian.BrownianState
}
type forwardEuler2D struct {
	w        brownian.BrownianState
	position [2]float64
	f        func([2]float64) [2]float64
	g        func([2]float64) [2]float64
}

func NewForwardEuler2D(seed int64,
	position [2]float64,
	f func([2]float64) [2]float64,
	g func([2]float64) [2]float64) *forwardEuler2D {
	b := brownian.NewBrownianState(seed)
	return &forwardEuler2D{
		w:        b,
		position: position,
		f:        f,
		g:        g,
	}
}

func (fe forwardEuler2D) GetPosition() [2]float64 {
	return fe.position
}

func (fe forwardEuler2D) GetRandomState() brownian.BrownianState {
	return fe.w
}

func (fe *forwardEuler2D) Update(dt float64) {
	// Calculate the random increments in x and y directions
	var dW [2]float64
	for i := 0; i < 2; i++ {
		dW[i] = fe.w.Timestep(dt)
	}

	// Calculate the necessary update directions
	df := fe.f(fe.position)
	dg := fe.g(fe.position)

	// Calculate the positional increments
	for i := 0; i < 2; i++ {
		fe.position[i] += df[i]*dt + dg[i]*dW[i]
	}
}
