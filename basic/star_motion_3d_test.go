package basic

import (
	"math"
	"testing"
	"time"
)

// 本文件锁定恒星自行的口径：无距离时二维、有距离时三维、两者只差二阶项。

func starSepArcsec(ra1, dec1, ra2, dec2 float64) float64 {
	ra1Rad := ra1 * math.Pi / 180
	dec1Rad := dec1 * math.Pi / 180
	ra2Rad := ra2 * math.Pi / 180
	dec2Rad := dec2 * math.Pi / 180
	cosine := math.Sin(dec1Rad)*math.Sin(dec2Rad) + math.Cos(dec1Rad)*math.Cos(dec2Rad)*math.Cos(ra1Rad-ra2Rad)
	return math.Acos(math.Max(-1, math.Min(1, cosine))) * 180 / math.Pi * 3600
}

func TestRaDecByJdeFallsBackToTwoDimensionsWithoutDistance(t *testing.T) {
	stars := []InnerStarData{
		{Ra: 10, Dec: 60, PmRA: 3600, PmDec: -1800, RadVel: 50},
		{Ra: 359.5, Dec: -89.99, PmRA: -1200, PmDec: 900, RadVel: -80},
		{Ra: 0, Dec: 90, PmRA: 1000, PmDec: 1000},
	}
	for _, star := range stars {
		for _, jde := range []float64{2451545.0, 2451545.0 + 26*365.25, 2451545.0 - 100*365.25} {
			gotRA, gotDec := star.RaDecByJde(jde)
			wantRA, wantDec := star.raDecByJde2D(jde)
			if gotRA != wantRA || gotDec != wantDec {
				t.Fatalf("no-distance path = %.12f %.12f, want bit-identical to 2D %.12f %.12f", gotRA, gotDec, wantRA, wantDec)
			}
		}
	}
}

func TestRaDecByJdeUsesJulianYear(t *testing.T) {
	// 高自行近星才让儒略年与回归年的年长差在三坐标上可测。
	star := InnerStarData{Ra: 10, Dec: 0, PmRA: 3600, Pc: 1, RadVel: -100}
	jde := 2451545.0 + 365.25
	gotRA, gotDec := star.RaDecByJde(jde)
	wantRA, wantDec := star.raDecByJde3D(jde)
	if gotRA != wantRA || gotDec != wantDec {
		t.Fatalf("RaDecByJde = %.12f %.12f, want 3D path %.12f %.12f", gotRA, gotDec, wantRA, wantDec)
	}
	tropicalRA, tropicalDec := star.raDecByJde3D(2451545.0 + 365.2422)
	// 两种年长每年差 11.23 分钟，乘以 2 度/年的自行量级实测约 0.078 角秒。
	if separation := starSepArcsec(gotRA, gotDec, tropicalRA, tropicalDec); separation < 0.05 {
		t.Fatalf("Julian and tropical year lengths differ by only %.6f arcsec, want a distinguishable epoch step", separation)
	}
}

func TestRaDecByJdeThreeDimensionsMatchesIntegration(t *testing.T) {
	stars := []InnerStarData{
		{Ra: 224.366667, Dec: -21.415556, PmRA: 1.045, PmDec: -1.729, Pc: 5.7803, RadVel: 20},
		{Ra: 101.287083, Dec: -16.716111, PmRA: -0.553, PmDec: -1.205, Pc: 2.6667, RadVel: -8},
		{Ra: 44.565278, Dec: 23.6, PmRA: 2.01, PmDec: -1.9, Pc: 1.83, RadVel: -110.6},
	}
	for _, star := range stars {
		for _, years := range []float64{26, 100, 1000} {
			jde := 2451545.0 + years*365.25
			gotRA, gotDec := star.RaDecByJde(jde)
			// 真值只替代自行推进这一段，岁差链必须与主路径一致。
			wantRA, wantDec := integrateInnerStarMotion(star, years)
			wantRA, wantDec = Precess(wantRA, wantDec, 2451545.0, jde)
			if separation := starSepArcsec(gotRA, gotDec, wantRA, wantDec); separation > 0.01 {
				t.Fatalf("motion at %.0f years = %.12f %.12f, want %.12f %.12f (%.6f arcsec apart)",
					years, gotRA, gotDec, wantRA, wantDec, separation)
			}
		}
	}
}

// integrateInnerStarMotion 用极小步长数值积分三维匀速直线运动，作为解析式的独立真值。
func integrateInnerStarMotion(star InnerStarData, years float64) (float64, float64) {
	const steps = 20000
	raRad := star.Ra * math.Pi / 180
	decRad := star.Dec * math.Pi / 180
	cosDec, sinDec := math.Cos(decRad), math.Sin(decRad)
	distance := star.distanceAU()
	pmRA := star.PmRA * math.Pi / (180 * 3600) * distance
	pmDec := star.PmDec * math.Pi / (180 * 3600) * distance
	radial := star.RadVel * 365.25 * 86400 / 149597870.7
	velocity := [3]float64{
		pmDec*(-sinDec*math.Cos(raRad)) - pmRA*math.Sin(raRad) + radial*cosDec*math.Cos(raRad),
		pmDec*(-sinDec*math.Sin(raRad)) + pmRA*math.Cos(raRad) + radial*cosDec*math.Sin(raRad),
		pmDec*cosDec + radial*sinDec,
	}
	position := [3]float64{distance * cosDec * math.Cos(raRad), distance * cosDec * math.Sin(raRad), distance * sinDec}
	step := years / steps
	for i := 0; i < steps; i++ {
		for axis := range position {
			position[axis] += step * velocity[axis]
		}
	}
	norm := math.Sqrt(position[0]*position[0] + position[1]*position[1] + position[2]*position[2])
	ra := math.Atan2(position[1], position[0]) * 180 / math.Pi
	if ra < 0 {
		ra += 360
	}
	return ra, math.Asin(position[2]/norm) * 180 / math.Pi
}

func TestRaDecByDateUsesTTEpoch(t *testing.T) {
	star := StarData{InnerStarData: InnerStarData{Ra: 120, Dec: 20, PmRA: 3600, PmDec: 3600, Pc: 0.5, RadVel: -60}}
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	gotRA, gotDec := star.RaDecByDate(date)
	wantRA, wantDec := star.InnerStarData.RaDecByJde(UTC2TT(Date2JD(date.UTC())))
	if gotRA != wantRA || gotDec != wantDec {
		t.Fatalf("RaDecByDate = %.12f %.12f, want TT propagation %.12f %.12f", gotRA, gotDec, wantRA, wantDec)
	}
	// 一千年跨度下径向项把 TT 与民用时的历元差放大到可测，务必不要退回 UTC。
	longDate := time.Date(3026, 1, 1, 0, 0, 0, 0, time.UTC)
	longTTRA, longTTDec := star.RaDecByDate(longDate)
	longUTCRA, longUTCDec := star.InnerStarData.RaDecByJde(Date2JD(longDate.UTC()))
	// 3026 年时 TT-UTC 为 1.28 小时，实测该口径差约 0.0043 角秒。
	if separation := starSepArcsec(longTTRA, longTTDec, longUTCRA, longUTCDec); separation < 0.002 {
		t.Fatalf("TT and UTC epochs differ by only %.9f arcsec at 1000 years, want a distinguishable TT path", separation)
	}
}

func TestRaDecByJdeDistanceGateUsesCatalogDistance(t *testing.T) {
	star := InnerStarData{Ra: 224.366667, Dec: -21.415556, PmRA: 1.045, PmDec: -1.729, Pc: 5.7803, RadVel: -100}
	jde := 2451545.0 + 100*365.25
	if star.distanceAU() == 0 {
		t.Fatal("catalog distance should be positive for a star with Pc > 0")
	}
	withDistanceRA, withDistanceDec := star.RaDecByJde(jde)
	star.Pc = 0
	withoutDistanceRA, withoutDistanceDec := star.RaDecByJde(jde)
	separation := starSepArcsec(withDistanceRA, withDistanceDec, withoutDistanceRA, withoutDistanceDec)
	if separation < 0.1 || separation > 1.5 {
		t.Fatalf("radial term should separate the two branches by a few tenths of an arcsec at 100 years, got %.6f", separation)
	}
}
