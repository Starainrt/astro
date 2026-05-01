package basic

import (
	"math"
	"sort"
)

const (
	localSolarEclipseDiagramDefaultStepDays = 5.0 / 1440.0
	localSolarEclipseDiagramMinStepDays     = 1.0 / 86400.0
	localSolarEclipseDiagramMaxSamples      = 2000
	localSolarEclipseDiagramDuplicateDays   = 1e-10
)

// LocalSolarEclipseDiagramOptions 控制站心日食视圆图采样。
// LocalSolarEclipseDiagramOptions controls local solar eclipse disk diagram sampling.
type LocalSolarEclipseDiagramOptions struct {
	// StepDays 是路径采样步长，单位为日；<=0 时使用 5 分钟。
	// StepDays is the path sampling step in days; values <= 0 use five minutes.
	StepDays float64
}

// LocalSolarEclipseDiagramFrame 表示一个时刻的站心日月视圆几何。
// LocalSolarEclipseDiagramFrame is local Sun-Moon disk geometry at one instant.
type LocalSolarEclipseDiagramFrame struct {
	// JDE 是力学时儒略日, TT Julian ephemeris day.
	JDE float64
	// MoonX / MoonY 是以太阳视半径为单位的月心相对日心坐标，X 向东为正，Y 向北为正。
	// MoonX/MoonY are Moon-center coordinates relative to the Sun center in Sun-radius units; X east, Y north.
	MoonX float64
	MoonY float64
	// SunRadius 是太阳视半径，单位为图上太阳半径；固定为 1。
	// SunRadius is the Sun radius in diagram Sun-radius units; always 1.
	SunRadius float64
	// MoonRadius 是月球视半径，单位为图上太阳半径。
	// MoonRadius is the Moon apparent radius in Sun-radius units.
	MoonRadius float64
	// Separation 是日月中心角距，单位为度。
	// Separation is the Sun-Moon center separation in degrees.
	Separation float64
	// PositionAngle 是月心相对日心的位置角，从北点起向东量，单位为度。
	// PositionAngle is the Moon-center position angle from north toward east, in degrees.
	PositionAngle float64
	// SunAltitude / SunAzimuth 是太阳站心高度角 / 方位角，单位为度。
	// SunAltitude/SunAzimuth are local Sun altitude/azimuth in degrees.
	SunAltitude float64
	SunAzimuth  float64
	// Label 是关键阶段标签，如 C1/Greatest/C4。
	// Label is a key phase label such as C1/Greatest/C4.
	Label string
	// Labels 是该点对应的全部关键阶段标签；若事件重合，这里会有多个值。
	// Labels contains all phase labels attached to this point.
	Labels []string
}

// LocalSolarEclipseDiagramResult 表示站心日食视圆图几何结果。
// LocalSolarEclipseDiagramResult is the geometry result for a local solar eclipse disk diagram.
type LocalSolarEclipseDiagramResult struct {
	// Eclipse 是对应的站心日食结果。
	// Eclipse is the local eclipse result used for the diagram.
	Eclipse LocalSolarEclipseResult
	// Frames 是采样帧，太阳固定在 (0,0)，月球按 MoonX/MoonY 绘制。
	// Frames are sampled frames; the Sun is fixed at (0,0), and the Moon uses MoonX/MoonY.
	Frames []LocalSolarEclipseDiagramFrame
	// StepDays 是实际采用的路径采样步长，单位为日。
	// StepDays is the effective path sampling step in days.
	StepDays float64
}

type localSolarEclipseDiagramTime struct {
	jde    float64
	labels []string
}

// LocalSolarEclipseDiagram 计算站心日食视圆图几何数据，默认使用 NASA bulletin Split-K 模型。
// LocalSolarEclipseDiagram computes local solar eclipse disk diagram geometry, using NASA bulletin Split-K by default.
func LocalSolarEclipseDiagram(seedJDE, lon, lat, height float64, options LocalSolarEclipseDiagramOptions) LocalSolarEclipseDiagramResult {
	return LocalSolarEclipseDiagramNASABulletinSplitK(seedJDE, lon, lat, height, options)
}

// LocalSolarEclipseDiagramIAUSingleK 计算站心日食视圆图几何数据，使用 IAU Single-K 模型。
// LocalSolarEclipseDiagramIAUSingleK computes local solar eclipse disk diagram geometry with the IAU Single-K model.
func LocalSolarEclipseDiagramIAUSingleK(seedJDE, lon, lat, height float64, options LocalSolarEclipseDiagramOptions) LocalSolarEclipseDiagramResult {
	return localSolarEclipseDiagram(seedJDE, lon, lat, height, SolarEclipseModelIAUSingleK, options)
}

// LocalSolarEclipseDiagramNASABulletinSplitK 计算站心日食视圆图几何数据，使用 NASA bulletin Split-K 模型。
// LocalSolarEclipseDiagramNASABulletinSplitK computes local solar eclipse disk diagram geometry with the NASA bulletin Split-K model.
func LocalSolarEclipseDiagramNASABulletinSplitK(seedJDE, lon, lat, height float64, options LocalSolarEclipseDiagramOptions) LocalSolarEclipseDiagramResult {
	return localSolarEclipseDiagram(seedJDE, lon, lat, height, SolarEclipseModelNASABulletinSplitK, options)
}

func localSolarEclipseDiagram(
	seedJDE, lonDeg, latDeg, heightMeters float64,
	model SolarEclipseRadiusModel,
	options LocalSolarEclipseDiagramOptions,
) LocalSolarEclipseDiagramResult {
	options = normalizeLocalSolarEclipseDiagramOptions(options)
	eclipse := localSolarEclipse(seedJDE, lonDeg, latDeg, heightMeters, model)
	result := LocalSolarEclipseDiagramResult{
		Eclipse:  eclipse,
		StepDays: options.StepDays,
	}
	if !eclipse.HasPartial {
		return result
	}

	lonRad := lonDeg * rad
	latRad := latDeg * rad
	heightKM := heightMeters / 1000.0
	params := solarEclipseModelParams(model)
	times, stepDays := localSolarEclipseDiagramTimes(eclipse, options.StepDays)
	result.StepDays = stepDays
	result.Frames = make([]LocalSolarEclipseDiagramFrame, 0, len(times))
	for _, item := range times {
		frame := localSolarEclipseDiagramFrameAt(item.jde, lonRad, latRad, heightKM, params)
		frame.Label = localSolarEclipseDiagramPrimaryLabel(item.labels)
		frame.Labels = append([]string(nil), item.labels...)
		result.Frames = append(result.Frames, frame)
	}
	return result
}

func normalizeLocalSolarEclipseDiagramOptions(options LocalSolarEclipseDiagramOptions) LocalSolarEclipseDiagramOptions {
	if options.StepDays <= 0 || math.IsNaN(options.StepDays) || math.IsInf(options.StepDays, 0) {
		options.StepDays = localSolarEclipseDiagramDefaultStepDays
	}
	if options.StepDays < localSolarEclipseDiagramMinStepDays {
		options.StepDays = localSolarEclipseDiagramMinStepDays
	}
	return options
}

func localSolarEclipseDiagramTimes(
	eclipse LocalSolarEclipseResult,
	stepDays float64,
) ([]localSolarEclipseDiagramTime, float64) {
	startJDE := eclipse.PartialStart
	endJDE := eclipse.PartialEnd
	if startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil, stepDays
	}

	if sampleCount := int(math.Ceil((endJDE-startJDE)/stepDays)) + 1; sampleCount > localSolarEclipseDiagramMaxSamples {
		stepDays = (endJDE - startJDE) / float64(localSolarEclipseDiagramMaxSamples-1)
	}

	times := []localSolarEclipseDiagramTime{
		{jde: startJDE, labels: []string{"C1"}},
		{jde: eclipse.GreatestEclipse, labels: []string{"Greatest"}},
		{jde: endJDE, labels: []string{"C4"}},
	}
	if eclipse.HasCentral {
		times = append(times,
			localSolarEclipseDiagramTime{jde: eclipse.CentralStart, labels: []string{"C2"}},
			localSolarEclipseDiagramTime{jde: eclipse.CentralEnd, labels: []string{"C3"}},
		)
	}
	for jde := startJDE + stepDays; jde < endJDE; jde += stepDays {
		times = append(times, localSolarEclipseDiagramTime{jde: jde})
	}

	sort.SliceStable(times, func(i, j int) bool {
		if times[i].jde == times[j].jde {
			return localSolarEclipseDiagramLabelPriority(times[i].labels) < localSolarEclipseDiagramLabelPriority(times[j].labels)
		}
		return times[i].jde < times[j].jde
	})
	return uniqueLocalSolarEclipseDiagramTimes(times), stepDays
}

func uniqueLocalSolarEclipseDiagramTimes(times []localSolarEclipseDiagramTime) []localSolarEclipseDiagramTime {
	if len(times) < 2 {
		return times
	}

	unique := times[:0]
	for _, item := range times {
		if item.jde == 0 {
			continue
		}
		if len(unique) == 0 || math.Abs(item.jde-unique[len(unique)-1].jde) > localSolarEclipseDiagramDuplicateDays {
			item.labels = append([]string(nil), item.labels...)
			unique = append(unique, item)
			continue
		}
		unique[len(unique)-1].labels = mergeLocalSolarEclipseDiagramLabels(unique[len(unique)-1].labels, item.labels)
	}
	return unique
}

func mergeLocalSolarEclipseDiagramLabels(existing, incoming []string) []string {
	if len(incoming) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return append([]string(nil), incoming...)
	}
	for _, label := range incoming {
		found := false
		for _, current := range existing {
			if current == label {
				found = true
				break
			}
		}
		if !found {
			existing = append(existing, label)
		}
	}
	return existing
}

func localSolarEclipseDiagramPrimaryLabel(labels []string) string {
	for _, label := range labels {
		if label == "Greatest" {
			return label
		}
	}
	if len(labels) == 0 {
		return ""
	}
	return labels[0]
}

func localSolarEclipseDiagramLabelPriority(labels []string) int {
	if len(labels) == 0 {
		return 99
	}
	switch labels[0] {
	case "C1":
		return 0
	case "C2":
		return 1
	case "Greatest":
		return 2
	case "C3":
		return 3
	case "C4":
		return 4
	default:
		return 99
	}
}

func localSolarEclipseDiagramFrameAt(
	jdTT, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
) LocalSolarEclipseDiagramFrame {
	sunEquatorial, moonEquatorial := solarEclipseSunMoonEquatorial(jdTT)
	sunXYZ := solarEclipseLLRToXYZ(sunEquatorial[0], sunEquatorial[1], sunEquatorial[2])
	moonXYZ := solarEclipseLLRToXYZ(moonEquatorial[0], moonEquatorial[1], moonEquatorial[2])

	utJDE := TD2UT(jdTT, false)
	gst := ApparentSiderealTime(utJDE) * 15 * rad
	observerXYZ := localSolarEclipseObserverXYZ(gst, lonRad, latRad, heightKM)

	sunTopocentric := solarEclipseXYZToLLR(
		sunXYZ[0]-observerXYZ[0],
		sunXYZ[1]-observerXYZ[1],
		sunXYZ[2]-observerXYZ[2],
	)
	moonTopocentric := solarEclipseXYZToLLR(
		moonXYZ[0]-observerXYZ[0],
		moonXYZ[1]-observerXYZ[1],
		moonXYZ[2]-observerXYZ[2],
	)

	sunUnit := solarEclipseLLRToXYZ(sunTopocentric[0], sunTopocentric[1], 1)
	moonUnit := solarEclipseLLRToXYZ(moonTopocentric[0], moonTopocentric[1], 1)
	dot := localSolarEclipseClampUnit(
		sunUnit[0]*moonUnit[0] + sunUnit[1]*moonUnit[1] + sunUnit[2]*moonUnit[2],
	)

	sunRadiusRad := math.Asin(localSolarEclipseClampUnit(
		solarEclipseEarthEquatorialRadiusKM * solarEclipseSolarRadiusRatio / sunTopocentric[2],
	))
	moonRadiusRad := math.Asin(localSolarEclipseClampUnit(
		solarEclipseEarthEquatorialRadiusKM * solarEclipsePenumbralK * localSolarMoonRadiusScale / moonTopocentric[2],
	))
	if params.umbralK > solarEclipsePenumbralK {
		moonRadiusRad = math.Asin(localSolarEclipseClampUnit(
			solarEclipseEarthEquatorialRadiusKM * params.umbralK * localSolarMoonRadiusScale / moonTopocentric[2],
		))
	}

	east := [3]float64{-math.Sin(sunTopocentric[0]), math.Cos(sunTopocentric[0]), 0}
	north := [3]float64{
		-math.Cos(sunTopocentric[0]) * math.Sin(sunTopocentric[1]),
		-math.Sin(sunTopocentric[0]) * math.Sin(sunTopocentric[1]),
		math.Cos(sunTopocentric[1]),
	}
	denominator := dot
	if math.Abs(denominator) < 1e-12 {
		denominator = 1
	}

	xRad := (moonUnit[0]*east[0] + moonUnit[1]*east[1] + moonUnit[2]*east[2]) / denominator
	yRad := (moonUnit[0]*north[0] + moonUnit[1]*north[1] + moonUnit[2]*north[2]) / denominator
	positionAngle := math.Atan2(xRad, yRad) / rad
	if positionAngle < 0 {
		positionAngle += 360
	}

	sunHorizontal := solarEclipseEquatorialToHorizontal(
		sunTopocentric[0],
		sunTopocentric[1],
		sunTopocentric[2],
		lonRad,
		latRad,
		gst,
	)

	return LocalSolarEclipseDiagramFrame{
		JDE:           jdTT,
		MoonX:         xRad / sunRadiusRad,
		MoonY:         yRad / sunRadiusRad,
		SunRadius:     1,
		MoonRadius:    moonRadiusRad / sunRadiusRad,
		Separation:    math.Acos(dot) / rad,
		PositionAngle: positionAngle,
		SunAltitude:   sunHorizontal[1] / rad,
		SunAzimuth:    solarEclipseNormalizeRadians(sunHorizontal[0]+math.Pi) / rad,
	}
}
