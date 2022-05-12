package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

/*
 * 坐标变换，黄道转赤道
 */
func LoToRa(lo, bo, jde float64) float64 {
	ra := math.Atan2(Sin(lo)*Cos(Sita(jde)-Tan(bo)*Sin(Sita(jde))), Cos(lo))
	ra = ra * 180 / math.Pi
	if ra < 0 {
		ra += 360
	}
	return ra
}

func BoToDec(lo, bo, jde float64) float64 {
	dec := ArcSin(Sin(bo)*Cos(Sita(jde)) + Cos(bo)*Sin(Sita(jde))*Sin(lo))
	return dec
}

/*
 * 赤道坐标岁差变换st end 为JDE时刻
 */
func ZuoBiaoSuiCha(ra, dec, st, end float64) (float64, float64) {
	t := (end - st) / 36525.0
	l := (2306.2181*t + 0.30188*t*t + 0.017998*t*t*t) / 3600
	z := (2306.2181*t + 1.09468*t*t + 0.018203*t*t*t) / 3600
	o := (2004.3109*t - 0.42665*t*t + 0.041833*t*t*t) / 3600
	A := Cos(dec) * Sin(ra+l)
	B := Cos(o)*Cos(dec)*Cos(ra+l) - Sin(o)*Sin(dec)
	C := Sin(o)*Cos(dec)*Cos(ra+l) + Cos(o)*Sin(dec)
	ras := math.Atan2(A, B)
	ras = ras * 180 / math.Pi
	if ras < 0 {
		ras += 360
	}
	ra = ras + z
	dec = ArcSin(C)
	return ra, dec
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

func ZhanXinRaDec(ra, dec, lat, lon, jd, au, h float64) (float64, float64) {
	sinpi := Sin(0.0024427777777) / au
	pcosi := pcosi(lat, h)
	psini := psini(lat, h)
	tH := Limit360(TD2UT(ApparentSiderealTime(jd), false)*15 + lon - ra)
	nra := math.Atan2(-pcosi*sinpi*Sin(tH), (Cos(dec)-pcosi*sinpi*Cos(tH))) * 180 / math.Pi

	ndec := math.Atan2((Sin(dec)-psini*sinpi)*Cos(nra), (Cos(dec)-pcosi*sinpi*Cos(tH))) * 180 / math.Pi
	return ra + nra, ndec
}

func ZhanXinRa(ra, dec, lat, lon, jd, au, h float64) float64 { //jd为格林尼治标准时
	sinpi := Sin(0.0024427777777) / au
	pcosi := pcosi(lat, h)
	tH := Limit360(TD2UT(ApparentSiderealTime(jd), false)*15 + lon - ra)
	nra := math.Atan2(-pcosi*sinpi*Sin(tH), (Cos(dec)-pcosi*sinpi*Cos(tH))) * 180 / math.Pi
	return ra + nra
}
func ZhanXinDec(ra, dec, lat, lon, jd, au, h float64) float64 { //jd为格林尼治标准时

	sinpi := Sin(0.0024427777777) / au
	pcosi := pcosi(lat, h)
	psini := psini(lat, h)
	tH := Limit360(TD2UT(ApparentSiderealTime(jd), false)*15 + lon - ra)
	nra := math.Atan2(-pcosi*sinpi*Sin(tH), (Cos(dec)-pcosi*sinpi*Cos(tH))) * 180 / math.Pi

	ndec := math.Atan2((Sin(dec)-psini*sinpi)*Cos(nra), (Cos(dec)-pcosi*sinpi*Cos(tH))) * 180 / math.Pi
	return ndec
}

func ZhanXinLo(lo, bo, lat, lon, jd, au, h float64) float64 { //jd为格林尼治标准时
	C := pcosi(lat, h)
	S := psini(lat, h)
	sinpi := Sin(0.0024427777777) / au
	ra := LoToRa(lo, bo, jd)
	tH := Limit360(TD2UT(ApparentSiderealTime(jd), false)*15 + lon - ra)
	N := Cos(lo)*Cos(bo) - C*sinpi*Cos(tH)
	nlo := math.Atan2(Sin(lo)*Cos(bo)-sinpi*(S*Sin(Sita(jd))+C*Cos(Sita(jd))*Sin(tH)), N) * 180 / math.Pi
	return nlo
}

func ZhanXinBo(lo, bo, lat, lon, jd, au, h float64) float64 { //jd为格林尼治标准时
	C := pcosi(lat, h)
	S := psini(lat, h)
	sinpi := Sin(0.0024427777777) / au
	ra := LoToRa(lo, bo, jd)
	tH := Limit360(TD2UT(ApparentSiderealTime(jd), false)*15 + lon - ra)
	N := Cos(lo)*Cos(bo) - C*sinpi*Cos(tH)
	nlo := math.Atan2(Sin(lo)*Cos(bo)-sinpi*(S*Sin(Sita(jd))+C*Cos(Sita(jd))*Sin(tH)), N) * 180 / math.Pi
	nbo := math.Atan2(Cos(nlo)*(Sin(bo)-sinpi*(S*Cos(Sita(jd))-C*Sin(Sita(jd))*Sin(tH))), N) * 180 / math.Pi
	return nbo
}

/*
 func GXCLo(lo,bo,jd float64) float64{  //光行差修正
k:=20.49552;
sunlo:=SunTrueLo(jd);
e:=Earthe(jd);
epi=earth->EarthPI(jd);
tmp=(-k*this->CosR(sunlo-lo)+e*k*this->CosR(epi-lo))/this->CosR(bo);
return tmp;
}

public function GXCBo(lo,bo,jd)
{
	earth=new Earth();
k=20.49552;
sunlo=earth->SunTrueLo(jd);
e=earth->Earthe(jd);
epi=earth->EarthPI(jd);
tmp=-k*this->SinR(bo)*(this->SinR(sunlo-lo)-e*this->SinR(epi-lo));
return tmp;
}

*/
