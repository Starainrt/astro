package geojson

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon"
)

const lunarOccultationEvent = "lunar-occultation"

// MarshalStarOccultation 将月掩恒星的全球掩带和中心线编码为 GeoJSON。
// MarshalStarOccultation encodes a global stellar occultation band and center line as GeoJSON.
func MarshalStarOccultation(path moon.StarOccultationPath) ([]byte, error) {
	return marshalStarOccultation(path, nil)
}

// MarshalStarOccultationWithTimeMarkers 编码恒星月掩，并沿中心线按固定间隔追加 Point 要素。
// MarshalStarOccultationWithTimeMarkers encodes a stellar occultation and adds Point Features at regular intervals along its center line.
func MarshalStarOccultationWithTimeMarkers(
	path moon.StarOccultationPath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return marshalStarOccultation(path, &options)
}

func marshalStarOccultation(path moon.StarOccultationPath, markerOptions *TimeMarkerOptions) ([]byte, error) {
	if markerOptions != nil {
		if err := validateTimeMarkerOptions(*markerOptions); err != nil {
			return nil, err
		}
	}
	if err := validateStarOccultationPathData(path); err != nil {
		return nil, err
	}
	properties := map[string]interface{}{
		"target_type":       "star",
		"target_id":         path.TargetID,
		"complete":          path.Complete,
		"step_seconds":      path.Step.Seconds(),
		"target_spacing_km": path.TargetSpacingKM,
	}
	band, err := occultationBandPolygon(path.NorthernLimit, path.SouthernLimit)
	if err != nil {
		return nil, fmt.Errorf("geojson: stellar occultation band: %w", err)
	}
	value, err := multiPolygonGeometry([][]geodata.GeoPoint{band})
	if err != nil {
		return nil, fmt.Errorf("geojson: stellar occultation band: %w", err)
	}
	features := []feature{
		newFeature(lunarOccultationEvent, "occultation-band", value, cloneProperties(properties)),
	}
	if len(path.CenterLine) > 0 {
		features, err = appendOccultationPathLine(features, "center-line", path.CenterLine, properties)
		if err != nil {
			return nil, err
		}
	}
	features, err = appendOccultationPathLine(features, "north-limit", path.NorthernLimit, properties)
	if err != nil {
		return nil, err
	}
	features, err = appendOccultationPathLine(features, "south-limit", path.SouthernLimit, properties)
	if err != nil {
		return nil, err
	}
	if markerOptions != nil && len(path.CenterLine) > 0 {
		features, err = appendTimeMarkerFeatures(
			features,
			lunarOccultationEvent,
			"center-line",
			occultationPathSamples(path.CenterLine),
			*markerOptions,
		)
		if err != nil {
			return nil, err
		}
	}
	for _, marker := range []struct {
		role  string
		point moon.OccultationPathPoint
	}{
		{role: "start", point: path.Start},
		{role: "greatest", point: path.Greatest},
		{role: "end", point: path.End},
	} {
		features, err = appendOccultationPoint(features, marker.role, marker.point, properties)
		if err != nil {
			return nil, err
		}
	}
	return marshalFeatureCollection(features)
}

// MarshalPlanetOccultation 将月掩行星的部分掩、全掩和中心线编码为 GeoJSON。
// MarshalPlanetOccultation encodes partial, total, and center-line planetary occultation geometry as GeoJSON.
func MarshalPlanetOccultation(path moon.PlanetOccultationPath) ([]byte, error) {
	return marshalPlanetOccultation(path, nil)
}

// MarshalPlanetOccultationWithTimeMarkers 编码行星月掩，并沿中心线按固定间隔追加 Point 要素。
// MarshalPlanetOccultationWithTimeMarkers encodes a planetary occultation and adds Point Features at regular intervals along its center line.
func MarshalPlanetOccultationWithTimeMarkers(
	path moon.PlanetOccultationPath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return marshalPlanetOccultation(path, &options)
}

func marshalPlanetOccultation(path moon.PlanetOccultationPath, markerOptions *TimeMarkerOptions) ([]byte, error) {
	if markerOptions != nil {
		if err := validateTimeMarkerOptions(*markerOptions); err != nil {
			return nil, err
		}
	}
	if err := path.Planet.Validate(); err != nil {
		return nil, fmt.Errorf("geojson: planetary occultation target: %w", err)
	}
	if err := validatePlanetOccultationPathData(path); err != nil {
		return nil, err
	}
	properties := map[string]interface{}{
		"target_type":             "planet",
		"target_id":               path.TargetID,
		"planet":                  string(path.Planet),
		"complete":                path.Complete,
		"has_total_band":          path.HasTotalBand,
		"total_complete":          path.TotalComplete,
		"step_seconds":            path.Step.Seconds(),
		"target_spacing_km":       path.TargetSpacingKM,
		"greatest_total_width_km": path.GreatestTotalWidthKM,
	}
	features := make([]feature, 0, len(path.PartialFootprints)+len(path.TotalFootprints)+12)
	var err error
	if len(path.PartialFootprints) > 0 {
		features, err = appendOccultationFootprints(
			features, "partial-footprint", path.PartialFootprints, properties,
		)
	} else {
		features, err = appendOccultationBand(
			features, "partial-band", path.NorthernLimit, path.SouthernLimit, properties,
		)
	}
	if err != nil {
		return nil, err
	}

	if path.HasTotalBand {
		if len(path.TotalFootprints) > 0 {
			features, err = appendOccultationFootprints(
				features, "total-footprint", path.TotalFootprints, properties,
			)
		} else {
			features, err = appendOccultationBand(
				features, "total-band", path.NorthernTotalLimit, path.SouthernTotalLimit, properties,
			)
		}
		if err != nil {
			return nil, err
		}
	}

	if len(path.CenterLine) > 0 {
		features, err = appendOccultationPathLine(features, "center-line", path.CenterLine, properties)
		if err != nil {
			return nil, err
		}
	}
	features, err = appendOccultationPathLine(features, "north-limit", path.NorthernLimit, properties)
	if err != nil {
		return nil, err
	}
	features, err = appendOccultationPathLine(features, "south-limit", path.SouthernLimit, properties)
	if err != nil {
		return nil, err
	}
	if path.HasTotalBand {
		features, err = appendOccultationPathLine(
			features, "north-total-limit", path.NorthernTotalLimit, properties,
		)
		if err != nil {
			return nil, err
		}
		features, err = appendOccultationPathLine(
			features, "south-total-limit", path.SouthernTotalLimit, properties,
		)
		if err != nil {
			return nil, err
		}
	}
	if markerOptions != nil && len(path.CenterLine) > 0 {
		features, err = appendTimeMarkerFeatures(
			features,
			lunarOccultationEvent,
			"center-line",
			occultationPathSamples(path.CenterLine),
			*markerOptions,
		)
		if err != nil {
			return nil, err
		}
	}

	markers := []struct {
		role  string
		point moon.OccultationPathPoint
	}{
		{role: "start", point: path.Start},
	}
	if path.HasTotalBand {
		markers = append(markers, struct {
			role  string
			point moon.OccultationPathPoint
		}{role: "total-start", point: path.TotalStart})
	}
	markers = append(markers, struct {
		role  string
		point moon.OccultationPathPoint
	}{role: "greatest", point: path.Greatest})
	if path.HasTotalBand {
		markers = append(markers, struct {
			role  string
			point moon.OccultationPathPoint
		}{role: "total-end", point: path.TotalEnd})
	}
	markers = append(markers, struct {
		role  string
		point moon.OccultationPathPoint
	}{role: "end", point: path.End})
	for _, marker := range markers {
		features, err = appendOccultationPoint(features, marker.role, marker.point, properties)
		if err != nil {
			return nil, err
		}
	}
	return marshalFeatureCollection(features)
}

func appendOccultationBand(
	features []feature,
	role string,
	northern, southern []moon.OccultationPathPoint,
	properties map[string]interface{},
) ([]feature, error) {
	band, err := occultationBandPolygon(northern, southern)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	value, err := multiPolygonGeometry([][]geodata.GeoPoint{band})
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	return append(features, newFeature(
		lunarOccultationEvent, role, value, cloneProperties(properties),
	)), nil
}

func appendOccultationFootprints(
	features []feature,
	role string,
	footprints []moon.PlanetOccultationFootprint,
	properties map[string]interface{},
) ([]feature, error) {
	appended := 0
	for _, footprint := range footprints {
		if footprint.Time.IsZero() {
			return nil, fmt.Errorf("geojson: %s time is required", role)
		}
		polygons := make([][]geodata.GeoPoint, 0, len(footprint.Polygons))
		for _, source := range footprint.Polygons {
			polygon := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			polygons = append(polygons, polygon)
		}
		value, err := multiPolygonGeometry(polygons)
		if err != nil {
			return nil, fmt.Errorf("geojson: %s at %s: %w", role, formatTime(footprint.Time), err)
		}
		footprintProperties := cloneProperties(properties)
		footprintProperties["time"] = formatTime(footprint.Time)
		features = append(features, newFeature(
			lunarOccultationEvent, role, value, footprintProperties,
		))
		appended++
	}
	if appended == 0 {
		return nil, fmt.Errorf("geojson: %s has no valid polygons", role)
	}
	return features, nil
}

func occultationBandPolygon(
	northern, southern []moon.OccultationPathPoint,
) ([]geodata.GeoPoint, error) {
	if len(northern) != len(southern) {
		return nil, fmt.Errorf("paired limits must have the same sample count")
	}
	count := len(northern)
	if count < 2 {
		return nil, fmt.Errorf("paired limits require at least two points per side")
	}
	for index := range northern {
		if northern[index].Time.IsZero() || southern[index].Time.IsZero() {
			return nil, fmt.Errorf("paired limit sample %d time is required", index)
		}
		if !northern[index].Time.Equal(southern[index].Time) {
			return nil, fmt.Errorf("paired limit sample %d times must match", index)
		}
	}
	polygon := make([]geodata.GeoPoint, 0, 2*count)
	for _, point := range northern[:count] {
		polygon = append(polygon, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	for index := count - 1; index >= 0; index-- {
		point := southern[index]
		polygon = append(polygon, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	return polygon, nil
}

func appendOccultationPathLine(
	features []feature,
	role string,
	points []moon.OccultationPathPoint,
	properties map[string]interface{},
) ([]feature, error) {
	samples := make([]pathSample, len(points))
	for index, point := range points {
		samples[index] = occultationPathSample(point)
	}
	return appendTimedLineFeature(features, lunarOccultationEvent, role, samples, properties)
}

func occultationPathSamples(points []moon.OccultationPathPoint) []pathSample {
	samples := make([]pathSample, len(points))
	for index, point := range points {
		samples[index] = occultationPathSample(point)
	}
	return samples
}

func appendOccultationPoint(
	features []feature,
	role string,
	point moon.OccultationPathPoint,
	properties map[string]interface{},
) ([]feature, error) {
	pointProperties := cloneProperties(properties)
	pointProperties["moon_altitude_deg"] = point.MoonAltitude
	pointProperties["width_km"] = point.WidthKM
	return appendPointFeature(
		features, lunarOccultationEvent, role, occultationPathSample(point), pointProperties,
	)
}

func occultationPathSample(point moon.OccultationPathPoint) pathSample {
	return pathSample{Time: point.Time, Longitude: point.Longitude, Latitude: point.Latitude}
}

func validateStarOccultationPathData(path moon.StarOccultationPath) error {
	if !path.Complete {
		return fmt.Errorf("geojson: occultation path is incomplete")
	}
	if err := (moon.OccultationPathOptions{Step: path.Step, TargetSpacingKM: path.TargetSpacingKM}).Validate(); err != nil {
		return fmt.Errorf("geojson: invalid occultation path sampling metadata: %w", err)
	}
	if err := validateOccultationPathPoint("start", path.Start); err != nil {
		return err
	}
	if err := validateOccultationPathPoint("greatest", path.Greatest); err != nil {
		return err
	}
	if err := validateOccultationPathPoint("end", path.End); err != nil {
		return err
	}
	if path.Greatest.Time.Before(path.Start.Time) || path.End.Time.Before(path.Greatest.Time) {
		return fmt.Errorf("geojson: occultation times must be ordered start, greatest, end")
	}
	if err := validateOccultationPathSeries("center line", path.CenterLine, false); err != nil {
		return err
	}
	if err := validateOccultationPathSeries("northern limit", path.NorthernLimit, true); err != nil {
		return err
	}
	if err := validateOccultationPathSeries("southern limit", path.SouthernLimit, true); err != nil {
		return err
	}
	if len(path.NorthernLimit) != len(path.SouthernLimit) {
		return fmt.Errorf("geojson: occultation northern and southern limits must have the same sample count")
	}
	for index := range path.NorthernLimit {
		if !path.NorthernLimit[index].Time.Equal(path.SouthernLimit[index].Time) {
			return fmt.Errorf("geojson: occultation limit sample %d times must match", index)
		}
	}
	last := len(path.NorthernLimit) - 1
	if !path.NorthernLimit[0].Time.Equal(path.Start.Time) || !path.SouthernLimit[0].Time.Equal(path.Start.Time) ||
		!path.NorthernLimit[last].Time.Equal(path.End.Time) || !path.SouthernLimit[last].Time.Equal(path.End.Time) {
		return fmt.Errorf("geojson: occultation limits must span start through end")
	}
	if len(path.CenterLine) > 0 {
		if path.CenterLine[0].Time.Before(path.Start.Time) ||
			path.CenterLine[len(path.CenterLine)-1].Time.After(path.End.Time) {
			return fmt.Errorf("geojson: occultation center line must be inside start and end")
		}
		if path.Greatest.Time.Before(path.CenterLine[0].Time) ||
			path.Greatest.Time.After(path.CenterLine[len(path.CenterLine)-1].Time) {
			return fmt.Errorf("geojson: occultation greatest time is outside the center-line interval")
		}
	}
	return nil
}

func validatePlanetOccultationPathData(path moon.PlanetOccultationPath) error {
	starPath := moon.StarOccultationPath{
		TargetID: path.TargetID, Start: path.Start, Greatest: path.Greatest, End: path.End,
		Complete: path.Complete, CenterLine: path.CenterLine,
		NorthernLimit: path.NorthernLimit, SouthernLimit: path.SouthernLimit,
		Step: path.Step, TargetSpacingKM: path.TargetSpacingKM,
	}
	if err := validateStarOccultationPathData(starPath); err != nil {
		return err
	}
	if !path.HasTotalBand {
		if path.TotalComplete || !path.TotalStart.Time.IsZero() || !path.TotalEnd.Time.IsZero() ||
			len(path.NorthernTotalLimit) != 0 || len(path.SouthernTotalLimit) != 0 ||
			len(path.TotalFootprints) != 0 || path.GreatestTotalWidthKM != 0 {
			return fmt.Errorf("geojson: total-band fields require HasTotalBand")
		}
		return validateOccultationFootprints("partial", path.PartialFootprints, path.Start.Time, path.End.Time)
	}
	if !path.TotalComplete {
		return fmt.Errorf("geojson: total-occultation band is incomplete")
	}
	if err := validateOccultationPathPoint("total start", path.TotalStart); err != nil {
		return err
	}
	if err := validateOccultationPathPoint("total end", path.TotalEnd); err != nil {
		return err
	}
	if !path.Start.Time.Before(path.TotalStart.Time) || !path.TotalStart.Time.Before(path.Greatest.Time) ||
		!path.Greatest.Time.Before(path.TotalEnd.Time) || !path.TotalEnd.Time.Before(path.End.Time) {
		return fmt.Errorf("geojson: total-band times must be inside outer start, greatest, and end")
	}
	if !finiteGeoJSON(path.GreatestTotalWidthKM) || path.GreatestTotalWidthKM <= 0 ||
		!finiteGeoJSON(path.Greatest.WidthKM) || path.GreatestTotalWidthKM >= path.Greatest.WidthKM {
		return fmt.Errorf("geojson: total-band width must be positive and narrower than the outer band")
	}
	if err := validateOccultationPathSeries("northern total limit", path.NorthernTotalLimit, true); err != nil {
		return err
	}
	if err := validateOccultationPathSeries("southern total limit", path.SouthernTotalLimit, true); err != nil {
		return err
	}
	if len(path.NorthernTotalLimit) != len(path.SouthernTotalLimit) {
		return fmt.Errorf("geojson: total northern and southern limits must have the same sample count")
	}
	for index := range path.NorthernTotalLimit {
		if !path.NorthernTotalLimit[index].Time.Equal(path.SouthernTotalLimit[index].Time) {
			return fmt.Errorf("geojson: total limit sample %d times must match", index)
		}
	}
	last := len(path.NorthernTotalLimit) - 1
	if !path.NorthernTotalLimit[0].Time.Equal(path.TotalStart.Time) ||
		!path.SouthernTotalLimit[0].Time.Equal(path.TotalStart.Time) ||
		!path.NorthernTotalLimit[last].Time.Equal(path.TotalEnd.Time) ||
		!path.SouthernTotalLimit[last].Time.Equal(path.TotalEnd.Time) {
		return fmt.Errorf("geojson: total limits must span total start through total end")
	}
	if err := validateOccultationFootprints("partial", path.PartialFootprints, path.Start.Time, path.End.Time); err != nil {
		return err
	}
	return validateOccultationFootprints("total", path.TotalFootprints, path.TotalStart.Time, path.TotalEnd.Time)
}

func validateOccultationPathSeries(name string, points []moon.OccultationPathPoint, required bool) error {
	if required && len(points) < 2 {
		return fmt.Errorf("geojson: %s requires at least two points", name)
	}
	previous := time.Time{}
	for index, point := range points {
		if err := validateOccultationPathPoint(fmt.Sprintf("%s[%d]", name, index), point); err != nil {
			return err
		}
		if !previous.IsZero() && !point.Time.After(previous) {
			return fmt.Errorf("geojson: %s times must be strictly increasing", name)
		}
		previous = point.Time
	}
	return nil
}

func validateOccultationPathPoint(name string, point moon.OccultationPathPoint) error {
	if point.Time.IsZero() {
		return fmt.Errorf("geojson: %s time is required", name)
	}
	if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
		return fmt.Errorf("geojson: %s: %w", name, err)
	}
	if !finiteGeoJSON(point.MoonAltitude) || point.MoonAltitude < -90 || point.MoonAltitude > 90 {
		return fmt.Errorf("geojson: %s Moon altitude must be finite and within [-90, 90]", name)
	}
	if !finiteGeoJSON(point.WidthKM) || point.WidthKM < 0 {
		return fmt.Errorf("geojson: %s width must be finite and non-negative", name)
	}
	return nil
}

func validateOccultationFootprints(
	name string,
	footprints []moon.PlanetOccultationFootprint,
	start, end time.Time,
) error {
	previous := time.Time{}
	for index, footprint := range footprints {
		if footprint.Time.IsZero() {
			return fmt.Errorf("geojson: %s footprint[%d] time is required", name, index)
		}
		if footprint.Time.Before(start) || footprint.Time.After(end) {
			return fmt.Errorf("geojson: %s footprint[%d] time must be inside its contact interval", name, index)
		}
		if !previous.IsZero() && !footprint.Time.After(previous) {
			return fmt.Errorf("geojson: %s footprint times must be strictly increasing", name)
		}
		previous = footprint.Time
	}
	return nil
}
