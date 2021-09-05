package brownian

import (
	"math"
	"math/rand"
)

type BrownianState interface {
	Timestep(float64) float64
	GetTime() float64
	GetWtotal() float64
}
type brownianState struct {
	t float64
	w float64
}

func NewBrownianState(seed int64) *brownianState {
	rand.Seed(seed)
	return &brownianState{t: 0.0, w: 0.0}
}

func (b *brownianState) GetTime() float64 {
	return b.t
}

func (b *brownianState) GetWtotal() float64 {
	return b.w
}

func (b *brownianState) setTime(t float64) {
	b.t = t
}

func (b *brownianState) setW(W float64) {
	b.w = W
}

func (b *brownianState) Timestep(dt float64) float64 {
	dW := rand.NormFloat64() * math.Sqrt(dt)
	b.setTime(b.t + dt)
	b.setW(b.w + dW)

	return dW
}
