package basic

import (
	"math"
	"time"
)

// AltitudeLimb 日月取高度的位置；行星与恒星忽略 / limb used for Sun and Moon altitude.
type AltitudeLimb int

const (
	// AltitudeLimbUpper 上缘，与日出日落口径相同 / upper limb.
	AltitudeLimbUpper AltitudeLimb = iota
	// AltitudeLimbCenter 中心 / centre.
	AltitudeLimbCenter
	// AltitudeLimbLower 下缘 / lower limb.
	AltitudeLimbLower
)

// AltitudeCrossingOptions 任意高度升落选项；零值为上缘、标准气象（1010 hPa、10 °C）、计入折射 / options for altitude crossings.
type AltitudeCrossingOptions struct {
	Limb AltitudeLimb
	// NoRefraction 为 true 时目标高度按几何高度比较 / compare geometric altitude when true.
	NoRefraction bool
	// PressureHPa 为 0 时用标准气压与气温 / zero selects the standard atmosphere.
	PressureHPa  float64
	TemperatureC float64
}

// AltitudeCrossingState 一个民用日内的穿越状态 / crossing state over one civil day.
type AltitudeCrossingState int

const (
	// AltitudeCrosses 至少有一次升或落 / at least one crossing.
	AltitudeCrosses AltitudeCrossingState = iota
	// AltitudeAlwaysAbove 全天高于目标高度 / always above the target altitude.
	AltitudeAlwaysAbove
	// AltitudeAlwaysBelow 全天低于目标高度 / always below the target altitude.
	AltitudeAlwaysBelow
	// AltitudeTangent 极值与目标高度相差不足 1″，视为相切 / extremum within 1″ of the target.
	AltitudeTangent
)

// AltitudeCrossing 民用日内首次升、落（本地儒略日，无则为 NaN）与残差范围（度）/ first rise and set in local JD with residual range in degrees.
type AltitudeCrossing struct {
	Rise, Set   float64
	State       AltitudeCrossingState
	MinResidual float64
	MaxResidual float64
}

// AltitudeCrossingResult date 所在民用日内首次升、落（无则为零值，对应 HasRise/HasSet 为 false）与残差范围（度）/ first rise and set on the civil day.
type AltitudeCrossingResult struct {
	Rise, Set       time.Time
	HasRise, HasSet bool
	State           AltitudeCrossingState
	MinResidual     float64
	MaxResidual     float64
}

const (
	altitudeScanStepDays   = 10.0 / 1440
	altitudeTangentDegrees = 1.0 / 3600
)

func (options AltitudeCrossingOptions) refraction(trueAltitude float64) float64 {
	if options.NoRefraction {
		return 0
	}
	if options.PressureHPa == 0 {
		return RefractionFromTrueAltitude(trueAltitude, refractionStandardPressureHPa, refractionStandardTemperatureC)
	}
	return RefractionFromTrueAltitude(trueAltitude, options.PressureHPa, options.TemperatureC)
}

func (options AltitudeCrossingOptions) limbOffset(semidiameterDeg float64) float64 {
	switch options.Limb {
	case AltitudeLimbUpper:
		return semidiameterDeg
	case AltitudeLimbLower:
		return -semidiameterDeg
	}
	return 0
}

func (options AltitudeCrossingOptions) valid() bool {
	if options.Limb < AltitudeLimbUpper || options.Limb > AltitudeLimbLower {
		return false
	}
	return options.NoRefraction || options.PressureHPa == 0 || !math.IsNaN(RefractionFromTrueAltitude(0, options.PressureHPa, options.TemperatureC))
}

// SkyAltitudeCrossingWithInstance 日、月或行星在本地民用日 julianDay 内穿过视高度 targetAltitude（度，相对天文地平）的时刻 /
// Sun, Moon or planet crossings of an apparent altitude on the local civil day.
//
// height 只影响月球站心视差；需要以下沉地平为基准时由调用方从 targetAltitude 中减去地平俯角。
//
// 这是按儒略日窗口求解的低级接口：julianDay 与 timeZone（小时）由调用方拼出本地民用日，
// 固定时区不处理夏令时切换。julianDay 须取本地零点（x.5），正午及之后的值会落到次日窗口。一般用途请用 SkyAltitudeCrossingOnDate，它从 time.Time 推出
// 当天的真实时区窗口 / low-level julian-day window API; prefer SkyAltitudeCrossingOnDate.
func SkyAltitudeCrossingWithInstance(body SkyBody, julianDay, longitude, latitude, timeZone, height, targetAltitude float64,
	options AltitudeCrossingOptions, inst *Instance) (AltitudeCrossing, error) {
	altitude, err := skyAltitudeFunc(body, longitude, latitude, timeZone, height, inst)
	if err != nil {
		return AltitudeCrossing{}, err
	}
	start := math.Floor(julianDay) + 0.5
	return altitudeCrossing(start, start+1, latitude, height, targetAltitude, options, altitude)
}

// SkyAltitudeCrossingOnDate 日、月或行星在 date 所在民用日内穿过视高度 targetAltitude 的时刻；夏令时切换日按实际日长搜索 /
// Sun, Moon or planet crossings of an apparent altitude on the civil day of date.
func SkyAltitudeCrossingOnDate(body SkyBody, date time.Time, longitude, latitude, height, targetAltitude float64,
	options AltitudeCrossingOptions, inst *Instance) (AltitudeCrossingResult, error) {
	start, end, timeZone := civilDayWindow(date)
	altitude, err := skyAltitudeFunc(body, longitude, latitude, timeZone, height, inst)
	if err != nil {
		return AltitudeCrossingResult{}, err
	}
	raw, err := altitudeCrossing(start, end, latitude, height, targetAltitude, options, altitude)
	return altitudeCrossingOnDate(date, timeZone, raw, err)
}

// StarAltitudeCrossing 恒星（视赤经、视赤纬，度）在本地民用日内穿过视高度 targetAltitude 的时刻 / star crossings of an apparent altitude.
//
// julianDay 须取本地零点（x.5），正午及之后的值会落到次日窗口；一般用途请用 StarAltitudeCrossingOnDate。
func StarAltitudeCrossing(julianDay, ra, dec, longitude, latitude, timeZone, targetAltitude float64, options AltitudeCrossingOptions) (AltitudeCrossing, error) {
	if !isFiniteFloat(ra) || !isFiniteFloat(dec) {
		return AltitudeCrossing{}, ErrInvalidObservationInput
	}
	start := math.Floor(julianDay) + 0.5
	return altitudeCrossing(start, start+1, latitude, 0, targetAltitude, options, starAltitudeFunc(ra, dec, longitude, latitude, timeZone))
}

// StarAltitudeCrossingOnDate 恒星在 date 所在民用日内穿过视高度 targetAltitude 的时刻；夏令时切换日按实际日长搜索 /
// star crossings of an apparent altitude on the civil day of date.
func StarAltitudeCrossingOnDate(date time.Time, ra, dec, longitude, latitude, targetAltitude float64, options AltitudeCrossingOptions) (AltitudeCrossingResult, error) {
	if !isFiniteFloat(ra) || !isFiniteFloat(dec) {
		return AltitudeCrossingResult{}, ErrInvalidObservationInput
	}
	start, end, timeZone := civilDayWindow(date)
	raw, err := altitudeCrossing(start, end, latitude, 0, targetAltitude, options, starAltitudeFunc(ra, dec, longitude, latitude, timeZone))
	return altitudeCrossingOnDate(date, timeZone, raw, err)
}

// civilDayWindow 以当天零点的偏移固定本地时标，窗口为真实日长（夏令时切换日 23 或 25 小时）。
func civilDayWindow(date time.Time) (start, end, timeZone float64) {
	year, month, day := date.Date()
	midnight := time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	next := time.Date(year, month, day+1, 0, 0, 0, 0, date.Location())
	_, offset := midnight.Zone()
	start = Date2JD(midnight)
	return start, start + next.Sub(midnight).Hours()/24, float64(offset) / 3600
}

func altitudeCrossingOnDate(date time.Time, timeZone float64, raw AltitudeCrossing, err error) (AltitudeCrossingResult, error) {
	if err != nil {
		return AltitudeCrossingResult{}, err
	}
	convert := func(jd float64) (time.Time, bool) {
		if math.IsNaN(jd) {
			return time.Time{}, false
		}
		return JD2DateByZone(jd-timeZone/24, date.Location(), false), true
	}
	out := AltitudeCrossingResult{State: raw.State, MinResidual: raw.MinResidual, MaxResidual: raw.MaxResidual}
	out.Rise, out.HasRise = convert(raw.Rise)
	out.Set, out.HasSet = convert(raw.Set)
	return out, nil
}

func skyAltitudeFunc(body SkyBody, longitude, latitude, timeZone, height float64, inst *Instance) (func(jd float64) (float64, float64), error) {
	switch {
	case body == SkySun:
		return func(jd float64) (float64, float64) {
			state := sunObservationStateN(jd, longitude, latitude, timeZone, -1, inst)
			return state.altitude, angularSemidiameterFromAU(sunEquatorialRadiusKM, state.distanceAU) / 3600
		}, nil
	case body == SkyMoon:
		return func(jd float64) (float64, float64) {
			state := hMoonObservationStateN(jd, longitude, latitude, timeZone, height, -1, inst)
			return state.altitude, angularSemidiameterArcsec(moonEquatorialRadiusKM, state.distanceKM) / 3600
		}, nil
	case body.valid():
		heightFn := skyPlanetHeight(body)
		return func(jd float64) (float64, float64) { return heightFn(jd, longitude, latitude, timeZone, inst), 0 }, nil
	}
	return nil, ErrInvalidObservationInput
}

func starAltitudeFunc(ra, dec, longitude, latitude, timeZone float64) func(jd float64) (float64, float64) {
	return func(jd float64) (float64, float64) {
		return StarHeight(jd, ra, dec, longitude, latitude, timeZone), 0
	}
}

func skyPlanetHeight(body SkyBody) func(localJD, lon, lat, timezone float64, inst *Instance) float64 {
	switch body {
	case SkyMercury:
		return MercuryHeightWithInstance
	case SkyVenus:
		return VenusHeightWithInstance
	case SkyMars:
		return MarsHeightWithInstance
	case SkyJupiter:
		return JupiterHeightWithInstance
	case SkySaturn:
		return SaturnHeightWithInstance
	case SkyUranus:
		return UranusHeightWithInstance
	}
	return NeptuneHeightWithInstance
}

func altitudeCrossing(start, end, latitude, height, targetAltitude float64, options AltitudeCrossingOptions,
	altitude func(jd float64) (float64, float64)) (AltitudeCrossing, error) {
	if !isFiniteFloat(start) || !isFiniteFloat(end) || !(end > start) || !isFiniteFloat(latitude) || math.Abs(latitude) > 90 ||
		!isFiniteFloat(height) || !isFiniteFloat(targetAltitude) || math.Abs(targetAltitude) > 90 || !options.valid() {
		return AltitudeCrossing{}, ErrInvalidObservationInput
	}
	residual := func(jd float64) float64 {
		alt, semidiameter := altitude(jd)
		return alt + options.refraction(alt) + options.limbOffset(semidiameter) - targetAltitude
	}
	result := AltitudeCrossing{Rise: math.NaN(), Set: math.NaN(), MinResidual: math.Inf(1), MaxResidual: math.Inf(-1)}
	prevT, prevR := start, residual(start)
	if math.IsNaN(prevR) {
		return AltitudeCrossing{}, ErrInvalidObservationInput
	}
	track := func(t, r float64) {
		result.MinResidual = math.Min(result.MinResidual, r)
		result.MaxResidual = math.Max(result.MaxResidual, r)
	}
	track(prevT, prevR)
	record := func(root float64, rising bool) {
		if rising && math.IsNaN(result.Rise) {
			result.Rise = root
		} else if !rising && math.IsNaN(result.Set) {
			result.Set = root
		}
	}
	// 同号区间内朝目标一侧的极值可能越过目标高度，采样点看不到这两次穿越。
	refine := func(a, b, ra float64) {
		peak, rp := orbitalGoldenExtremum(residual, a, b, ra < 0)
		if math.IsNaN(rp) {
			return
		}
		track(peak, rp)
		if (rp < 0) == (ra < 0) || peak <= a || peak >= b {
			return
		}
		record(orbitalBisect(residual, a, peak, ra), ra < 0)
		record(orbitalBisect(residual, peak, b, rp), ra >= 0)
	}
	prev2T, prev2R, samples := prevT, prevR, 1
	for t := start + altitudeScanStepDays; prevT < end; t += altitudeScanStepDays {
		t = math.Min(t, end)
		r := residual(t)
		if math.IsNaN(r) {
			return AltitudeCrossing{}, ErrInvalidObservationInput
		}
		track(t, r)
		sameSign := (prevR < 0) == (r < 0)
		switch {
		case !sameSign:
			record(orbitalBisect(residual, prevT, t, prevR), r >= 0)
		case samples == 1:
			refine(prevT, t, prevR)
		case (prev2R < 0) == (prevR < 0) && altitudeTowardTarget(prev2R, prevR, r):
			refine(prev2T, t, prev2R)
		}
		prev2T, prev2R = prevT, prevR
		prevT, prevR = t, r
		samples++
	}
	if samples > 2 && (prev2R < 0) == (prevR < 0) {
		refine(prev2T, prevT, prev2R)
	}
	switch {
	case !math.IsNaN(result.Rise) || !math.IsNaN(result.Set):
		result.State = AltitudeCrosses
	case math.Abs(result.MinResidual) < altitudeTangentDegrees || math.Abs(result.MaxResidual) < altitudeTangentDegrees:
		result.State = AltitudeTangent
	case result.MinResidual > 0:
		result.State = AltitudeAlwaysAbove
	default:
		result.State = AltitudeAlwaysBelow
	}
	return result, nil
}

// altitudeTowardTarget 报告中间样本是否为朝零点一侧的局部极值：负残差取极大，正残差取极小。
func altitudeTowardTarget(left, middle, right float64) bool {
	if middle < 0 {
		return middle >= left && middle >= right
	}
	return middle <= left && middle <= right
}
