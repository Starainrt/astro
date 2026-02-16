// Package astro
package astro

import "github.com/starainrt/astro/basic"

func DeltaT() func(float64, bool) float64 {
	return basic.GetDeltaTFn()
}

func SetDeltaT(deltaT func(float64, bool) float64) {
	basic.SetDeltaTFn(deltaT)
}

func DefaultDeltaT() func(float64, bool) float64 {
	return basic.DefaultDeltaTv2
}
