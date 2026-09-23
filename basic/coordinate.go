package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

/*
 * 坐标变换，黄道转赤道
 */
func LoToRa(jde, lo, bo float64) float64 {
	eps := TrueObliquity(jde)
	ra := math.Atan2(Sin(lo)*Cos(eps)-Tan(bo)*Sin(eps), Cos(lo))
	ra = ra * 180 / math.Pi
	if ra < 0 {
		ra += 360
	}
	return ra
}

func BoToDec(jde, lo, bo float64) float64 {
	eps := TrueObliquity(jde)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func LoBoToRaDec(jde, lo, bo float64) (float64, float64) {
	eps := TrueObliquity(jde)
	sinEps := Sin(eps)
	cosEps := Cos(eps)
	dec := ArcSin(Sin(bo)*cosEps + Cos(bo)*sinEps*Sin(lo))
	ra := math.Atan2(Sin(lo)*cosEps-Tan(bo)*sinEps, Cos(lo))
	ra = ra * 180 / math.Pi
	if ra < 0 {
		ra += 360
	}
	return ra, dec
}

func RaDecToLoBo(jde, ra, dec float64) (float64, float64) {
	//tan(λ) = (sin(α)*cos(ε) + tan(δ)*sin(ε)) / cos(α)
	//sin(β)=sin(δ)*cos(ε)-cos(δ)*sin(ε)*sin(α)
	eps := TrueObliquity(jde)
	sinBo := Sin(dec)*Cos(eps) - Cos(dec)*Sin(eps)*Sin(ra)
	lo := math.Atan2((Sin(ra)*Cos(eps) + Tan(dec)*Sin(eps)), Cos(ra))
	lo = Limit360(lo * 180 / math.Pi)
	return lo, ArcSin(sinBo)
}

func RaToLo(jde, ra, dec float64) float64 {
	//tan(λ) = (sin(α)*cos(ε) + tan(δ)*sin(ε)) / cos(α)
	//sin(β)=sin(δ)*cos(ε)-cos(δ)*sin(ε)*sin(α)
	eps := TrueObliquity(jde)
	lo := math.Atan2((Sin(ra)*Cos(eps) + Tan(dec)*Sin(eps)), Cos(ra))
	lo = Limit360(lo * 180 / math.Pi)
	return lo
}

func DecToBo(jde, ra, dec float64) float64 {
	//tan(λ) = (sin(α)*cos(ε) + tan(δ)*sin(ε)) / cos(α)
	//sin(β)=sin(δ)*cos(ε)-cos(δ)*sin(ε)*sin(α)
	eps := TrueObliquity(jde)
	sinBo := Sin(dec)*Cos(eps) - Cos(dec)*Sin(eps)*Sin(ra)
	return ArcSin(sinBo)
}

/*
 * 地心坐标转站心坐标，参数分别为，地心赤经赤纬 纬度经度，jde，离地心位置au
 */
func pcosi(lat, h float64) float64 {
	b := 6356.755
	a := 6378.14
	u := ArcTan(b / a * Tan(lat))
	//psin=b/a*Sin(u)+h/6378140*Sin(lat);
	pcos := Cos(u) + h/6378140.0*Cos(lat)
	return pcos
}
func psini(lat, h float64) float64 {
	b := 6356.755
	a := 6378.14
	u := ArcTan(b / a * Tan(lat))
	psin := b/a*Sin(u) + h/6378140*Sin(lat)
	//pcos=Cos(u)+h/6378140*Cos(lat);
	return psin
}

func TopocentricRaDec(ra, dec, lat, lon, jd, au, h float64) (float64, float64) {
	return topocentricRaDecWithSidereal(ra, dec, lat, lon, ApparentSiderealTime(UTC2UT1(jd))*15, au, h)
}

func topocentricRaDecWithSidereal(ra, dec, lat, lon, siderealDegrees, au, h float64) (float64, float64) {
	sinpi := Sin(0.0024427777777) / au
	pcosi := pcosi(lat, h)
	psini := psini(lat, h)
	tH := Limit360(siderealDegrees + lon - ra)
	nra := math.Atan2(-pcosi*sinpi*Sin(tH), (Cos(dec)-pcosi*sinpi*Cos(tH))) * 180 / math.Pi

	ndec := math.Atan2((Sin(dec)-psini*sinpi)*Cos(nra), (Cos(dec)-pcosi*sinpi*Cos(tH))) * 180 / math.Pi
	return ra + nra, ndec
}

func TopocentricRa(ra, dec, lat, lon, jd, au, h float64) float64 { //jd为格林尼治标准时
	topocentricRA, _ := TopocentricRaDec(ra, dec, lat, lon, jd, au, h)
	return topocentricRA
}
func TopocentricDec(ra, dec, lat, lon, jd, au, h float64) float64 { //jd为格林尼治标准时
	_, topocentricDec := TopocentricRaDec(ra, dec, lat, lon, jd, au, h)
	return topocentricDec
}

// TopocentricLoBo 一次求值给出站心黄经与黄纬 / topocentric ecliptic longitude and latitude in one solve.
//
// 先解站心赤道坐标再转黄道：黄道版公式的分母在黄经 90°–270° 时变号，直接 atan2 会切到对顶象限。
func TopocentricLoBo(lo, bo, lat, lon, jde, au, h float64) (float64, float64) {
	ra, dec := LoBoToRaDec(jde, lo, bo)
	topRA, topDec := TopocentricRaDec(ra, dec, lat, lon, jde, au, h)
	return RaDecToLoBo(jde, topRA, topDec)
}

// TopocentricLo 站心黄经 / topocentric ecliptic longitude.
func TopocentricLo(lo, bo, lat, lon, jde, au, h float64) float64 {
	nlo, _ := TopocentricLoBo(lo, bo, lat, lon, jde, au, h)
	return nlo
}

// TopocentricBo 站心黄纬 / topocentric ecliptic latitude.
func TopocentricBo(lo, bo, lat, lon, jde, au, h float64) float64 {
	_, nbo := TopocentricLoBo(lo, bo, lat, lon, jde, au, h)
	return nbo
}

func GXCLo(lo, bo, jde float64) float64 { //光行差修正
	k := 20.49552
	sunlo := SunTrueLo(jde)
	e := Earthe(jde)
	epi := EarthPI(jde)
	tmp := (-k*Cos(sunlo-lo) + e*k*Cos(epi-lo)) / Cos(bo)
	return tmp
}

func GXCBo(lo, bo, jde float64) float64 {
	k := 20.49552
	sunlo := SunTrueLo(jde)
	e := Earthe(jde)
	epi := EarthPI(jde)
	tmp := -k * Sin(bo) * (Sin(sunlo-lo) - e*Sin(epi-lo))
	return tmp
}
