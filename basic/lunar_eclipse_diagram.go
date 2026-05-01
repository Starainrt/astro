package basic

import (
	"math"
	"sort"
)

const (
	lunarEclipseDiagramDefaultStepDays = 5.0 / 1440.0
	lunarEclipseDiagramMinStepDays     = 1.0 / 86400.0
	lunarEclipseDiagramMaxSamples      = 2000
	lunarEclipseDiagramDuplicateDays   = 1e-10
)

// LunarEclipseDiagramOptions 控制月食穿影图采样。
// LunarEclipseDiagramOptions controls lunar eclipse shadow-path diagram sampling.
type LunarEclipseDiagramOptions struct {
	// StepDays 是路径采样步长，单位为日；<=0 时使用 5 分钟。
	// StepDays is the path sampling step in days; values <= 0 use five minutes.
	StepDays float64
}

// LunarEclipseDiagramPoint 表示月食穿影图上的一个月心位置。
// LunarEclipseDiagramPoint is one Moon-center point in a lunar eclipse diagram.
type LunarEclipseDiagramPoint struct {
	// JDE 是力学时儒略日, TT Julian ephemeris day.
	JDE float64
	// X / Y 是以月球半径为单位的月心相对地影中心坐标。
	// X/Y are Moon-center coordinates relative to the shadow center, in Moon-radius units.
	X float64
	Y float64
	// Label 是关键接触标签，如 P1/U1/U2/Greatest/U3/U4/P4。
	// Label is a key contact label such as P1/U1/U2/Greatest/U3/U4/P4.
	Label string
	// Labels 是该点对应的全部关键接触标签；若事件重合，这里会有多个值。
	// Labels contains all contact labels attached to this point.
	Labels []string
}

// LunarEclipseDiagramResult 表示月食穿影图几何结果。
// LunarEclipseDiagramResult is the geometry result for a lunar eclipse diagram.
type LunarEclipseDiagramResult struct {
	// Eclipse 是对应的月食结果。
	// Eclipse is the eclipse result used for the diagram.
	Eclipse LunarEclipseResult
	// MoonRadius 是月球半径，单位为图上月球半径；固定为 1。
	// MoonRadius is the Moon radius in diagram Moon-radius units; always 1.
	MoonRadius float64
	// UmbraRadius 是本影半径，单位为图上月球半径。
	// UmbraRadius is the umbral shadow radius in Moon-radius units.
	UmbraRadius float64
	// PenumbraRadius 是半影半径，单位为图上月球半径。
	// PenumbraRadius is the penumbral shadow radius in Moon-radius units.
	PenumbraRadius float64
	// Points 是月心路径点，包含接触点与采样点。
	// Points are Moon-center path points, including contact and sampled points.
	Points []LunarEclipseDiagramPoint
	// StepDays 是实际采用的路径采样步长，单位为日。
	// StepDays is the effective path sampling step in days.
	StepDays float64
}

type lunarEclipseDiagramTime struct {
	jde    float64
	labels []string
}

// LunarEclipseDiagram 计算月食穿影图几何数据，默认使用 Danjon 影半径模型。
// LunarEclipseDiagram computes lunar eclipse diagram geometry, using the Danjon shadow model by default.
func LunarEclipseDiagram(seedJDE float64, options LunarEclipseDiagramOptions) LunarEclipseDiagramResult {
	return LunarEclipseDiagramDanjon(seedJDE, options)
}

// LunarEclipseDiagramDanjon 计算月食穿影图几何数据，使用 Danjon 影半径模型。
// LunarEclipseDiagramDanjon computes lunar eclipse diagram geometry with the Danjon shadow model.
func LunarEclipseDiagramDanjon(seedJDE float64, options LunarEclipseDiagramOptions) LunarEclipseDiagramResult {
	return lunarEclipseDiagram(seedJDE, lunarEclipseShadowDanjon, options)
}

// LunarEclipseDiagramChauvenet 计算月食穿影图几何数据，使用 Chauvenet 影半径模型。
// LunarEclipseDiagramChauvenet computes lunar eclipse diagram geometry with the Chauvenet shadow model.
func LunarEclipseDiagramChauvenet(seedJDE float64, options LunarEclipseDiagramOptions) LunarEclipseDiagramResult {
	return lunarEclipseDiagram(seedJDE, lunarEclipseShadowChauvenet, options)
}

func lunarEclipseDiagram(
	seedJDE float64,
	shadowModel lunarEclipseShadowModel,
	options LunarEclipseDiagramOptions,
) LunarEclipseDiagramResult {
	options = normalizeLunarEclipseDiagramOptions(options)
	eclipse := lunarEclipse(seedJDE, shadowModel)
	result := LunarEclipseDiagramResult{
		Eclipse:  eclipse,
		StepDays: options.StepDays,
	}
	if !eclipse.HasPenumbral {
		return result
	}

	maximumState := computeLunarShadowState(eclipse.Maximum, shadowModel)
	if maximumState.moonRadiusRad <= 0 {
		return result
	}

	result.MoonRadius = 1
	result.UmbraRadius = maximumState.umbraRadiusRad / maximumState.moonRadiusRad
	result.PenumbraRadius = maximumState.penumbraRadiusRad / maximumState.moonRadiusRad

	times, stepDays := lunarEclipseDiagramTimes(eclipse, options.StepDays)
	result.StepDays = stepDays
	result.Points = make([]LunarEclipseDiagramPoint, 0, len(times))
	for _, item := range times {
		state := computeLunarShadowState(item.jde, shadowModel)
		result.Points = append(result.Points, LunarEclipseDiagramPoint{
			JDE:    item.jde,
			X:      state.x / maximumState.moonRadiusRad,
			Y:      state.y / maximumState.moonRadiusRad,
			Label:  lunarEclipseDiagramPrimaryLabel(item.labels),
			Labels: append([]string(nil), item.labels...),
		})
	}
	return result
}

func normalizeLunarEclipseDiagramOptions(options LunarEclipseDiagramOptions) LunarEclipseDiagramOptions {
	if options.StepDays <= 0 || math.IsNaN(options.StepDays) || math.IsInf(options.StepDays, 0) {
		options.StepDays = lunarEclipseDiagramDefaultStepDays
	}
	if options.StepDays < lunarEclipseDiagramMinStepDays {
		options.StepDays = lunarEclipseDiagramMinStepDays
	}
	return options
}

func lunarEclipseDiagramTimes(eclipse LunarEclipseResult, stepDays float64) ([]lunarEclipseDiagramTime, float64) {
	startJDE := eclipse.PenumbralStart
	endJDE := eclipse.PenumbralEnd
	if startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil, stepDays
	}

	if sampleCount := int(math.Ceil((endJDE-startJDE)/stepDays)) + 1; sampleCount > lunarEclipseDiagramMaxSamples {
		stepDays = (endJDE - startJDE) / float64(lunarEclipseDiagramMaxSamples-1)
	}

	times := []lunarEclipseDiagramTime{
		{jde: startJDE, labels: []string{"P1"}},
		{jde: eclipse.Maximum, labels: []string{"Greatest"}},
		{jde: endJDE, labels: []string{"P4"}},
	}
	if eclipse.HasPartial {
		times = append(times,
			lunarEclipseDiagramTime{jde: eclipse.PartialStart, labels: []string{"U1"}},
			lunarEclipseDiagramTime{jde: eclipse.PartialEnd, labels: []string{"U4"}},
		)
	}
	if eclipse.HasTotal {
		times = append(times,
			lunarEclipseDiagramTime{jde: eclipse.TotalStart, labels: []string{"U2"}},
			lunarEclipseDiagramTime{jde: eclipse.TotalEnd, labels: []string{"U3"}},
		)
	}
	for jde := startJDE + stepDays; jde < endJDE; jde += stepDays {
		times = append(times, lunarEclipseDiagramTime{jde: jde})
	}

	sort.SliceStable(times, func(i, j int) bool {
		if times[i].jde == times[j].jde {
			return lunarEclipseDiagramLabelPriority(times[i].labels) < lunarEclipseDiagramLabelPriority(times[j].labels)
		}
		return times[i].jde < times[j].jde
	})
	return uniqueLunarEclipseDiagramTimes(times), stepDays
}

func uniqueLunarEclipseDiagramTimes(times []lunarEclipseDiagramTime) []lunarEclipseDiagramTime {
	if len(times) < 2 {
		return times
	}

	unique := times[:0]
	for _, item := range times {
		if item.jde == 0 {
			continue
		}
		if len(unique) == 0 || math.Abs(item.jde-unique[len(unique)-1].jde) > lunarEclipseDiagramDuplicateDays {
			item.labels = append([]string(nil), item.labels...)
			unique = append(unique, item)
			continue
		}
		unique[len(unique)-1].labels = mergeLunarEclipseDiagramLabels(unique[len(unique)-1].labels, item.labels)
	}
	return unique
}

func mergeLunarEclipseDiagramLabels(existing, incoming []string) []string {
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

func lunarEclipseDiagramPrimaryLabel(labels []string) string {
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

func lunarEclipseDiagramLabelPriority(labels []string) int {
	if len(labels) == 0 {
		return 99
	}
	switch labels[0] {
	case "P1":
		return 0
	case "U1":
		return 1
	case "U2":
		return 2
	case "Greatest":
		return 3
	case "U3":
		return 4
	case "U4":
		return 5
	case "P4":
		return 6
	default:
		return 99
	}
}
