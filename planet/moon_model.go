package planet

import "github.com/starainrt/astro/internal/ephem"

// WhereMoonModelN 按指定模型计算月球几何量 / evaluate lunar geometry with an explicit model.
func WhereMoonModelN(model ephem.Model, coordIndex int, jd float64, n int) float64 {
	if model == ephem.Semianalytic && coordIndex >= 0 && coordIndex <= 2 {
		return semiMoon(coordIndex, jd, n)
	}
	return elpMoonN(coordIndex, jd, n)
}
