package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func UranusL(JD float64) float64 {
	return planet.WherePlanet(6, 0, JD)
}

func UranusB(JD float64) float64 {
	return planet.WherePlanet(6, 1, JD)
}
func UranusR(JD float64) float64 {
	return planet.WherePlanet(6, 2, JD)
}
func AUranusX(JD float64) float64 {
	l := UranusL(JD)
	b := UranusB(JD)
	r := UranusR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AUranusY(JD float64) float64 {

	l := UranusL(JD)
	b := UranusB(JD)
	r := UranusR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AUranusZ(JD float64) float64 {
	//l := UranusL(JD)
	b := UranusB(JD)
	r := UranusR(JD)
	//	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AUranusXYZ(JD float64) (float64, float64, float64) {
	l := UranusL(JD)
	b := UranusB(JD)
	r := UranusR(JD)
	el := planet.WherePlanet(-1, 0, JD)
	eb := planet.WherePlanet(-1, 1, JD)
	er := planet.WherePlanet(-1, 2, JD)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func UranusSeeRa(JD float64) float64 {
	lo, bo := UranusSeeLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func UranusSeeDec(JD float64) float64 {
	lo, bo := UranusSeeLoBo(JD)
	sita := Sita(JD)
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return dec
}

func UranusSeeRaDec(JD float64) (float64, float64) {
	lo, bo := UranusSeeLoBo(JD)
	sita := Sita(JD)
	ra := math.Atan2((Sin(lo)*Cos(sita) - Tan(bo)*Sin(sita)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(sita) + Cos(bo)*Sin(sita)*Sin(lo))
	return Limit360(ra), dec
}

func EarthUranusAway(JD float64) float64 {
	x, y, z := AUranusXYZ(JD)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func UranusSeeLo(JD float64) float64 {
	x, y, z := AUranusXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AUranusXYZ(JD - to)
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

func UranusSeeBo(JD float64) float64 {
	x, y, z := AUranusXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AUranusXYZ(JD - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,JD);
	//bo+=GXCBo(lo,bo,JD)/3600;
	//lo+=HJZD(JD);
	return bo
}

func UranusSeeLoBo(JD float64) (float64, float64) {
	x, y, z := AUranusXYZ(JD)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AUranusXYZ(JD - to)
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

func UranusMag(JD float64) float64 {
	AwaySun := UranusR(JD)
	AwayEarth := EarthUranusAway(JD)
	Away := planet.WherePlanet(-1, 2, JD)
	i := (AwaySun*AwaySun + AwayEarth*AwayEarth - Away*Away) / (2 * AwaySun * AwayEarth)
	i = ArcCos(i)
	Mag := -7.19 + 5*math.Log10(AwaySun*AwayEarth) + 0.016*i
	return FloatRound(Mag, 2)
}
