package basic

import (
	"math"
	"sort"
)

const (
	schaeferNightSkyNanolambert = 180.0
	schaeferNanolambertToErg    = 1.02e-15
	schaeferSunMagnitude        = -26.74
	schaeferZeroMagnitudeRef    = -11.05
	schaeferMinSeparationDeg    = 0.1
	schaeferTwilightLowSunDeg   = -3.0
	schaeferDaylightHighSunDeg  = 4.0
	schaeferScotopicLimitNL     = 1479.1
)

// HeliacalOptions 偕日可见计算参数；零值使用消光系数 0.25 星等/大气质量、目标最低高度 0.5° / heliacal visibility options.
type HeliacalOptions struct {
	// ExtinctionCoefficient 大气消光系数（星等/大气质量）/ extinction in magnitudes per airmass.
	ExtinctionCoefficient float64
	// MinTargetAltitude 目标最低几何高度（度）/ minimum geometric target altitude in degrees.
	MinTargetAltitude float64
	// Moonlight 为 true 时计入月光散射（Krisciunas–Schaefer）；默认不计 / include scattered moonlight when true.
	Moonlight bool
}

func (options HeliacalOptions) normalized() HeliacalOptions {
	if !(options.ExtinctionCoefficient > 0) {
		options.ExtinctionCoefficient = 0.25
	}
	if !(options.MinTargetAltitude > 0) {
		options.MinTargetAltitude = 0.5
	}
	return options
}

func schaeferAirmass(altitudeDeg float64) float64 {
	z := math.Min(90, math.Max(0, 90-altitudeDeg)) * rad
	c := math.Cos(z)
	return 1 / (c + 0.025*math.Exp(-11*c))
}

func schaeferTransmission(k, airmass float64) float64 { return math.Pow(10, -0.4*k*airmass) }

func schaeferPhase(separationDeg float64) float64 {
	s := math.Max(schaeferMinSeparationDeg, separationDeg)
	c := math.Cos(s * rad)
	return 6.2e7/(s*s) + math.Pow(10, 6.15-s/40) + math.Pow(10, 5.36)*(1.06+c*c)
}

func validExtinction(extinction float64) bool {
	return extinction > 0 && !math.IsInf(extinction, 0)
}

// SkyBrightnessNanolambert 无月夜晨昏天光亮度（nL，Schaefer 模型），消光系数须为正，否则返回 NaN / twilight sky brightness without moonlight, Schaefer model.
func SkyBrightnessNanolambert(targetAltitude, sunAltitude, separation, extinction float64) float64 {
	if !validExtinction(extinction) {
		return math.NaN()
	}
	k := extinction
	targetT := schaeferTransmission(k, schaeferAirmass(targetAltitude))
	sunT := schaeferTransmission(k, schaeferAirmass(sunAltitude))
	targetZ := math.Min(90, math.Max(0, 90-targetAltitude))
	sunZ := math.Min(180, math.Max(0, 90-sunAltitude))
	sep := math.Max(schaeferMinSeparationDeg, separation)
	scale := math.Pow(10, -0.4*(schaeferSunMagnitude-schaeferZeroMagnitudeRef+43.27))
	day := scale * (1 - targetT) * (schaeferPhase(sep)*sunT + 440000*(1-sunT))
	twilight := math.Pow(10, -0.4*(schaeferSunMagnitude-schaeferZeroMagnitudeRef-57.5+sunZ-targetZ/(360*k))) *
		(1 - targetT) * 100 / sep
	sky := math.Min(day, twilight)
	switch {
	case sunAltitude < schaeferTwilightLowSunDeg:
		sky = twilight
	case sunAltitude > schaeferDaylightHighSunDeg:
		sky = day
	}
	sz := math.Sin(targetZ * rad)
	night := schaeferNightSkyNanolambert * (0.4 + 0.6/math.Sqrt(1-0.96*sz*sz)) * targetT
	return sky/schaeferNanolambertToErg + night
}

// LimitingMagnitude 肉眼极限星等（已扣除目标方向的大气消光），消光系数须为正，否则返回 NaN / naked-eye limiting magnitude after extinction at the target altitude.
func LimitingMagnitude(targetAltitude, sunAltitude, separation, extinction float64) float64 {
	return limitingMagnitudeForSky(SkyBrightnessNanolambert(targetAltitude, sunAltitude, separation, extinction), targetAltitude, extinction)
}

// MoonlightNanolambert 月光在目标方向造成的天光亮度（nL，Krisciunas–Schaefer）；月在地平下为 0，消光系数非正为 NaN /
// sky brightness from scattered moonlight; zero when the Moon is below the horizon.
func MoonlightNanolambert(targetAltitude, moonAltitude, moonSeparation, moonPhaseAngle, extinction float64) float64 {
	if !validExtinction(extinction) {
		return math.NaN()
	}
	if moonAltitude <= 0 || targetAltitude <= 0 {
		return 0
	}
	alpha := math.Abs(moonPhaseAngle)
	mag := -12.73 + 0.026*alpha + 4e-9*math.Pow(alpha, 4)
	illuminance := math.Pow(10, -0.4*(mag+16.57))
	rho := math.Max(schaeferMinSeparationDeg, moonSeparation)
	c := math.Cos(rho * rad)
	scatter := math.Pow(10, 5.36)*(1.06+c*c) + math.Pow(10, 6.15-rho/40)
	airmass := func(alt float64) float64 {
		s := math.Sin((90 - alt) * rad)
		return 1 / math.Sqrt(1-0.96*s*s)
	}
	return scatter * illuminance * math.Pow(10, -0.4*extinction*airmass(moonAltitude)) *
		(1 - math.Pow(10, -0.4*extinction*airmass(targetAltitude)))
}

func limitingMagnitudeForSky(b, targetAltitude, extinction float64) float64 {
	c, k := 4.4668359215096e-9, 1.2589254117942e-6
	if b < schaeferScotopicLimitNL {
		c, k = 1.5848931924611e-10, 0.012589254117942
	}
	threshold := c * math.Pow(1+math.Sqrt(k*b), 2)
	return -16.57 - 2.5*math.Log10(threshold) - extinction*schaeferAirmass(targetAltitude)
}

// HeliacalEventKind 偕日事件类型 / heliacal event kind.
type HeliacalEventKind int

const (
	// HeliacalMorningFirst 晨始见（偕日升）/ morning first visibility.
	HeliacalMorningFirst HeliacalEventKind = iota
	// HeliacalMorningLast 晨末见 / morning last visibility.
	HeliacalMorningLast
	// HeliacalEveningFirst 昏始见 / evening first visibility.
	HeliacalEveningFirst
	// HeliacalEveningLast 昏末见（偕日落）/ evening last visibility.
	HeliacalEveningLast
)

// HeliacalEvent 偕日事件：JD 为当天最佳可见时刻（UT 儒略日），Margin 为极限星等减目标星等 / heliacal event at the best moment of that twilight.
type HeliacalEvent struct {
	Kind           HeliacalEventKind
	JD             float64
	Margin         float64
	TargetAltitude float64
	SunAltitude    float64
}

// heliacalTarget 返回 TT 时刻目标的视赤经、赤纬（度）与视星等。
type heliacalTarget func(jde float64) (ra, dec, magnitude float64)

type heliacalBest struct {
	jd, margin, targetAlt, sunAlt float64
	visible                       bool
}

func heliacalAltitude(ra, dec, lst, lat float64) float64 {
	h := (lst - ra) * rad
	d, p := dec*rad, lat*rad
	return math.Asin(math.Sin(p)*math.Sin(d)+math.Cos(p)*math.Cos(d)*math.Cos(h)) / rad
}

// heliacalTwilight 扫描一次晨（morning）或昏的太阳高度 −18°..0° 时段，取可见余量最大的样本。
// 日月位置在窗口中点取一次，窗口不足两小时，引入的高度误差远小于模型本身。
func heliacalTwilight(dayJD, lon, lat float64, morning bool, target heliacalTarget, options HeliacalOptions, inst *Instance) heliacalBest {
	center := dayJD - lon/360 + 0.5
	if morning {
		center -= 0.25
	} else {
		center += 0.25
	}
	jde := UTC2TT(center)
	sunRA, sunDec := HSunApparentRaDecWithInstance(jde, inst)
	ra, dec, mag := target(jde)
	sep := skyAngularDistance(ra, dec, sunRA, sunDec)
	best := heliacalBest{margin: math.Inf(-1)}
	if math.IsNaN(ra) || math.IsNaN(mag) {
		return best
	}
	var moonRA, moonDec, moonSep, moonPhase float64
	if options.Moonlight {
		lo, bo := SkyBodyApparentLoBoWithInstance(SkyMoon, jde, inst)
		moonRA, moonDec = LoBoToRaDec(jde, lo, bo)
		moonSep = skyAngularDistance(ra, dec, moonRA, moonDec)
		moonPhase = 180 - skyAngularDistance(moonRA, moonDec, sunRA, sunDec)
	}
	lst0 := ApparentSiderealTime(UTC2UT1(center))*15 + lon
	const step = 2.0 / 1440
	for dt := -0.25; dt <= 0.25; dt += step {
		lst := lst0 + dt*360.98564736629
		sunAlt := heliacalAltitude(sunRA, sunDec, lst, lat)
		if sunAlt > 0 || sunAlt < -18 {
			continue
		}
		alt := heliacalAltitude(ra, dec, lst, lat)
		if alt < options.MinTargetAltitude {
			continue
		}
		sky := SkyBrightnessNanolambert(alt, sunAlt, sep, options.ExtinctionCoefficient)
		if options.Moonlight {
			sky += MoonlightNanolambert(alt, heliacalAltitude(moonRA, moonDec, lst, lat), moonSep, moonPhase, options.ExtinctionCoefficient)
		}
		margin := limitingMagnitudeForSky(sky, alt, options.ExtinctionCoefficient) - mag
		if margin > best.margin {
			best = heliacalBest{center + dt, margin, alt, sunAlt, margin >= 0}
		}
	}
	return best
}

// heliacalMaxDays 逐日扫描的天数上限（约 2.3 万年），挡住非法跨度造成的近乎无界扫描。
const heliacalMaxDays = 1 << 23

// findHeliacalEvents 逐日扫描晨、昏可见性，从不可见变为可见记“始见”，反之记“末见”（取末个可见日）。
func findHeliacalEvents(startJD, endJD, lon, lat float64, target heliacalTarget, options HeliacalOptions, inst *Instance) []HeliacalEvent {
	options = options.normalized()
	if !finite(startJD) || !finite(endJD) || !(endJD > startJD) || math.Abs(lat) >= 90 {
		return nil
	}
	day0 := math.Floor(startJD-0.5) + 0.5
	span := math.Floor(endJD - day0)
	if !(span < heliacalMaxDays) {
		return nil
	}
	days := int(span)
	var out []HeliacalEvent
	for _, morning := range []bool{true, false} {
		first, last := HeliacalEveningFirst, HeliacalEveningLast
		if morning {
			first, last = HeliacalMorningFirst, HeliacalMorningLast
		}
		prev := heliacalTwilight(day0-1, lon, lat, morning, target, options, inst)
		for i := 0; i <= days; i++ {
			cur := heliacalTwilight(day0+float64(i), lon, lat, morning, target, options, inst)
			// 事件时刻是采样日内的细化值，可能越过 endJD，所以上界按事件时刻判断，不能按采样日判断。
			switch {
			case cur.visible && !prev.visible && cur.jd >= startJD && cur.jd <= endJD:
				out = append(out, HeliacalEvent{first, cur.jd, cur.margin, cur.targetAlt, cur.sunAlt})
			case !cur.visible && prev.visible && prev.jd >= startJD && prev.jd <= endJD:
				out = append(out, HeliacalEvent{last, prev.jd, prev.margin, prev.targetAlt, prev.sunAlt})
			}
			prev = cur
		}
	}
	sortHeliacalEvents(out)
	return out
}

func sortHeliacalEvents(events []HeliacalEvent) {
	sort.SliceStable(events, func(i, j int) bool { return events[i].JD < events[j].JD })
}

// FindPlanetHeliacalEventsWithInstance 行星偕日事件（UT 儒略日范围，经度东正）/ planet heliacal events in a UT JD range.
func FindPlanetHeliacalEventsWithInstance(body SkyBody, startJD, endJD, lon, lat float64, options HeliacalOptions, inst *Instance) []HeliacalEvent {
	if body < SkyMercury || body > SkyNeptune {
		return nil
	}
	index := int(body-SkyMercury) + 1
	target := func(jde float64) (float64, float64, float64) {
		lo, bo := SkyBodyApparentLoBoWithInstance(body, jde, inst)
		ra, dec := LoBoToRaDec(jde, lo, bo)
		return ra, dec, planetMagnitude2018(index, jde, inst)
	}
	return findHeliacalEvents(startJD, endJD, lon, lat, target, options, inst)
}

// FindStarHeliacalEventsWithInstance 恒星偕日事件；magnitude 为恒星 V 星等 / star heliacal events with the star's V magnitude.
func FindStarHeliacalEventsWithInstance(star StarCoordinate, magnitude, startJD, endJD, lon, lat float64, options HeliacalOptions, inst *Instance) []HeliacalEvent {
	if math.IsNaN(magnitude) || math.IsInf(magnitude, 0) {
		return nil
	}
	target := func(jde float64) (float64, float64, float64) {
		ra, dec := starApparentRaDecGeocentric(jde, star, inst)
		return ra, dec, magnitude
	}
	return findHeliacalEvents(startJD, endJD, lon, lat, target, options, inst)
}
