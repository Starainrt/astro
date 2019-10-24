package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func NeptuneL(JD float64) float64 {
	return planet.WherePlanet(7, 0, JD)
}

func NeptuneB(JD float64) float64 {
	return planet.WherePlanet(7, 1, JD)
}
func NeptuneR(JD float64) float64 {
	return planet.WherePlanet(7, 2, JD)
}
func ANeptuneX(JD float64) float64 {
	l := NeptuneL(JD)
	b := NeptuneB(JD)
	r := NeptuneR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func ANeptuneY(JD float64) float64 {

	l := NeptuneL(JD)
	b := NeptuneB(JD)
	r := NeptuneR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func ANeptuneZ(JD float64) float64 {
	//l := NeptuneL(JD)
	b := NeptuneB(JD)
	r := NeptuneR(JD)
	//	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func ANeptuneXYZ(JD float64) (float64, float64, float64) {
	l := NeptuneL(JD)
	b := NeptuneB(JD)
	r := NeptuneR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func NeptuneSeeRa(JD float64) float64 {
	lo, bo := NeptuneSeeLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func NeptuneSeeDec(JD float64) float64 {
	lo, bo := NeptuneSeeLoBo(JD)
	sita := Sita(JD)
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return dec
}

func NeptuneSeeRaDec(JD float64) (float64, float64) {
	lo, bo := NeptuneSeeLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return Limit360(ra), dec
}

func EarthNeptuneAway(JD float64) float64 {
	x, y, z := ANeptuneXYZ(JD)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func NeptuneSeeLo(JD float64) float64 {
	x, y, z := ANeptuneXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ANeptuneXYZ(JD - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo-=GXCLo(lo,bo,JD)/3600;
	//bo+=GXCBo(lo,bo,JD);
	lo += HJZD(JD)
	return lo
}

func NeptuneSeeBo(JD float64) float64 {
	x, y, z := ANeptuneXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ANeptuneXYZ(JD - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,JD);
	//bo+=GXCBo(lo,bo,JD)/3600;
	//lo+=HJZD(JD);
	return bo
}

func NeptuneSeeLoBo(JD float64) (float64, float64) {
	x, y, z := ANeptuneXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ANeptuneXYZ(JD - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo-=GXCLo(lo,bo,JD)/3600;
	//bo+=GXCBo(lo,bo,JD);
	lo += HJZD(JD)
	return lo, bo
}

func NeptuneMag(JD float64) float64 {
	AwaySun := NeptuneR(JD)
	AwayEarth := EarthNeptuneAway(JD)
	Away := planet.WherePlanet(-1, 2, JD)
	i := (AwaySun*AwaySun + AwayEarth*AwayEarth - Away*Away) / (2 * AwaySun * AwayEarth)
	i = ArcCos(i)
	Mag := -6.87 + 5*math.Log10(AwaySun*AwayEarth)
	return FloatRound(Mag, 2)
}
