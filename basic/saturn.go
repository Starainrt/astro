package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func SaturnL(JD float64) float64 {
	return planet.WherePlanet(5, 0, JD)
}

func SaturnB(JD float64) float64 {
	return planet.WherePlanet(5, 1, JD)
}
func SaturnR(JD float64) float64 {
	return planet.WherePlanet(5, 2, JD)
}
func ASaturnX(JD float64) float64 {
	l := SaturnL(JD)
	b := SaturnB(JD)
	r := SaturnR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func ASaturnY(JD float64) float64 {

	l := SaturnL(JD)
	b := SaturnB(JD)
	r := SaturnR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func ASaturnZ(JD float64) float64 {
	//l := SaturnL(JD)
	b := SaturnB(JD)
	r := SaturnR(JD)
	//	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func ASaturnXYZ(JD float64) (float64, float64, float64) {
	l := SaturnL(JD)
	b := SaturnB(JD)
	r := SaturnR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func SaturnSeeRa(JD float64) float64 {
	lo, bo := SaturnSeeLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func SaturnSeeDec(JD float64) float64 {
	lo, bo := SaturnSeeLoBo(JD)
	sita := Sita(JD)
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return dec
}

func SaturnSeeRaDec(JD float64) (float64, float64) {
	lo, bo := SaturnSeeLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return Limit360(ra), dec
}

func EarthSaturnAway(JD float64) float64 {
	x, y, z := ASaturnXYZ(JD)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func SaturnSeeLo(JD float64) float64 {
	x, y, z := ASaturnXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(JD - to)
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

func SaturnSeeBo(JD float64) float64 {
	x, y, z := ASaturnXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(JD - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,JD);
	//bo+=GXCBo(lo,bo,JD)/3600;
	//lo+=HJZD(JD);
	return bo
}

func SaturnSeeLoBo(JD float64) (float64, float64) {
	x, y, z := ASaturnXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = ASaturnXYZ(JD - to)
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

func SaturnMag(JD float64) float64 {
	AwaySun := SaturnR(JD)
	AwayEarth := EarthSaturnAway(JD)
	Away := planet.WherePlanet(-1, 2, JD)
	i := (AwaySun*AwaySun + AwayEarth*AwayEarth - Away*Away) / (2 * AwaySun * AwayEarth)
	i = ArcCos(i)
	Mag := -8.68 + 5*math.Log10(AwaySun*AwayEarth) + 0.044*i - 2.6*Sin(math.Abs(SaturnRingB(JD))) + 1.25*Sin(math.Abs(SaturnRingB(JD)))*Sin(math.Abs(SaturnRingB(JD)))
	return FloatRound(Mag, 2)
}

func SaturnRingB(JD float64) float64 {
	T := (JD - 2451545) / 36525
	i := 28.075216 - 0.012998*T + 0.000004*T*T
	omi := 169.508470 + 1.394681*T + 0.000412*T*T
	lo, bo := SaturnSeeLoBo(JD)
	B := Sin(i)*Cos(bo)*Sin(lo-omi) - Cos(i)*Cos(bo)
	return ArcSin(B)
}
