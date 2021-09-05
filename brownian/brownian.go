package brownian

import (
	"math"
	"math/rand"
)

type BrownianState interface {
	Timestep(float64) float64
	NewBrownianState() brownianState
	GetTime() float64
	GetWtotal() float64
}
type brownianState struct {
	t float64
	w float64
}

func NewBrownianState(seed int64) brownianState {
	rand.Seed(seed)
	return brownianState{t: 0.0, w: 0.0}
}

func (p *brownianState) GetTime() float64 {
	return p.t
}

func (p *brownianState) GetWtotal() float64 {
	return p.w
}

func (p *brownianState) Timestep(dt float64) float64 {
	dW := rand.NormFloat64() * math.Sqrt(dt)
	p.t += dt
	p.w += dW

	return dW
}
