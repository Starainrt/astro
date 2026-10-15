package geojson

import (
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
)

const planetTransitEvent = "planet-transit"

// 标准大气，与 basic 站心凌日的视地平判定及凌日 SVG 一致。
const (
	planetTransitRefractionPressureHPa  = 1010.0
	planetTransitRefractionTemperatureC = 10.0
	minimumPlanetTransitIsochroneStep   = time.Second
)

// PlanetTransitOptions 是凌日 GeoJSON 导出选项；所有要素都被丢弃时导出失败。
// PlanetTransitOptions are the planetary-transit GeoJSON export options; the export fails when every
// Feature is dropped.
type PlanetTransitOptions struct {
	// BoundaryPoints 是地平线采样点数，同时决定分区的经度列数 max(360, BoundaryPoints)（上限 720）；
	// 小于等于零时使用 360，其他值限制在 [12, 1440]。
	// BoundaryPoints is the horizon sample count and also sets the region longitude columns to
	// max(360, BoundaryPoints), capped at 720. Values <= 0 use 360; others are clamped to [12, 1440].
	BoundaryPoints int
	// IsochroneStep 为正时输出一触到四触的站心接触等时线（掠凌没有二触、三触），至少一秒；零值不输出。
	// A positive IsochroneStep adds topocentric C1-C4 isochrones (no C2/C3 for a grazing transit) and must be
	// at least one second. The zero value adds none.
	IsochroneStep time.Duration
	// TimeMarkers 非空时沿一触到四触的太阳直射点轨迹追加时间标记 Point 要素。
	// TimeMarkers adds time-marker Point Features along the subsolar track from C1 through C4 when non-nil.
	TimeMarkers *TimeMarkerOptions
	// SkipRoles 列出不写进输出的 role；分区三个 role 都跳过时不做分区计算。
	// SkipRoles lists roles to leave out; skipping all three region roles skips the region computation.
	SkipRoles []string
}

// planetTransitInput 是水星、金星两种 TransitInfo 的公共视图。
type planetTransitInput struct {
	planet      string
	planetIndex int
	start       time.Time
	c2          time.Time
	greatest    time.Time
	c3          time.Time
	end         time.Time
	hasInternal bool
	separation  float64
	sunRadius   float64
	planetRad   float64
}

// MarshalMercuryTransit 将水星凌日的全球可见分区、一触/四触地平线与可选站心等时线编码为 GeoJSON / encode a Mercury transit as GeoJSON.
func (module *Generator) MarshalMercuryTransit(info mercury.TransitInfo, options PlanetTransitOptions) ([]byte, error) {
	if !info.Valid {
		return nil, fmt.Errorf("geojson: mercury transit is invalid")
	}
	return marshalPlanetTransit(planetTransitInput{
		planet: "mercury", planetIndex: 1,
		start: info.Start, c2: info.InternalStart, greatest: info.Greatest, c3: info.InternalEnd, end: info.End,
		hasInternal: info.HasInternal, separation: info.MinimumSeparationArcsec,
		sunRadius: info.SunSemidiameterArcsec, planetRad: info.PlanetSemidiameterArcsec,
	}, options, module)
}

// MarshalVenusTransit 将金星凌日的全球可见分区、一触/四触地平线与可选站心等时线编码为 GeoJSON / encode a Venus transit as GeoJSON.
func (module *Generator) MarshalVenusTransit(info venus.TransitInfo, options PlanetTransitOptions) ([]byte, error) {
	if !info.Valid {
		return nil, fmt.Errorf("geojson: venus transit is invalid")
	}
	return marshalPlanetTransit(planetTransitInput{
		planet: "venus", planetIndex: 2,
		start: info.Start, c2: info.InternalStart, greatest: info.Greatest, c3: info.InternalEnd, end: info.End,
		hasInternal: info.HasInternal, separation: info.MinimumSeparationArcsec,
		sunRadius: info.SunSemidiameterArcsec, planetRad: info.PlanetSemidiameterArcsec,
	}, options, module)
}

func validatePlanetTransitInput(input planetTransitInput) error {
	if input.start.IsZero() || input.greatest.IsZero() || input.end.IsZero() {
		return fmt.Errorf("geojson: %s transit contact times are required", input.planet)
	}
	if !input.hasInternal && (!input.c2.IsZero() || !input.c3.IsZero()) {
		return fmt.Errorf("geojson: %s transit cannot have internal contact times without internal contacts", input.planet)
	}
	ordered := []time.Time{input.start}
	if input.hasInternal {
		if input.c2.IsZero() || input.c3.IsZero() {
			return fmt.Errorf("geojson: %s transit internal contact times are required", input.planet)
		}
		ordered = append(ordered, input.c2)
	}
	ordered = append(ordered, input.greatest)
	if input.hasInternal {
		ordered = append(ordered, input.c3)
	}
	ordered = append(ordered, input.end)
	for index := 1; index < len(ordered); index++ {
		if !ordered[index-1].Before(ordered[index]) {
			return fmt.Errorf("geojson: %s transit contact times are not strictly ordered", input.planet)
		}
	}
	return nil
}

func marshalPlanetTransit(input planetTransitInput, options PlanetTransitOptions, module *Generator) ([]byte, error) {
	markerOptions := options.TimeMarkers
	if markerOptions != nil {
		if err := validateTimeMarkerOptions(*markerOptions); err != nil {
			return nil, err
		}
	}
	if options.IsochroneStep < 0 || (options.IsochroneStep > 0 && options.IsochroneStep < minimumPlanetTransitIsochroneStep) {
		return nil, fmt.Errorf("geojson: transit isochrone step must be zero or at least %s", minimumPlanetTransitIsochroneStep)
	}
	if err := validatePlanetTransitInput(input); err != nil {
		return nil, err
	}
	timeScale, scaleErr := timeScaleForMarkers(markerOptions)
	if scaleErr != nil {
		return nil, scaleErr
	}
	// 几何一律用民用时刻，只有写进属性的时刻换时标。
	label := func(value time.Time) string {
		if value.IsZero() {
			return ""
		}
		return formatTime(basic.LabelIn(timeScale, value.UTC()))
	}
	boundaryPoints := normalizeLunarBoundaryPoints(options.BoundaryPoints)
	longitudePoints := lunarVisibilityLongitudePoints(boundaryPoints, LunarEclipseOptions{})
	longitudes := make([]float64, longitudePoints+1)
	for index := range longitudes {
		longitudes[index] = -180 + 360*float64(index)/float64(longitudePoints)
	}
	refraction := basic.RefractionFromApparentAltitude(0, planetTransitRefractionPressureHPa, planetTransitRefractionTemperatureC)
	base := map[string]interface{}{"planet": input.planet, "has_internal": input.hasInternal}
	withBase := func(extra map[string]interface{}) map[string]interface{} {
		properties := cloneProperties(base)
		for key, value := range extra {
			properties[key] = value
		}
		return properties
	}
	features := make([]feature, 0, 12)

	columnsAt := func(at time.Time) [][]lunarLatitudeInterval {
		subsolar := solarSubsolarPoint(at, module)
		columns := make([][]lunarLatitudeInterval, len(longitudes))
		for index, longitude := range longitudes {
			columns[index] = planetTransitSunUpIntervals(subsolar, longitude, refraction)
		}
		return columns
	}
	wantEntire := !skippedLunarRole(options.SkipRoles, "transit-entire")
	wantSunrise := !skippedLunarRole(options.SkipRoles, "transit-at-sunrise")
	wantSunset := !skippedLunarRole(options.SkipRoles, "transit-at-sunset")
	wantRegions := wantEntire || wantSunrise || wantSunset
	needC1 := wantRegions || !skippedLunarRole(options.SkipRoles, "visible-at-c1")
	needGreatest := wantRegions
	needC4 := wantRegions || !skippedLunarRole(options.SkipRoles, "visible-at-c4")
	var c1Columns, greatestColumns, c4Columns [][]lunarLatitudeInterval
	if needC1 {
		c1Columns = columnsAt(input.start)
	}
	if needGreatest {
		greatestColumns = columnsAt(input.greatest)
	}
	if needC4 {
		c4Columns = columnsAt(input.end)
	}

	for _, contact := range []struct {
		role, horizonRole string
		at                time.Time
		columns           [][]lunarLatitudeInterval
	}{
		{"visible-at-c1", "c1-horizon", input.start, c1Columns},
		{"visible-at-c4", "c4-horizon", input.end, c4Columns},
	} {
		if !skippedLunarRole(options.SkipRoles, contact.role) {
			value, err := planetTransitRegionGeometry(longitudes, contact.columns)
			if err != nil {
				return nil, fmt.Errorf("geojson: %s: %w", contact.role, err)
			}
			features = append(features, newFeature(planetTransitEvent, contact.role, value,
				withBase(map[string]interface{}{"time": label(contact.at)})))
		}
		if !skippedLunarRole(options.SkipRoles, contact.horizonRole) {
			subsolar := solarSubsolarPoint(contact.at, module)
			antisolar := geodata.GeoPoint{Longitude: geodata.NormalizeLongitude(subsolar.Longitude + 180), Latitude: -subsolar.Latitude}
			value, err := geoMultiLineGeometry(geodata.SphericalCircle(antisolar, 90-refraction, boundaryPoints), true)
			if err != nil {
				return nil, fmt.Errorf("geojson: %s: %w", contact.horizonRole, err)
			}
			features = append(features, newFeature(planetTransitEvent, contact.horizonRole, value,
				withBase(map[string]interface{}{"time": label(contact.at)})))
		}
	}

	regionFeatures, err := planetTransitRegionFeatures(longitudes, c1Columns, greatestColumns, c4Columns, options.SkipRoles,
		withBase(map[string]interface{}{
			"time_start":       label(input.start),
			"time_end":         label(input.end),
			"longitude_points": longitudePoints,
		}))
	if err != nil {
		return nil, err
	}
	features = append(features, regionFeatures...)

	if options.IsochroneStep > 0 {
		isochrones, isoErr := planetTransitIsochroneFeatures(input, options.IsochroneStep, timeScale, options.SkipRoles, withBase, module)
		if isoErr != nil {
			return nil, isoErr
		}
		features = append(features, isochrones...)
	}

	greatest := solarSubsolarPoint(input.greatest, module)
	metadata := withBase(map[string]interface{}{
		"start":                      label(input.start),
		"greatest":                   label(input.greatest),
		"end":                        label(input.end),
		"minimum_separation_arcsec":  input.separation,
		"sun_semidiameter_arcsec":    input.sunRadius,
		"planet_semidiameter_arcsec": input.planetRad,
	})
	if input.hasInternal {
		metadata["internal_start"], metadata["internal_end"] = label(input.c2), label(input.c3)
	}
	features, err = appendPointFeature(features, planetTransitEvent, "greatest",
		pathSample{Time: basic.LabelIn(timeScale, input.greatest.UTC()), Longitude: greatest.Longitude, Latitude: greatest.Latitude},
		metadata)
	if err != nil {
		return nil, err
	}
	if markerOptions != nil {
		markers, markerErr := planetTransitTimeMarkerSamples(input.start, input.end, timeScale, *markerOptions, module)
		if markerErr != nil {
			return nil, markerErr
		}
		features, err = appendTimeMarkerPointFeatures(features, planetTransitEvent, "subsolar-track", markers, markerOptions.Location)
		if err != nil {
			return nil, err
		}
	}
	return marshalFeatureCollectionWithTimeScale(dropFeaturesByRole(features, options.SkipRoles), timeScale)
}

// planetTransitSunUpIntervals 返回 longitude 子午线上太阳中心在视地平上的纬度区间：
// 地点到直射点的角距小于 90°+折射即可见，即 sinδ·sinφ + cosδ·cosΔλ·cosφ > -sin(折射)。
func planetTransitSunUpIntervals(subsolar geodata.GeoPoint, longitude, refraction float64) []lunarLatitudeInterval {
	rad := math.Pi / 180
	a := math.Sin(subsolar.Latitude * rad)
	b := math.Cos(subsolar.Latitude*rad) * math.Cos((longitude-subsolar.Longitude)*rad)
	threshold := -math.Sin(refraction * rad)
	visible := func(latitude float64) bool {
		return a*math.Sin(latitude*rad)+b*math.Cos(latitude*rad) > threshold
	}
	boundaries := []float64{-90}
	if amplitude := math.Hypot(a, b); amplitude > 0 && math.Abs(threshold/amplitude) < 1 {
		// a·sinφ + b·cosφ = R·sin(φ+α)，α = atan2(b, a)。
		alpha := math.Atan2(b, a) / rad
		base := math.Asin(threshold/amplitude) / rad
		roots := make([]float64, 0, 2)
		for _, value := range []float64{base - alpha, 180 - base - alpha} {
			for _, shift := range []float64{-360, 0, 360} {
				if root := value + shift; root > -90 && root < 90 {
					roots = append(roots, root)
				}
			}
		}
		if len(roots) == 2 && roots[0] > roots[1] {
			roots[0], roots[1] = roots[1], roots[0]
		}
		boundaries = append(boundaries, roots...)
	}
	boundaries = append(boundaries, 90)
	intervals := make([]lunarLatitudeInterval, 0, 2)
	for index := 0; index+1 < len(boundaries); index++ {
		low, high := boundaries[index], boundaries[index+1]
		if high-low > 1e-9 && visible((low+high)/2) {
			intervals = append(intervals, lunarLatitudeInterval{low: low, high: high})
		}
	}
	return lunarUnionLatitudeIntervals(intervals...)
}

func planetTransitRegionGeometry(longitudes []float64, columns [][]lunarLatitudeInterval) (geometry, error) {
	value, err := lunarVisibilityEnvelopeGeometry(longitudes, columns)
	if err != nil {
		return geometry{}, err
	}
	return dropDegenerateMultiPolygonRings(value), nil
}

// planetTransitRegionFeatures 按与凌日 SVG 相同的口径分区：全程可见是三刻都在地平上，日出带是四触可见而一触不可见，
// 其余凡有一刻可见的都归入日落带（含凌甚前后可见的极区透镜与两端可见、凌甚不可见的高纬带）。
func planetTransitRegionFeatures(
	longitudes []float64,
	c1, greatest, c4 [][]lunarLatitudeInterval,
	skip []string,
	properties map[string]interface{},
) ([]feature, error) {
	wantEntire := !skippedLunarRole(skip, "transit-entire")
	wantSunrise := !skippedLunarRole(skip, "transit-at-sunrise")
	wantSunset := !skippedLunarRole(skip, "transit-at-sunset")
	if !wantEntire && !wantSunrise && !wantSunset {
		return nil, nil
	}
	union := func(groups ...[]lunarLatitudeInterval) []lunarLatitudeInterval {
		var merged []lunarLatitudeInterval
		for _, group := range groups {
			merged = append(merged, group...)
		}
		return lunarUnionLatitudeIntervals(merged...)
	}
	entire := make([][]lunarLatitudeInterval, len(longitudes))
	sunrise := make([][]lunarLatitudeInterval, len(longitudes))
	sunset := make([][]lunarLatitudeInterval, len(longitudes))
	for index := range longitudes {
		bothEnds := lunarIntersectLatitudeIntervals(c1[index], c4[index])
		entire[index] = lunarIntersectLatitudeIntervals(bothEnds, greatest[index])
		sunrise[index] = lunarSubtractLatitudeIntervals(c4[index], c1[index])
		sunset[index] = lunarSubtractLatitudeIntervals(union(c1[index], c4[index], greatest[index]),
			union(entire[index], sunrise[index]))
	}
	features := make([]feature, 0, 3)
	for _, region := range []struct {
		role    string
		want    bool
		columns [][]lunarLatitudeInterval
	}{
		{"transit-entire", wantEntire, entire},
		{"transit-at-sunrise", wantSunrise, sunrise},
		{"transit-at-sunset", wantSunset, sunset},
	} {
		if !region.want {
			continue
		}
		value, err := planetTransitRegionGeometry(longitudes, region.columns)
		if err != nil {
			return nil, fmt.Errorf("geojson: %s: %w", region.role, err)
		}
		features = append(features, newFeature(planetTransitEvent, region.role, value, cloneProperties(properties)))
	}
	return features, nil
}

// planetTransitIsochroneFeatures 把每个接触的站心等时线编码为一条带 times 的 MultiLineString：
// 每段是一条支路，times 与坐标对齐，同一支路上所有点同刻。
func planetTransitIsochroneFeatures(
	input planetTransitInput,
	step time.Duration,
	timeScale basic.TimeScale,
	skip []string,
	withBase func(map[string]interface{}) map[string]interface{},
	module *Generator,
) ([]feature, error) {
	result := basic.PlanetTransitResultFromTimes(input.planetIndex, input.start, input.c2, input.greatest, input.c3, input.end)
	features := make([]feature, 0, 4)
	for _, contact := range []struct {
		kind basic.PlanetTransitContactKind
		role string
	}{
		{basic.PlanetTransitC1, "c1-isochrone"},
		{basic.PlanetTransitC2, "c2-isochrone"},
		{basic.PlanetTransitC3, "c3-isochrone"},
		{basic.PlanetTransitC4, "c4-isochrone"},
	} {
		if skippedLunarRole(skip, contact.role) {
			continue
		}
		lines := basic.PlanetTransitContactIsochronesWithInstance(result, contact.kind, step, module.inst)
		coordinates := make([][][]float64, 0, len(lines))
		times := make([][]string, 0, len(lines))
		for _, line := range lines {
			at := basic.LabelIn(timeScale, basic.JD2DateByZone(line.JD, time.UTC, false))
			for _, segment := range line.Segments {
				if len(segment) < 2 {
					continue
				}
				points := make([]pathSample, len(segment))
				for index, point := range segment {
					points[index] = pathSample{Time: at, Longitude: point.Longitude, Latitude: point.Latitude}
				}
				pieces, err := splitTimedLine(points, false)
				if err != nil {
					return nil, fmt.Errorf("geojson: %s: %w", contact.role, err)
				}
				for _, piece := range pieces {
					if len(piece) < 2 {
						continue
					}
					coordinate := make([][]float64, len(piece))
					stamp := make([]string, len(piece))
					for index, point := range piece {
						if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
							return nil, fmt.Errorf("geojson: %s: %w", contact.role, err)
						}
						coordinate[index] = []float64{point.Longitude, point.Latitude}
						stamp[index] = formatTime(point.Time)
					}
					coordinates = append(coordinates, coordinate)
					times = append(times, stamp)
				}
			}
		}
		if len(coordinates) == 0 {
			continue
		}
		properties := withBase(map[string]interface{}{"step_seconds": step.Seconds(), "times": times})
		features = append(features, newFeature(planetTransitEvent, contact.role,
			geometry{Type: "MultiLineString", Coordinates: coordinates}, properties))
	}
	return features, nil
}

func planetTransitTimeMarkerSamples(
	start, end time.Time,
	scale basic.TimeScale,
	options TimeMarkerOptions,
	module *Generator,
) ([]pathSample, error) {
	step, err := normalizeTimeMarkerStep(options.Step)
	if err != nil {
		return nil, fmt.Errorf("geojson: transit time markers: %w", err)
	}
	location := normalizeTimeMarkerLocation(options.Location)
	capacity, err := timeMarkerCapacity(start, end, step, location)
	if err != nil {
		return nil, fmt.Errorf("geojson: transit time markers: %w", err)
	}
	markers := make([]pathSample, 0, capacity)
	for current := firstTimeMarkerAfter(start, step, location); current.Before(end); current = current.Add(step) {
		point := solarSubsolarPoint(current, module)
		markers = append(markers, pathSample{
			Time:      basic.LabelIn(scale, current),
			Longitude: point.Longitude,
			Latitude:  point.Latitude,
		})
	}
	return markers, nil
}
