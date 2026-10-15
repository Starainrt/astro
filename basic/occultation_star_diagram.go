package basic

import (
	"math"
	"sort"
)

const (
	starOccultationDiagramDefaultStepDays = 2.0 / 1440.0
	starOccultationDiagramMinStepDays     = 1.0 / 86400.0
	starOccultationDiagramMaxSamples      = 2000
	starOccultationDiagramDuplicateDays   = 1e-10
	starOccultationDiagramGeometryArcsec  = 0.05
	starOccultationDiagramPositionDeg     = 0.01
)

// StarOccultationDiagramOptions 控制本地恒星月掩图的轨迹采样。
// StarOccultationDiagramOptions controls local stellar-occultation diagram sampling.
type StarOccultationDiagramOptions struct {
	// StepDays 是请求的轨迹采样步长，单位为日；非正值或非有限值使用两分钟，正值小于一秒时使用一秒。长事件可能增大实际步长，使基础轨迹不超过 2000 个采样点；必要阶段帧仍会额外保留。结果会报告实际采用的值。
	// StepDays is the requested track sampling step in days. Non-positive or non-finite values use two minutes, and positive values below one second use one second. Long events may increase the effective step to keep the base track within 2000 samples; required phase frames are retained in addition. The result reports the effective value.
	StepDays float64
}

// StarOccultationDiagramFrame 描述一个时刻的站心月球与恒星几何。
// StarOccultationDiagramFrame describes topocentric Moon-star geometry at one instant.
type StarOccultationDiagramFrame struct {
	// JDE 是 TT 儒略历书日。
	// JDE is the TT Julian ephemeris day.
	JDE float64
	// StarXArcsec 和 StarYArcsec 是相对月心的切平面偏移，单位为角秒。X 向东为正，Y 向北为正。
	// StarXArcsec and StarYArcsec are tangent-plane offsets from the lunar center. X is positive east and Y is positive north.
	StarXArcsec float64
	StarYArcsec float64
	// MoonRadiusArcsec 是站心月球视半径，单位为角秒。
	// MoonRadiusArcsec is the topocentric apparent lunar semidiameter.
	MoonRadiusArcsec float64
	// SeparationArcsec 和 PositionAngleDeg 描述恒星相对月心的位置。
	// SeparationArcsec and PositionAngleDeg describe the star relative to the lunar center.
	SeparationArcsec float64
	PositionAngleDeg float64
	// MoonAltitudeDeg 和 MoonAzimuthDeg 是站心地平坐标。
	// MoonAltitudeDeg and MoonAzimuthDeg are topocentric horizontal coordinates.
	MoonAltitudeDeg float64
	MoonAzimuthDeg  float64
	// BehindMoon 表示点光源恒星位于月缘内侧。
	// BehindMoon is true while the point-source star lies strictly inside the lunar limb.
	BehindMoon bool
	// Label 是主阶段标识；Labels 在掠掩事件中保留重合阶段。
	// Label is the primary key phase; Labels retains coincident phases for grazing events.
	Label  string
	Labels []string
}

// StarOccultationDiagramResult 包含固定地点恒星月掩的几何数据。
// StarOccultationDiagramResult contains geometry for a fixed-site stellar occultation.
type StarOccultationDiagramResult struct {
	Occultation StarOccultationInfo
	Frames      []StarOccultationDiagramFrame
	// StepDays 是实际采用的基础轨迹采样步长，单位为日。
	// StepDays is the effective base-track sampling step in days.
	StepDays float64
}

type starOccultationDiagramTime struct {
	jde    float64
	labels []string
}

// StarOccultationDiagramWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func StarOccultationDiagramWithInstance(
	info StarOccultationInfo,
	star StarCoordinate,
	options StarOccultationDiagramOptions, inst *Instance) StarOccultationDiagramResult {
	options = normalizeStarOccultationDiagramOptions(options)
	result := StarOccultationDiagramResult{Occultation: info, StepDays: options.StepDays}
	if star.Validate() != nil || info.Observer.Validate() != nil ||
		!info.ContactsComplete || info.Immersion.IsZero() || info.Greatest.IsZero() || info.Emersion.IsZero() ||
		info.Greatest.Before(info.Immersion) || info.Emersion.Before(info.Greatest) ||
		(info.Type != OccultationTotal && info.Type != OccultationGrazing) {
		return result
	}

	startTT := occultationTimeToTT(info.Immersion)
	greatestTT := occultationTimeToTT(info.Greatest)
	endTT := occultationTimeToTT(info.Emersion)
	immersionFrame, immersionOK := starOccultationDiagramFrameAt(startTT, star, info.Observer, inst)
	greatestFrame, greatestOK := starOccultationDiagramFrameAt(greatestTT, star, info.Observer, inst)
	emersionFrame, emersionOK := starOccultationDiagramFrameAt(endTT, star, info.Observer, inst)
	if !immersionOK || !greatestOK || !emersionOK ||
		!starOccultationDiagramMatchesInfo(info, immersionFrame, greatestFrame, emersionFrame) {
		return result
	}
	times, stepDays := starOccultationDiagramTimes(startTT, greatestTT, endTT, options.StepDays)
	result.StepDays = stepDays
	result.Frames = make([]StarOccultationDiagramFrame, 0, len(times))
	for _, item := range times {
		frame, ok := starOccultationDiagramFrameAt(item.jde, star, info.Observer, inst)
		if !ok {
			return StarOccultationDiagramResult{Occultation: info, StepDays: stepDays}
		}
		frame.Labels = append([]string(nil), item.labels...)
		frame.Label = starOccultationDiagramPrimaryLabel(item.labels)
		result.Frames = append(result.Frames, frame)
	}
	return result
}

func starOccultationDiagramMatchesInfo(
	info StarOccultationInfo,
	immersion, greatest, emersion StarOccultationDiagramFrame,
) bool {
	if !finite(info.MinimumSeparationArcsec) || !finite(info.MoonSemidiameterArcsec) ||
		!finite(info.PositionAngleDeg) {
		return false
	}
	// 已按月缘剖面收敛的接触不在圆月面上，圆月面接触判据只适用于未修正的结果。
	if !info.LunarLimbApplied && (math.Abs(immersion.SeparationArcsec-immersion.MoonRadiusArcsec) > starOccultationDiagramGeometryArcsec ||
		math.Abs(emersion.SeparationArcsec-emersion.MoonRadiusArcsec) > starOccultationDiagramGeometryArcsec) {
		return false
	}
	return math.Abs(greatest.SeparationArcsec-info.MinimumSeparationArcsec) <= starOccultationDiagramGeometryArcsec &&
		math.Abs(greatest.MoonRadiusArcsec-info.MoonSemidiameterArcsec) <= starOccultationDiagramGeometryArcsec &&
		math.Abs(signedAngleDifference(greatest.PositionAngleDeg, info.PositionAngleDeg)) <= starOccultationDiagramPositionDeg
}

func normalizeStarOccultationDiagramOptions(options StarOccultationDiagramOptions) StarOccultationDiagramOptions {
	if options.StepDays <= 0 || !finite(options.StepDays) {
		options.StepDays = starOccultationDiagramDefaultStepDays
	}
	if options.StepDays < starOccultationDiagramMinStepDays {
		options.StepDays = starOccultationDiagramMinStepDays
	}
	return options
}

func starOccultationDiagramTimes(startTT, greatestTT, endTT, stepDays float64) ([]starOccultationDiagramTime, float64) {
	if !finite(startTT) || !finite(greatestTT) || !finite(endTT) || greatestTT < startTT || endTT < greatestTT {
		return nil, stepDays
	}
	if endTT > startTT {
		if sampleCount := int(math.Ceil((endTT-startTT)/stepDays)) + 1; sampleCount > starOccultationDiagramMaxSamples {
			stepDays = (endTT - startTT) / float64(starOccultationDiagramMaxSamples-1)
		}
	}
	times := []starOccultationDiagramTime{
		{jde: startTT, labels: []string{"Immersion"}},
		{jde: greatestTT, labels: []string{"Greatest"}},
		{jde: endTT, labels: []string{"Emersion"}},
	}
	for jde := startTT + stepDays; jde < endTT; jde += stepDays {
		times = append(times, starOccultationDiagramTime{jde: jde})
	}
	sort.SliceStable(times, func(i, j int) bool {
		if times[i].jde == times[j].jde {
			return starOccultationDiagramLabelPriority(times[i].labels) < starOccultationDiagramLabelPriority(times[j].labels)
		}
		return times[i].jde < times[j].jde
	})
	return uniqueStarOccultationDiagramTimes(times), stepDays
}

func uniqueStarOccultationDiagramTimes(times []starOccultationDiagramTime) []starOccultationDiagramTime {
	unique := times[:0]
	for _, item := range times {
		if !finite(item.jde) {
			continue
		}
		if len(unique) == 0 || math.Abs(item.jde-unique[len(unique)-1].jde) > starOccultationDiagramDuplicateDays {
			item.labels = append([]string(nil), item.labels...)
			unique = append(unique, item)
			continue
		}
		unique[len(unique)-1].labels = mergeStarOccultationDiagramLabels(unique[len(unique)-1].labels, item.labels)
	}
	return unique
}

func mergeStarOccultationDiagramLabels(existing, incoming []string) []string {
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

func starOccultationDiagramPrimaryLabel(labels []string) string {
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

func starOccultationDiagramLabelPriority(labels []string) int {
	if len(labels) == 0 {
		return 99
	}
	switch labels[0] {
	case "Immersion":
		return 0
	case "Greatest":
		return 1
	case "Emersion":
		return 2
	default:
		return 99
	}
}

func starOccultationDiagramFrameAt(tt float64, star StarCoordinate, observer Observer, instances ...*Instance) (StarOccultationDiagramFrame, bool) {
	inst := instanceValue(instances)

	position := starMoonPositionAt(tt, star, observer, inst)
	moonRadius := moonTopocentricSemidiameterN(tt, observer, -1, inst)
	if !position.valid || !finite(moonRadius) || moonRadius <= 0 {
		return StarOccultationDiagramFrame{}, false
	}
	separation := angularSeparationDegrees(position.moonRA, position.moonDec, position.starRA, position.starDec) * 3600
	positionAngle := occultationPositionAngle(position.moonRA, position.moonDec, position.starRA, position.starDec)
	if !finite(separation) || !finite(positionAngle) {
		return StarOccultationDiagramFrame{}, false
	}
	angle := positionAngle * math.Pi / 180
	return StarOccultationDiagramFrame{
		JDE:              tt,
		StarXArcsec:      separation * math.Sin(angle),
		StarYArcsec:      separation * math.Cos(angle),
		MoonRadiusArcsec: moonRadius,
		SeparationArcsec: separation,
		PositionAngleDeg: positionAngle,
		MoonAltitudeDeg:  occultationAltitude(tt, observer, position.moonRA, position.moonDec),
		MoonAzimuthDeg:   occultationAzimuth(tt, observer, position.moonRA, position.moonDec),
		BehindMoon:       separation < moonRadius-starOccultationGrazingTolerance,
	}, true
}
