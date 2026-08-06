package basic

import (
	"math"
	"sort"
)

const (
	planetOccultationDiagramDefaultStepDays = 2.0 / 1440.0
	planetOccultationDiagramMinStepDays     = 1.0 / 86400.0
	planetOccultationDiagramMaxSamples      = 2000
	planetOccultationDiagramDuplicateDays   = 1e-10
	planetOccultationDiagramGeometryArcsec  = 0.05
	planetOccultationDiagramPositionDeg     = 0.01
	planetOccultationDiagramContactTimeDays = 1e-6
)

// PlanetOccultationDiagramOptions 控制本地行星月掩轨迹采样。
// PlanetOccultationDiagramOptions controls local planetary-occultation track sampling.
type PlanetOccultationDiagramOptions struct {
	// StepDays 是请求的轨迹采样步长，单位为日；非正值或非有限值使用两分钟，正值小于一秒时使用一秒。长事件可能增大实际步长，使基础轨迹不超过 2000 个采样点；必要阶段帧仍会额外保留。结果会报告实际采用的值。
	// StepDays is the requested track sampling step in days. Non-positive or non-finite values use two minutes, and positive values below one second use one second. Long events may increase the effective step to keep the base track within 2000 samples; required phase frames are retained in addition. The result reports the effective value.
	StepDays float64
}

// PlanetOccultationDiagramFrame 描述一个时刻的站心月球与行星几何。
// PlanetOccultationDiagramFrame describes topocentric Moon-planet geometry at one instant.
type PlanetOccultationDiagramFrame struct {
	// JDE 是 TT 儒略历书日。
	// JDE is the TT Julian ephemeris day.
	JDE float64
	// PlanetXArcsec 和 PlanetYArcsec 是相对月心的切平面偏移，单位为角秒。X 向东为正，Y 向北为正。
	// PlanetXArcsec and PlanetYArcsec are tangent-plane offsets from the lunar center. X is positive east and Y is positive north.
	PlanetXArcsec float64
	PlanetYArcsec float64
	// MoonRadiusArcsec 和 PlanetRadiusArcsec 是站心视半径，单位为角秒。
	// MoonRadiusArcsec and PlanetRadiusArcsec are topocentric apparent semidiameters.
	MoonRadiusArcsec   float64
	PlanetRadiusArcsec float64
	// SeparationArcsec 和 PositionAngleDeg 描述行星中心相对月心的位置。
	// SeparationArcsec and PositionAngleDeg describe the planet center relative to the lunar center.
	SeparationArcsec float64
	PositionAngleDeg float64
	// MoonAltitudeDeg 和 MoonAzimuthDeg 是站心地平坐标。
	// MoonAltitudeDeg and MoonAzimuthDeg are topocentric horizontal coordinates.
	MoonAltitudeDeg float64
	MoonAzimuthDeg  float64
	// DisksOverlap 表示两个视盘面存在正面积交集。
	// DisksOverlap is true while the two apparent disks have a positive-area intersection.
	DisksOverlap bool
	// FullyOcculted 表示行星盘面完全位于月缘内侧。
	// FullyOcculted is true while the planet disk lies strictly inside the lunar limb.
	FullyOcculted bool
	// Label 是主阶段标识；Labels 在掠掩事件中保留重合阶段。
	// Label is the primary key phase; Labels retains coincident phases for grazing events.
	Label  string
	Labels []string
}

// PlanetOccultationDiagramResult 包含固定地点行星月掩的几何数据。
// PlanetOccultationDiagramResult contains geometry for a fixed-site planetary occultation.
type PlanetOccultationDiagramResult struct {
	Occultation PlanetOccultationInfo
	Frames      []PlanetOccultationDiagramFrame
	// StepDays 是实际采用的基础轨迹采样步长，单位为日。
	// StepDays is the effective base-track sampling step in days.
	StepDays float64
}

type planetOccultationDiagramTime struct {
	jde    float64
	labels []string
}

// PlanetOccultationDiagram 为已求解的固定地点行星月掩计算以月心为原点的切平面轨迹。事件数据无效或不完整时，结果不含帧。
// PlanetOccultationDiagram computes a Moon-centered tangent-plane track for an already solved fixed-site planetary occultation. Invalid or incomplete event data produces a result without frames.
func PlanetOccultationDiagram(
	info PlanetOccultationInfo,
	options PlanetOccultationDiagramOptions,
) PlanetOccultationDiagramResult {
	options = normalizePlanetOccultationDiagramOptions(options)
	result := PlanetOccultationDiagramResult{Occultation: info, StepDays: options.StepDays}
	if !planetOccultationDiagramInputValid(info) {
		return result
	}

	config, ok := planetOccultationConfigFor(info.Planet)
	if !ok {
		return result
	}
	startTT := occultationTimeToTT(info.ExternalImmersion)
	greatestTT := occultationTimeToTT(info.Greatest)
	endTT := occultationTimeToTT(info.ExternalEmersion)
	externalImmersionFrame, externalImmersionOK := planetOccultationDiagramFrameAt(startTT, config, info.Observer)
	greatestFrame, greatestOK := planetOccultationDiagramFrameAt(greatestTT, config, info.Observer)
	externalEmersionFrame, externalEmersionOK := planetOccultationDiagramFrameAt(endTT, config, info.Observer)
	if !externalImmersionOK || !greatestOK || !externalEmersionOK {
		return result
	}
	var (
		internalImmersionFrame, internalEmersionFrame PlanetOccultationDiagramFrame
		internalImmersionOK, internalEmersionOK       bool
	)
	if info.HasInternalContacts {
		internalImmersionFrame, internalImmersionOK = planetOccultationDiagramFrameAt(
			occultationTimeToTT(info.InternalImmersion), config, info.Observer,
		)
		internalEmersionFrame, internalEmersionOK = planetOccultationDiagramFrameAt(
			occultationTimeToTT(info.InternalEmersion), config, info.Observer,
		)
	}
	if (info.HasInternalContacts && (!internalImmersionOK || !internalEmersionOK)) ||
		!planetOccultationDiagramMatchesInfo(
			info, startTT, greatestTT, endTT,
			externalImmersionFrame, internalImmersionFrame, greatestFrame, internalEmersionFrame, externalEmersionFrame,
		) {
		return result
	}
	times, stepDays := planetOccultationDiagramTimes(info, startTT, greatestTT, endTT, options.StepDays)
	result.StepDays = stepDays
	result.Frames = make([]PlanetOccultationDiagramFrame, 0, len(times))
	for _, item := range times {
		frame, frameOK := planetOccultationDiagramFrameAt(item.jde, config, info.Observer)
		if !frameOK {
			return PlanetOccultationDiagramResult{Occultation: info, StepDays: stepDays}
		}
		frame.Labels = append([]string(nil), item.labels...)
		frame.Label = planetOccultationDiagramPrimaryLabel(item.labels)
		result.Frames = append(result.Frames, frame)
	}
	return result
}

func planetOccultationDiagramMatchesInfo(
	info PlanetOccultationInfo,
	startTT, greatestTT, endTT float64,
	externalImmersion, internalImmersion, greatest, internalEmersion, externalEmersion PlanetOccultationDiagramFrame,
) bool {
	if !finite(info.MinimumSeparationArcsec) || info.MinimumSeparationArcsec < 0 ||
		!finite(info.PositionAngleDeg) ||
		!finite(info.MoonSemidiameterArcsec) || info.MoonSemidiameterArcsec <= 0 ||
		!finite(info.PlanetSemidiameterArcsec) || info.PlanetSemidiameterArcsec <= 0 {
		return false
	}
	if math.Abs(externalImmersion.SeparationArcsec-externalImmersion.MoonRadiusArcsec-externalImmersion.PlanetRadiusArcsec) > planetOccultationDiagramGeometryArcsec ||
		math.Abs(externalEmersion.SeparationArcsec-externalEmersion.MoonRadiusArcsec-externalEmersion.PlanetRadiusArcsec) > planetOccultationDiagramGeometryArcsec {
		return false
	}
	if math.Abs(greatest.SeparationArcsec-info.MinimumSeparationArcsec) > planetOccultationDiagramGeometryArcsec ||
		math.Abs(greatest.MoonRadiusArcsec-info.MoonSemidiameterArcsec) > planetOccultationDiagramGeometryArcsec ||
		math.Abs(greatest.PlanetRadiusArcsec-info.PlanetSemidiameterArcsec) > planetOccultationDiagramGeometryArcsec ||
		math.Abs(signedAngleDifference(greatest.PositionAngleDeg, info.PositionAngleDeg)) > planetOccultationDiagramPositionDeg {
		return false
	}

	externalMetric := greatest.SeparationArcsec - greatest.MoonRadiusArcsec - greatest.PlanetRadiusArcsec
	internalMetric := greatest.SeparationArcsec - greatest.MoonRadiusArcsec + greatest.PlanetRadiusArcsec
	switch info.Type {
	case OccultationPartial:
		return externalMetric < -planetOccultationGrazingToleranceArcsec &&
			internalMetric >= -planetOccultationGrazingToleranceArcsec
	case OccultationGrazing:
		return math.Abs(externalMetric) <= planetOccultationGrazingToleranceArcsec &&
			math.Abs(startTT-greatestTT) <= planetOccultationDiagramContactTimeDays &&
			math.Abs(endTT-greatestTT) <= planetOccultationDiagramContactTimeDays
	case OccultationTotal:
		if math.Abs(internalImmersion.SeparationArcsec-internalImmersion.MoonRadiusArcsec+internalImmersion.PlanetRadiusArcsec) > planetOccultationDiagramGeometryArcsec ||
			math.Abs(internalEmersion.SeparationArcsec-internalEmersion.MoonRadiusArcsec+internalEmersion.PlanetRadiusArcsec) > planetOccultationDiagramGeometryArcsec {
			return false
		}
		return externalMetric < -planetOccultationGrazingToleranceArcsec &&
			internalMetric < -planetOccultationGrazingToleranceArcsec
	default:
		return false
	}
}

func normalizePlanetOccultationDiagramOptions(options PlanetOccultationDiagramOptions) PlanetOccultationDiagramOptions {
	if options.StepDays <= 0 || !finite(options.StepDays) {
		options.StepDays = planetOccultationDiagramDefaultStepDays
	}
	if options.StepDays < planetOccultationDiagramMinStepDays {
		options.StepDays = planetOccultationDiagramMinStepDays
	}
	return options
}

func planetOccultationDiagramInputValid(info PlanetOccultationInfo) bool {
	if info.Planet.Validate() != nil || info.Observer.Validate() != nil || !info.ContactsComplete ||
		info.ExternalImmersion.IsZero() || info.Greatest.IsZero() || info.ExternalEmersion.IsZero() ||
		info.Greatest.Before(info.ExternalImmersion) || info.ExternalEmersion.Before(info.Greatest) {
		return false
	}
	switch info.Type {
	case OccultationTotal:
		return info.HasInternalContacts && !info.InternalImmersion.IsZero() && !info.InternalEmersion.IsZero() &&
			info.InternalImmersion.After(info.ExternalImmersion) && info.InternalImmersion.Before(info.Greatest) &&
			info.InternalEmersion.After(info.Greatest) && info.InternalEmersion.Before(info.ExternalEmersion)
	case OccultationPartial:
		return !info.HasInternalContacts && info.InternalImmersion.IsZero() && info.InternalEmersion.IsZero()
	case OccultationGrazing:
		return !info.HasInternalContacts && info.InternalImmersion.IsZero() && info.InternalEmersion.IsZero()
	default:
		return false
	}
}

func planetOccultationDiagramTimes(
	info PlanetOccultationInfo,
	startTT, greatestTT, endTT, stepDays float64,
) ([]planetOccultationDiagramTime, float64) {
	if !finite(startTT) || !finite(greatestTT) || !finite(endTT) || greatestTT < startTT || endTT < greatestTT {
		return nil, stepDays
	}
	if endTT > startTT {
		if sampleCount := int(math.Ceil((endTT-startTT)/stepDays)) + 1; sampleCount > planetOccultationDiagramMaxSamples {
			stepDays = (endTT - startTT) / float64(planetOccultationDiagramMaxSamples-1)
		}
	}
	times := []planetOccultationDiagramTime{
		{jde: startTT, labels: []string{"C1"}},
		{jde: greatestTT, labels: []string{"Greatest"}},
		{jde: endTT, labels: []string{"C4"}},
	}
	if info.HasInternalContacts {
		times = append(times,
			planetOccultationDiagramTime{jde: occultationTimeToTT(info.InternalImmersion), labels: []string{"C2"}},
			planetOccultationDiagramTime{jde: occultationTimeToTT(info.InternalEmersion), labels: []string{"C3"}},
		)
	}
	for jde := startTT + stepDays; jde < endTT; jde += stepDays {
		times = append(times, planetOccultationDiagramTime{jde: jde})
	}
	sort.SliceStable(times, func(i, j int) bool {
		if times[i].jde == times[j].jde {
			return planetOccultationDiagramLabelPriority(times[i].labels) < planetOccultationDiagramLabelPriority(times[j].labels)
		}
		return times[i].jde < times[j].jde
	})
	return uniquePlanetOccultationDiagramTimes(times), stepDays
}

func uniquePlanetOccultationDiagramTimes(times []planetOccultationDiagramTime) []planetOccultationDiagramTime {
	unique := times[:0]
	for _, item := range times {
		if !finite(item.jde) {
			continue
		}
		if len(unique) == 0 || math.Abs(item.jde-unique[len(unique)-1].jde) > planetOccultationDiagramDuplicateDays {
			item.labels = append([]string(nil), item.labels...)
			unique = append(unique, item)
			continue
		}
		unique[len(unique)-1].labels = mergeStarOccultationDiagramLabels(unique[len(unique)-1].labels, item.labels)
	}
	return unique
}

func planetOccultationDiagramPrimaryLabel(labels []string) string {
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

func planetOccultationDiagramLabelPriority(labels []string) int {
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

func planetOccultationDiagramFrameAt(
	tt float64,
	config planetOccultationConfig,
	observer Observer,
) (PlanetOccultationDiagramFrame, bool) {
	state := planetOccultationStateAt(tt, config, &observer, -1)
	if !state.valid {
		return PlanetOccultationDiagramFrame{}, false
	}
	positionAngle := occultationPositionAngle(
		state.position.moonRA, state.position.moonDec,
		state.position.planetRA, state.position.planetDec,
	)
	if !finite(positionAngle) {
		return PlanetOccultationDiagramFrame{}, false
	}
	angle := positionAngle * math.Pi / 180
	return PlanetOccultationDiagramFrame{
		JDE:                tt,
		PlanetXArcsec:      state.separationArcsec * math.Sin(angle),
		PlanetYArcsec:      state.separationArcsec * math.Cos(angle),
		MoonRadiusArcsec:   state.moonSemidiameter,
		PlanetRadiusArcsec: state.planetSemidiameter,
		SeparationArcsec:   state.separationArcsec,
		PositionAngleDeg:   positionAngle,
		MoonAltitudeDeg:    occultationAltitude(tt, observer, state.position.moonRA, state.position.moonDec),
		MoonAzimuthDeg:     occultationAzimuth(tt, observer, state.position.moonRA, state.position.moonDec),
		DisksOverlap:       state.externalContactMetric < -planetOccultationGrazingToleranceArcsec,
		FullyOcculted:      state.internalContactMetric < -planetOccultationGrazingToleranceArcsec,
	}, true
}
