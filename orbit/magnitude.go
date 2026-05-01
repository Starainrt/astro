package orbit

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// AsteroidMagnitudeHG 小行星 H-G 模型视星等 / asteroid apparent magnitude using the H-G model.
//
// absoluteMagnitude 为绝对星等 H，slopeParameter 为斜率参数 G。
// absoluteMagnitude is the absolute magnitude H, and slopeParameter is the slope parameter G.
func AsteroidMagnitudeHG(date time.Time, elements Elements, absoluteMagnitude, slopeParameter float64) float64 {
	return basic.OrbitAsteroidMagnitudeHG(ttJulianDay(date), toBasicElements(elements), absoluteMagnitude, slopeParameter)
}
