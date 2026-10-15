package basic

import (
	"fmt"
	"math"
	"time"
)

// StarOccultationInstant 包含指定时刻的点源恒星月掩可见足迹；若接触锥在该时刻未到达可见地球，Footprint 为 nil。
// StarOccultationInstant contains the visible point-source footprint at one requested instant; Footprint is nil when no part of the contact cone reaches the visible Earth.
type StarOccultationInstant struct {
	// Time 是本次查询的时刻。
	// Time is the instant this result describes.
	Time time.Time
	// TargetID 是查询恒星的标识，与 StarCoordinate.ID 同值。
	// TargetID identifies the queried star and equals StarCoordinate.ID.
	TargetID string
	// DeltaTSeconds 本次实际使用的 ΔT / ΔT actually used.
	DeltaTSeconds float64
	// SublunarLongitude 与 SublunarLatitude 是该时刻月下点，用于声明地平闭合弧 / sublunar point for the horizon closure.
	SublunarLongitude float64
	SublunarLatitude  float64
	Footprint         *OccultationFootprint
}

// PlanetOccultationInstant 包含指定时刻的行星外接触和内接触可见足迹；相应接触锥未到达可见地球时，Partial 或 Total 为 nil。
// PlanetOccultationInstant contains the visible outer- and inner-contact footprints at one requested instant; Partial or Total is nil when that contact cone does not reach the visible Earth.
type PlanetOccultationInstant struct {
	// Time 是本次查询的时刻。
	// Time is the instant this result describes.
	Time time.Time
	// Planet 是本次查询的行星。
	// Planet is the queried planet.
	Planet OccultationPlanet
	// TargetID 是行星标识，与 Planet.String() 同值。
	// TargetID identifies the planet and equals Planet.String().
	TargetID string
	// DeltaTSeconds 本次实际使用的 ΔT / ΔT actually used.
	DeltaTSeconds float64
	// SublunarLongitude 与 SublunarLatitude 是该时刻月下点，用于声明地平闭合弧 / sublunar point for the horizon closure.
	SublunarLongitude float64
	SublunarLatitude  float64
	Partial           *PlanetOccultationFootprint
	Total             *PlanetOccultationFootprint
}

// StarOccultationFootprintAtWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func StarOccultationFootprintAtWithInstance(at time.Time, star StarCoordinate, inst *Instance) (StarOccultationInstant, error) {
	result := StarOccultationInstant{Time: at, TargetID: star.ID}
	if at.IsZero() {
		return result, fmt.Errorf("%w: time is required", ErrInvalidOccultationInput)
	}
	if err := star.Validate(); err != nil {
		return result, err
	}

	cache := newStarOccultationEventCache(star, inst)
	tt := occultationTimeToTT(at)
	result.DeltaTSeconds = DeltaT(tt, true)
	result.SublunarLongitude, result.SublunarLatitude = occultationSublunarPoint(tt, cache.frameAt)
	result.Footprint = occultationInstantFootprint(
		at, cache.frameAt, cache.riseSetContextAt, false,
	)
	return result, nil
}

// PlanetOccultationFootprintsAtWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func PlanetOccultationFootprintsAtWithInstance(at time.Time, planet OccultationPlanet, inst *Instance) (PlanetOccultationInstant, error) {
	result := PlanetOccultationInstant{Time: at, Planet: planet, TargetID: planet.String()}
	if at.IsZero() {
		return result, fmt.Errorf("%w: time is required", ErrInvalidOccultationInput)
	}
	if err := planet.Validate(); err != nil {
		return result, err
	}

	config, _ := planetOccultationConfigFor(planet, inst)
	cache := newPlanetOccultationEventCache(config, inst)
	tt := occultationTimeToTT(at)
	result.DeltaTSeconds = DeltaT(tt, true)
	result.SublunarLongitude, result.SublunarLatitude = occultationSublunarPoint(tt, cache.outerFrameAt)
	result.Partial = occultationInstantFootprint(
		at, cache.outerFrameAt, cache.riseSetContextAt, false,
	)
	result.Total = occultationInstantFootprint(
		at, cache.totalFrameAt, cache.riseSetContextAt, true,
	)
	return result, nil
}

func occultationSublunarPoint(tt float64, frameAt occultationPathFrameFunc) (float64, float64) {
	frame, ok := frameAt(tt)
	if !ok {
		return math.NaN(), math.NaN()
	}
	distance := math.Sqrt(frame.moon.x*frame.moon.x + frame.moon.y*frame.moon.y + frame.moon.z*frame.moon.z)
	if distance == 0 {
		return math.NaN(), math.NaN()
	}
	rightAscension := math.Atan2(frame.moon.y, frame.moon.x) * 180 / math.Pi
	declination := math.Asin(math.Max(-1, math.Min(1, frame.moon.z/distance))) * 180 / math.Pi
	longitude := rightAscension - ApparentSiderealTime(TT2UT1(tt))*15
	for longitude > 180 {
		longitude -= 360
	}
	for longitude < -180 {
		longitude += 360
	}
	return longitude, declination
}

func occultationInstantFootprint(
	at time.Time,
	frameAt occultationPathFrameFunc,
	contextAt occultationRiseSetContextFunc,
	total bool,
) *OccultationFootprint {
	tt := occultationTimeToTT(at)
	footprint, ok := planetOccultationFootprintAtWithResolution(
		tt, frameAt, at.Location(),
		planetOccultationTimelineBoundaryPoints,
		planetOccultationTimelineHorizonPoints,
		planetOccultationTimelineTargetSpacingKM,
	)
	if !ok {
		return nil
	}

	// A one-element correction deliberately treats this instant as both ends of
	// the sequence, so every contact-arc sample receives the exact station solve.
	corrected := occultationStationCorrectFootprintEdges(
		[]PlanetOccultationFootprint{footprint}, frameAt, contextAt, total, at.Location(),
	)
	if len(corrected) != 1 {
		return nil
	}
	footprint = corrected[0]
	normalizeOccultationInstantTime(&footprint, at)
	return &footprint
}

func normalizeOccultationInstantTime(footprint *OccultationFootprint, at time.Time) {
	if footprint == nil {
		return
	}
	footprint.Time = at
	for _, polygons := range [][][]OccultationPathPoint{
		footprint.Polygons,
		footprint.InteriorPolygons,
		footprint.Boundaries,
	} {
		for polygonIndex := range polygons {
			for pointIndex := range polygons[polygonIndex] {
				polygons[polygonIndex][pointIndex].Time = at
			}
		}
	}
}
