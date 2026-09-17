package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

const (
	// brownLunationEpochJDE 是布朗月序数 0 号朔的儒略日（1923-01-17）。
	brownLunationEpochJDE = 2423436.0
	// brownLunationSynodicMonth 是平均朔望月长度，用于把朔的儒略日换算成月序数。
	brownLunationSynodicMonth = 29.530588853
	// 视半径换算地平视差用的半径比：地球赤道半径除以天体半径。
	horizontalParallaxMoonRatio = 6378.137 / 1737.4
	horizontalParallaxSunRatio  = 6378.137 / 696000.0
)

// SolarEclipseGeocentricPanel 汇总 详细版式面板所需的食甚时刻地心量。
// SolarEclipseGeocentricPanel carries the geocentric quantities that detailed panels print.
type SolarEclipseGeocentricPanel struct {
	// Conjunction 是本次朔，即地心视黄经相等的时刻；与 NASA 全球图上的 Geocentric Conjunction 不是同一个量。
	// Conjunction is the new moon, the instant of equal geocentric apparent ecliptic longitude. It is not
	// the same quantity as the Geocentric Conjunction printed on NASA world maps.
	Conjunction    time.Time
	ConjunctionJDE float64
	// RightAscensionConjunction 是地心视赤经相等的时刻，也就是 NASA 全球图上 Geocentric Conjunction 的口径；
	// 2009-07-22 两者相差约 90 s（视黄经相等在 02:34:34 UT，视赤经相等在 02:33:04 UT）。
	// RightAscensionConjunction is the instant of equal geocentric apparent right ascension, the quantity NASA
	// world maps print as Geocentric Conjunction. For 2009-07-22 the two differ by about 90 s.
	RightAscensionConjunction    time.Time
	RightAscensionConjunctionJDE float64
	// RightAscensionConjunctionJD 是同一时刻的世界时儒略日，NASA 全球图上印的就是它。
	// RightAscensionConjunctionJD is the universal-time Julian day of that instant, the value NASA maps print.
	RightAscensionConjunctionJD float64
	// DeltaTSeconds 是食甚时刻实际使用的 ΔT。
	// DeltaTSeconds is the ΔT used at greatest eclipse.
	DeltaTSeconds float64

	SunRightAscensionDeg  float64
	SunDeclinationDeg     float64
	SunSemidiameterArcsec float64
	SunParallaxArcsec     float64

	MoonRightAscensionDeg  float64
	MoonDeclinationDeg     float64
	MoonSemidiameterArcsec float64
	MoonParallaxArcsec     float64

	LibrationLongitudeDeg     float64
	LibrationLatitudeDeg      float64
	LibrationPositionAngleDeg float64

	BrownLunationNumber int
	// PenumbralK 与 UmbralK 是月地半径比 k1/k2。
	PenumbralK float64
	UmbralK    float64
	// BodyShiftLongitudeArcsec 与 BodyShiftLatitudeArcsec 是星历表里的 Δl/Δb；本库模型不做月面位置平移，恒为零。
	BodyShiftLongitudeArcsec float64
	BodyShiftLatitudeArcsec  float64
	// Ephemeris 是所用模型名称。
	Ephemeris string
	// SingleK 表示使用的是 IAU Single-K（k1 同时用于半影与本影）。
	SingleK bool
}

// horizontalParallaxArcsec 由视半径换算地平视差：sin(HP) = sin(SD) × R⊕ / R天体。
func horizontalParallaxArcsec(semidiameterArcsec, radiusRatio float64) float64 {
	sine := math.Sin(semidiameterArcsec / 3600 * math.Pi / 180)
	return math.Asin(math.Max(-1, math.Min(1, sine*radiusRatio))) * 180 / math.Pi * 3600
}

// SolarEclipseGeocentricPanelAt 计算给定日食在食甚时刻的地心量面板。
// SolarEclipseGeocentricPanelAt computes the geocentric panel of one eclipse at greatest eclipse.
func SolarEclipseGeocentricPanelAt(date time.Time) (SolarEclipseGeocentricPanel, bool) {
	info, ok := SolarEclipseOnDateNASABulletinSplitK(date)
	if !ok {
		return SolarEclipseGeocentricPanel{}, false
	}
	panel := solarEclipseGeocentricPanelAt(info)
	panel.Ephemeris = "NASA bulletin Split-K"
	panel.PenumbralK = basic.SolarEclipsePenumbralK
	panel.UmbralK = basic.SolarEclipseUmbralK
	return panel, true
}

// SolarEclipseGeocentricPanelIAUSingleK 使用 IAU Single-K 计算地心量面板。
// SolarEclipseGeocentricPanelIAUSingleK computes the geocentric panel with the IAU Single-K model.
func SolarEclipseGeocentricPanelIAUSingleK(date time.Time) (SolarEclipseGeocentricPanel, bool) {
	info, ok := SolarEclipseOnDateIAUSingleK(date)
	if !ok {
		return SolarEclipseGeocentricPanel{}, false
	}
	panel := solarEclipseGeocentricPanelAt(info)
	panel.Ephemeris = "IAU Single-K"
	panel.PenumbralK = basic.SolarEclipsePenumbralK
	panel.UmbralK = basic.SolarEclipsePenumbralK
	panel.SingleK = true
	return panel, true
}

func solarEclipseGeocentricPanelAt(info SolarEclipseInfo) SolarEclipseGeocentricPanel {
	tt := solarEclipseTimeToTTJDE(info.GreatestEclipse)
	conjunctionJDE := basic.CalcMoonSHByJDE(tt, 0)
	rightAscensionJDE := solarEclipseRightAscensionConjunction(tt)
	panel := SolarEclipseGeocentricPanel{
		Conjunction:                  solarEclipseTTJDEToTime(conjunctionJDE, info.GreatestEclipse.Location()),
		ConjunctionJDE:               conjunctionJDE,
		RightAscensionConjunction:    solarEclipseTTJDEToTime(rightAscensionJDE, info.GreatestEclipse.Location()),
		RightAscensionConjunctionJDE: rightAscensionJDE,
		RightAscensionConjunctionJD:  rightAscensionJDE - basic.DeltaT(tt, true)/86400,
		DeltaTSeconds:                basic.DeltaT(tt, true),
	}
	panel.SunRightAscensionDeg, panel.SunDeclinationDeg = basic.SunApparentRaDec(tt)
	panel.MoonRightAscensionDeg, panel.MoonDeclinationDeg = basic.HMoonTrueRaDec(tt)
	panel.SunSemidiameterArcsec = basic.SunSemidiameter(tt)
	panel.MoonSemidiameterArcsec = basic.MoonSemidiameter(tt)
	panel.SunParallaxArcsec = horizontalParallaxArcsec(panel.SunSemidiameterArcsec, horizontalParallaxSunRatio)
	panel.MoonParallaxArcsec = horizontalParallaxArcsec(panel.MoonSemidiameterArcsec, horizontalParallaxMoonRatio)
	physical := basic.MoonPhysical(tt)
	panel.LibrationLongitudeDeg = physical.LibrationLongitude
	panel.LibrationLatitudeDeg = physical.LibrationLatitude
	panel.LibrationPositionAngleDeg = physical.PositionAngle
	panel.BrownLunationNumber = int(math.Floor((panel.ConjunctionJDE-brownLunationEpochJDE)/brownLunationSynodicMonth)) + 1
	return panel
}

// solarEclipseRightAscensionGap 返回月亮与太阳的地心视赤经差，单位弧度。
func solarEclipseRightAscensionGap(tt float64) float64 {
	sunRightAscension, _ := basic.SunApparentRaDec(tt)
	moonRightAscension, _ := basic.HMoonTrueRaDec(tt)
	return math.Remainder(moonRightAscension-sunRightAscension, 360) * math.Pi / 180
}

// solarEclipseRightAscensionConjunction 求地心视赤经相等的时刻。
// 赤经差在合附近单调，用牛顿法从食甚时刻收敛即可，不需要预先求朔。
func solarEclipseRightAscensionConjunction(tt float64) float64 {
	estimate := tt
	for iteration := 0; iteration < 40; iteration++ {
		value := solarEclipseRightAscensionGap(estimate)
		if math.Abs(value) < 1e-10 {
			break
		}
		derivative := (solarEclipseRightAscensionGap(estimate+1e-5) - solarEclipseRightAscensionGap(estimate-1e-5)) / 2e-5
		if derivative == 0 || math.IsNaN(derivative) || math.IsInf(derivative, 0) {
			break
		}
		estimate -= value / derivative
	}
	return estimate
}
