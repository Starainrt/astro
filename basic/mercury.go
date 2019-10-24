package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func MercuryL(JD float64) float64 {
	return planet.WherePlanet(1, 0, JD)
}

func MercuryB(JD float64) float64 {
	return planet.WherePlanet(1, 1, JD)
}
func MercuryR(JD float64) float64 {
	return planet.WherePlanet(1, 2, JD)
}
func AMercuryX(JD float64) float64 {
	l := MercuryL(JD)
	b := MercuryB(JD)
	r := MercuryR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AMercuryY(JD float64) float64 {

	l := MercuryL(JD)
	b := MercuryB(JD)
	r := MercuryR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AMercuryZ(JD float64) float64 {
	//l := MercuryL(JD)
	b := MercuryB(JD)
	r := MercuryR(JD)
	//	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AMercuryXYZ(JD float64) (float64, float64, float64) {
	l := MercuryL(JD)
	b := MercuryB(JD)
	r := MercuryR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func MercurySeeRa(JD float64) float64 {
	lo, bo := MercurySeeLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func MercurySeeDec(JD float64) float64 {
	lo, bo := MercurySeeLoBo(JD)
	sita := Sita(JD)
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return dec
}

func MercurySeeRaDec(JD float64) (float64, float64) {
	lo, bo := MercurySeeLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return Limit360(ra), dec
}

func EarthMercuryAway(JD float64) float64 {
	x, y, z := AMercuryXYZ(JD)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func MercurySeeLo(JD float64) float64 {
	x, y, z := AMercuryXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(JD - to)
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

func MercurySeeBo(JD float64) float64 {
	x, y, z := AMercuryXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(JD - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,JD);
	//bo+=GXCBo(lo,bo,JD)/3600;
	//lo+=HJZD(JD);
	return bo
}

func MercurySeeLoBo(JD float64) (float64, float64) {
	x, y, z := AMercuryXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(JD - to)
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

func MercuryMag(JD float64) float64 {
	AwaySun := MercuryR(JD)
	AwayEarth := EarthMercuryAway(JD)
	Away := planet.WherePlanet(-1, 2, JD)
	i := (AwaySun*AwaySun + AwayEarth*AwayEarth - Away*Away) / (2 * AwaySun * AwayEarth)
	i = ArcCos(i)
	Mag := -0.42 + 5*math.Log10(AwaySun*AwayEarth) + 0.0380*i - 0.000273*i*i + 0.000002*i*i*i
	return FloatRound(Mag, 2)
}
