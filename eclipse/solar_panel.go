package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

const (
	// brownLunationEpochJDE 是布朗月序数 0 号朔的儒略日（1923-01-17）。
	brownLunationEpochJDE = 2423436.0
	// brownLunationSynodicMonth 是平均朔望月长度，用于把朔的儒略日换算成月序数。
	brownLunationSynodicMonth = 29.530588853
	// 视半径换算地平视差用的半径比：地球赤道半径除以天体半径。
	horizontalParallaxMoonRatio = geodata.EarthWGS84RadiusKM / 1737.4
	horizontalParallaxSunRatio  = geodata.EarthWGS84RadiusKM / 696000.0
)

// SolarEclipseGeocentricPanel 详细面板所需的食甚时刻地心量 / geocentric quantities at greatest eclipse for detailed panels.
type SolarEclipseGeocentricPanel struct {
	// Conjunction 地心视黄经相等的朔时刻 / new moon at equal geocentric apparent ecliptic longitude.
	Conjunction    time.Time
	ConjunctionJDE float64
	// RightAscensionConjunction 地心视赤经相等的时刻，与朔的时差不固定 / equal geocentric apparent right ascension, with a variable gap from new moon.
	RightAscensionConjunction    time.Time
	RightAscensionConjunctionJDE float64
	// RightAscensionConjunctionJD 同一合时刻的 UT1 儒略日 / UT1 Julian day of the same conjunction.
	RightAscensionConjunctionJD float64
	// DeltaTSeconds 食甚时刻实际使用的 ΔT / ΔT used at greatest eclipse.
	DeltaTSeconds float64
	// SunRadiusModel SunSemidiameterArcsec 所用的太阳半径口径 / solar radius convention behind SunSemidiameterArcsec.
	SunRadiusModel SolarEclipseSunRadiusModel

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
	// PenumbralK and UmbralK are the Moon-to-Earth radius ratios k1 and k2.
	PenumbralK float64
	UmbralK    float64
	// BodyShiftLongitudeArcsec 与 BodyShiftLatitudeArcsec 是星历表里的 Δl/Δb；本库模型不做月面位置平移，恒为零。
	// BodyShiftLongitudeArcsec and BodyShiftLatitudeArcsec are the Δl/Δb of published element tables;
	// this model shifts nothing on the lunar surface, so both stay zero.
	BodyShiftLongitudeArcsec float64
	BodyShiftLatitudeArcsec  float64
	// Ephemeris 是所用模型名称。
	// Ephemeris names the model in use.
	Ephemeris string
	// SingleK 表示使用的是 IAU Single-K（k1 同时用于半影与本影）。
	// SingleK reports the IAU Single-K convention, where k1 serves both the penumbra and the umbra.
	SingleK bool
}

// horizontalParallaxArcsec 由视半径换算地平视差：sin(HP) = sin(SD) × R⊕ / R天体。
func horizontalParallaxArcsec(semidiameterArcsec, radiusRatio float64) float64 {
	sine := math.Sin(semidiameterArcsec / 3600 * math.Pi / 180)
	return math.Asin(math.Max(-1, math.Min(1, sine*radiusRatio))) * 180 / math.Pi * 3600
}

// SolarEclipseGeocentricPanelAt 计算给定日食在食甚时刻的地心量面板 / geocentric solar eclipse panel at greatest eclipse.
func (module *Eclipse) SolarEclipseGeocentricPanelAt(date time.Time) (SolarEclipseGeocentricPanel, bool) {
	return module.SolarEclipseGeocentricPanelWithOptions(date, SolarEclipseOptions{RadiusModel: SolarEclipseModelNASABulletinSplitK})
}

// SolarEclipseGeocentricPanelIAUSingleK 使用 IAU Single-K 计算地心量面板 / geocentric solar eclipse panel with IAU Single-K.
func (module *Eclipse) SolarEclipseGeocentricPanelIAUSingleK(date time.Time) (SolarEclipseGeocentricPanel, bool) {
	return module.SolarEclipseGeocentricPanelWithOptions(date, SolarEclipseOptions{RadiusModel: SolarEclipseModelIAUSingleK})
}

// SolarEclipseGeocentricPanelWithOptions 使用自定义半径口径计算食甚时刻的地心量面板 / geocentric solar eclipse panel with custom radius options.
func (module *Eclipse) SolarEclipseGeocentricPanelWithOptions(date time.Time, options SolarEclipseOptions) (SolarEclipseGeocentricPanel, bool) {
	info, ok := module.SolarEclipseOnDateWithOptions(date, options)
	if !ok {
		return SolarEclipseGeocentricPanel{}, false
	}
	panel := solarEclipseGeocentricPanelAt(info, module)
	if info.Model == SolarEclipseModelIAUSingleK {
		panel.Ephemeris = "IAU Single-K"
		panel.PenumbralK = basic.SolarEclipseIAUSingleRadiusK
		panel.UmbralK = basic.SolarEclipseIAUSingleRadiusK
		panel.SingleK = true
		return panel, true
	}
	panel.Ephemeris = "NASA bulletin Split-K"
	panel.PenumbralK = basic.SolarEclipsePenumbralK
	panel.UmbralK = basic.SolarEclipseUmbralK
	return panel, true
}

func solarEclipseGeocentricPanelAt(info SolarEclipseInfo, modules ...*Eclipse) SolarEclipseGeocentricPanel {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	tt := solarEclipseTimeToTTJDE(info.GreatestEclipse)
	conjunctionJDE := basic.CalcMoonSHByJDEWithInstance(tt, 0, module.inst)
	rightAscensionJDE := solarEclipseRightAscensionConjunction(tt, module)
	panel := SolarEclipseGeocentricPanel{
		Conjunction:                  solarEclipseTTJDEToTime(conjunctionJDE, info.GreatestEclipse.Location()),
		ConjunctionJDE:               conjunctionJDE,
		RightAscensionConjunction:    solarEclipseTTJDEToTime(rightAscensionJDE, info.GreatestEclipse.Location()),
		RightAscensionConjunctionJDE: rightAscensionJDE,
		RightAscensionConjunctionJD:  rightAscensionJDE - basic.DeltaT(tt, true)/86400,
		DeltaTSeconds:                basic.DeltaT(tt, true),
	}
	panel.SunRightAscensionDeg, panel.SunDeclinationDeg = basic.SunApparentRaDecWithInstance(tt, module.inst)
	panel.MoonRightAscensionDeg, panel.MoonDeclinationDeg = basic.HMoonTrueRaDecWithInstance(tt, module.inst)
	panel.SunRadiusModel = info.SunRadiusModel
	panel.SunSemidiameterArcsec = basic.SolarEclipseSunSemidiameterWithInstance(tt, info.SunRadiusModel, module.inst)
	panel.MoonSemidiameterArcsec = basic.MoonSemidiameterWithInstance(tt, module.inst)
	panel.SunParallaxArcsec = horizontalParallaxArcsec(panel.SunSemidiameterArcsec, horizontalParallaxSunRatio)
	panel.MoonParallaxArcsec = horizontalParallaxArcsec(panel.MoonSemidiameterArcsec, horizontalParallaxMoonRatio)
	physical := basic.MoonPhysicalWithInstance(tt, module.inst)
	panel.LibrationLongitudeDeg = physical.LibrationLongitude
	panel.LibrationLatitudeDeg = physical.LibrationLatitude
	panel.LibrationPositionAngleDeg = physical.PositionAngle
	panel.BrownLunationNumber = int(math.Floor((panel.ConjunctionJDE-brownLunationEpochJDE)/brownLunationSynodicMonth)) + 1
	return panel
}

// solarEclipseRightAscensionGap 返回月亮与太阳的地心视赤经差，单位弧度。
func solarEclipseRightAscensionGap(tt float64, modules ...*Eclipse) float64 {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	sunRightAscension, _ := basic.SunApparentRaDecWithInstance(tt, module.inst)
	moonRightAscension, _ := basic.HMoonTrueRaDecWithInstance(tt, module.inst)
	return math.Remainder(moonRightAscension-sunRightAscension, 360) * math.Pi / 180
}

// solarEclipseRightAscensionConjunction 求地心视赤经相等的时刻。
// 赤经差在合附近单调，用牛顿法从食甚时刻收敛即可，不需要预先求朔。
func solarEclipseRightAscensionConjunction(tt float64, modules ...*Eclipse) float64 {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	estimate := tt
	for iteration := 0; iteration < 40; iteration++ {
		value := solarEclipseRightAscensionGap(estimate, module)
		if math.Abs(value) < 1e-10 {
			break
		}
		derivative := (solarEclipseRightAscensionGap(estimate+1e-5, module) - solarEclipseRightAscensionGap(estimate-1e-5, module)) / 2e-5
		if derivative == 0 || math.IsNaN(derivative) || math.IsInf(derivative, 0) {
			break
		}
		estimate -= value / derivative
	}
	return estimate
}
