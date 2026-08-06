package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

type decodedCollection struct {
	Type     string           `json:"type"`
	Features []decodedFeature `json:"features"`
}

type decodedFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	} `json:"geometry"`
}

func TestMarshalSolarEclipseFeatureCollection(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step:           20 * time.Minute,
		BoundaryPoints: 36,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 5 * time.Minute})
	if !ok {
		t.Fatal("expected solar central path")
	}

	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection,
		"partial-footprint", "central-band", "center-line", "north-limit", "south-limit", "greatest")
	assertCollectionCoordinates(t, collection)
	assertClosedMultiPolygon(t, featureWithRole(t, collection, "central-band"))
	assertTimedLineAligned(t, featureWithRole(t, collection, "center-line"))
}

func TestMarshalSolarEclipseAllowsSingleLimitCentrality(t *testing.T) {
	date := time.Date(2003, time.May, 30, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 20 * time.Minute, BoundaryPoints: 24,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 10 * time.Minute})
	if !ok || central.Eclipse.Centrality != eclipse.SolarEclipseCentralOneLimit {
		t.Fatalf("expected one-limit central eclipse, got ok=%v centrality=%s", ok, central.Eclipse.Centrality)
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	if len(featuresWithRole(collection, "center-line")) != 1 || len(featuresWithRole(collection, "central-band")) != 0 {
		t.Fatal("single-limit central eclipse should export center line without a band")
	}
}

func TestMarshalSolarEclipseAllowsLowSampleOpenFootprints(t *testing.T) {
	for _, fixture := range []struct {
		date time.Time
		step time.Duration
	}{
		{time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), 5 * time.Minute},
		{time.Date(2025, time.March, 29, 0, 0, 0, 0, time.UTC), 5 * time.Minute},
	} {
		partial, ok := eclipse.SolarEclipsePartialFootprints(fixture.date, eclipse.SolarEclipsePartialFootprintOptions{
			Step: fixture.step, BoundaryPoints: 12,
		})
		if !ok {
			t.Fatalf("%s: expected solar partial footprints", fixture.date.Format("2006-01-02"))
		}
		if _, err := geojson.MarshalSolarEclipse(partial, nil); err != nil {
			t.Fatalf("%s low-sample GeoJSON: %v", fixture.date.Format("2006-01-02"), err)
		}
	}
}

func TestMarshalSolarEclipseWithTimeMarkers(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step:           20 * time.Minute,
		BoundaryPoints: 36,
	})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 5 * time.Minute})
	if !ok {
		t.Fatal("expected solar central path")
	}

	data, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, &central, geojson.TimeMarkerOptions{
		Step:     time.Hour,
		Location: time.FixedZone("CST", 8*60*60),
	})
	if err != nil {
		t.Fatalf("MarshalSolarEclipseWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	markers := featuresWithRole(collection, "time-marker")
	if len(markers) == 0 {
		t.Fatal("solar eclipse has no time markers")
	}
	for _, marker := range markers {
		if marker.Properties["source_role"] != "center-line" {
			t.Fatalf("time marker source_role=%v, want center-line", marker.Properties["source_role"])
		}
		label, ok := marker.Properties["label"].(string)
		if !ok || len(label) != len("15:04") || label[2] != ':' {
			t.Fatalf("invalid time marker label %q", label)
		}
	}
}

func TestMarshalLunarEclipseUsesRequestedBoundarySampling(t *testing.T) {
	info, ok := eclipse.LunarEclipseOnDate(time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("expected lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 24)
	if err != nil {
		t.Fatalf("MarshalLunarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection, "visible-at-p1", "visible-at-p4", "p1-horizon", "p4-horizon", "greatest")
	assertCollectionCoordinates(t, collection)
	visible := featureWithRole(t, collection, "visible-at-p1")
	if got := visible.Properties["boundary_points"]; got != float64(24) {
		t.Fatalf("boundary_points=%v, want 24", got)
	}
	assertClosedMultiPolygon(t, visible)
	horizon := featureWithRole(t, collection, "p1-horizon")
	var lines [][][]float64
	if err := json.Unmarshal(horizon.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode P1 horizon: %v", err)
	}
	pointCount := 0
	for _, line := range lines {
		pointCount += len(line)
	}
	if pointCount < 24 {
		t.Fatalf("P1 horizon has %d points, want at least 24", pointCount)
	}
}

func TestMarshalLunarEclipseWithTimeMarkers(t *testing.T) {
	info, ok := eclipse.LunarEclipseOnDate(time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("expected lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipseWithTimeMarkers(info, 24, geojson.TimeMarkerOptions{Step: time.Hour})
	if err != nil {
		t.Fatalf("MarshalLunarEclipseWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	markers := featuresWithRole(collection, "time-marker")
	if len(markers) == 0 {
		t.Fatal("lunar eclipse has no time markers")
	}
	for _, marker := range markers {
		if marker.Properties["source_role"] != "sublunar-track" {
			t.Fatalf("time marker source_role=%v, want sublunar-track", marker.Properties["source_role"])
		}
	}
	firstLabel, _ := markers[0].Properties["label"].(string)
	lastLabel, _ := markers[len(markers)-1].Properties["label"].(string)
	if firstLabel != "09:00" || lastLabel != "14:00" {
		t.Fatalf("lunar marker endpoints = %q..%q, want 09:00..14:00", firstLabel, lastLabel)
	}
}

func TestMarshalLunarEclipseRejectsInvalidContactOrder(t *testing.T) {
	info, ok := eclipse.LunarEclipseOnDate(time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("expected lunar eclipse")
	}
	info.Maximum = info.PenumbralStart.Add(-time.Minute)
	if _, err := geojson.MarshalLunarEclipse(info, 24); err == nil {
		t.Fatal("reversed lunar eclipse contacts were accepted")
	}
}

func TestMarshalStarOccultationSplitsAntimeridian(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{160, 175, -175, -160}, []float64{8, 4, 0, -4})
	north := occultationSamples(start, []float64{158, 174, -174, -158}, []float64{18, 14, 10, 6})
	south := occultationSamples(start, []float64{162, 176, -176, -162}, []float64{-2, -6, -10, -14})
	path := moon.StarOccultationPath{
		TargetID:      "HR 4799",
		Start:         north[0],
		Greatest:      center[2],
		End:           north[len(north)-1],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: north,
		SouthernLimit: south,
		Step:          time.Hour,
	}

	data, err := geojson.MarshalStarOccultation(path)
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection, "occultation-band", "center-line", "north-limit", "south-limit", "start", "greatest", "end")
	assertCollectionCoordinates(t, collection)
	centerFeature := featureWithRole(t, collection, "center-line")
	var lines [][][]float64
	if err := json.Unmarshal(centerFeature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode center line: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("center line has %d antimeridian segments, want 2", len(lines))
	}
	for _, line := range lines {
		for index := 1; index < len(line); index++ {
			if math.Abs(line[index][0]-line[index-1][0]) > 180 {
				t.Fatalf("center line still crosses antimeridian: %#v", line)
			}
		}
	}
	assertTimedLineAligned(t, centerFeature)
}

func TestMarshalStarOccultationHandlesExactAntimeridian(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{-180, 180, 150}, []float64{2, 1, 0})
	north := occultationSamples(start, []float64{-179, 179, 178}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{-179, 179, 178}, []float64{-8, -9, -10})
	path := moon.StarOccultationPath{
		TargetID: "exact-antimeridian", Start: north[0], Greatest: center[1], End: north[2],
		Complete: true, CenterLine: center, NorthernLimit: north, SouthernLimit: south,
		Step: time.Hour,
	}
	if _, err := geojson.MarshalStarOccultation(path); err != nil {
		t.Fatalf("exact antimeridian path: %v", err)
	}
}

func TestMarshalStarOccultationWithTimeMarkers(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{20, 30, 40, 50}, []float64{2, 1, 0, -1})
	north := occultationSamples(start, []float64{20, 30, 40, 50}, []float64{12, 11, 10, 9})
	south := occultationSamples(start, []float64{20, 30, 40, 50}, []float64{-8, -9, -10, -11})
	path := moon.StarOccultationPath{
		TargetID:      "HR 4799",
		Start:         north[0],
		Greatest:      center[1],
		End:           north[len(north)-1],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: north,
		SouthernLimit: south,
		Step:          time.Hour,
	}

	data, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: time.Hour})
	if err != nil {
		t.Fatalf("MarshalStarOccultationWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	markers := featuresWithRole(collection, "time-marker")
	if len(markers) != 3 {
		t.Fatalf("got %d time markers, want 3", len(markers))
	}
	for _, marker := range markers {
		if marker.Properties["source_role"] != "center-line" {
			t.Fatalf("time marker source_role=%v, want center-line", marker.Properties["source_role"])
		}
		if marker.Geometry.Type != "Point" {
			t.Fatalf("time marker geometry=%q, want Point", marker.Geometry.Type)
		}
	}
}

func TestTimeMarkerInterpolationUsesShortestAntimeridianPath(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{170, -170}, []float64{2, 0})
	north := occultationSamples(start, []float64{168, -168}, []float64{12, 10})
	south := occultationSamples(start, []float64{172, -172}, []float64{-8, -10})
	path := moon.StarOccultationPath{
		TargetID:      "HR 4799",
		Start:         north[0],
		Greatest:      center[0],
		End:           north[len(north)-1],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: north,
		SouthernLimit: south,
		Step:          time.Hour,
	}

	data, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: 15 * time.Minute})
	if err != nil {
		t.Fatalf("MarshalStarOccultationWithTimeMarkers: %v", err)
	}
	markers := featuresWithRole(decodeCollection(t, data), "time-marker")
	if len(markers) != 3 {
		t.Fatalf("got %d time markers, want 3", len(markers))
	}
	for _, marker := range markers {
		var coordinate []float64
		if err := json.Unmarshal(marker.Geometry.Coordinates, &coordinate); err != nil {
			t.Fatalf("decode time marker: %v", err)
		}
		if math.Abs(coordinate[0]) < 170 {
			t.Fatalf("time marker crossed through longitude %.6f instead of the antimeridian", coordinate[0])
		}
	}
}

func TestMarshalPlanetOccultationIncludesPartialAndTotalFootprints(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{20, 30, 40}, []float64{2, 1, 0})
	north := occultationSamples(start, []float64{20, 30, 40}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{20, 30, 40}, []float64{-8, -9, -10})
	totalNorth := occultationSamples(start, []float64{22, 30, 38}, []float64{8, 7, 6})
	totalSouth := occultationSamples(start, []float64{22, 30, 38}, []float64{-4, -5, -6})
	for index := range totalNorth {
		at := start.Add(time.Duration(index+1) * 30 * time.Minute)
		totalNorth[index].Time = at
		totalSouth[index].Time = at
	}
	path := moon.PlanetOccultationPath{
		Planet:               moon.OccultationSaturn,
		TargetID:             "Saturn",
		Start:                north[0],
		Greatest:             center[1],
		End:                  north[len(north)-1],
		Complete:             true,
		CenterLine:           center,
		NorthernLimit:        north,
		SouthernLimit:        south,
		PartialFootprints:    []moon.PlanetOccultationFootprint{sampleFootprint(start.Add(time.Hour), 18, -10, 42, 14)},
		HasTotalBand:         true,
		TotalStart:           totalNorth[0],
		TotalEnd:             totalNorth[len(totalNorth)-1],
		TotalComplete:        true,
		NorthernTotalLimit:   totalNorth,
		SouthernTotalLimit:   totalSouth,
		TotalFootprints:      []moon.PlanetOccultationFootprint{sampleFootprint(start.Add(time.Hour), 23, -5, 37, 9)},
		GreatestTotalWidthKM: 2500,
		Step:                 time.Hour,
		TargetSpacingKM:      50,
	}

	data, err := geojson.MarshalPlanetOccultation(path)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	assertRoles(t, collection,
		"partial-footprint", "total-footprint", "center-line", "north-limit", "south-limit",
		"north-total-limit", "south-total-limit", "start", "total-start", "greatest", "total-end", "end")
	assertCollectionCoordinates(t, collection)
	if got := featureWithRole(t, collection, "greatest").Properties["planet"]; got != "saturn" {
		t.Fatalf("planet=%v, want saturn", got)
	}
}

func TestMarshalPlanetOccultationAllowsMissingCenterLine(t *testing.T) {
	start := time.Date(2024, time.September, 5, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.AddDate(0, 0, 1), moon.OccultationVenus, moon.OccultationPathOptions{},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultationPaths: %v", err)
	}
	if len(paths) != 1 || len(paths[0].CenterLine) != 0 {
		t.Fatalf("unexpected Venus path count/center line: paths=%d center=%d", len(paths), len(paths[0].CenterLine))
	}

	data, err := geojson.MarshalPlanetOccultationWithTimeMarkers(
		paths[0], geojson.TimeMarkerOptions{Step: 30 * time.Minute},
	)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultationWithTimeMarkers: %v", err)
	}
	collection := decodeCollection(t, data)
	if len(featuresWithRole(collection, "center-line")) != 0 || len(featuresWithRole(collection, "time-marker")) != 0 {
		t.Fatal("edge-only planetary path contains center-line features")
	}
	assertRoles(t, collection, "north-limit", "south-limit", "start", "greatest", "end")
}

func TestMarshalSolarEclipseRejectsMisalignedLimits(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{})
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{})
	if !ok {
		t.Fatal("expected solar central path")
	}
	central.SouthernLimit = central.SouthernLimit[:len(central.SouthernLimit)-1]
	if _, err := geojson.MarshalSolarEclipse(partial, &central); err == nil {
		t.Fatal("misaligned solar limits were accepted")
	}
}

func TestMarshalStarOccultationRejectsInvalidPathData(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	valid := sampleStarOccultationPath(start)
	tests := []struct {
		name   string
		mutate func(*moon.StarOccultationPath)
	}{
		{name: "misaligned limits", mutate: func(path *moon.StarOccultationPath) {
			path.SouthernLimit = path.SouthernLimit[:len(path.SouthernLimit)-1]
		}},
		{name: "mismatched limit times", mutate: func(path *moon.StarOccultationPath) {
			path.SouthernLimit[1].Time = path.SouthernLimit[1].Time.Add(time.Second)
		}},
		{name: "non-monotonic line", mutate: func(path *moon.StarOccultationPath) {
			path.CenterLine[1].Time = path.CenterLine[0].Time
		}},
		{name: "zero event time", mutate: func(path *moon.StarOccultationPath) {
			path.Start.Time = time.Time{}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := valid
			path.CenterLine = append([]moon.OccultationPathPoint(nil), valid.CenterLine...)
			path.NorthernLimit = append([]moon.OccultationPathPoint(nil), valid.NorthernLimit...)
			path.SouthernLimit = append([]moon.OccultationPathPoint(nil), valid.SouthernLimit...)
			test.mutate(&path)
			if _, err := geojson.MarshalStarOccultation(path); err == nil {
				t.Fatal("invalid stellar occultation path was accepted")
			}
		})
	}
}

func TestMarshalPlanetOccultationRejectsInvalidFootprintPolygon(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	path := samplePlanetOccultationPath(start)
	path.PartialFootprints = []moon.PlanetOccultationFootprint{sampleFootprint(start, 10, -10, 20, 10)}
	path.PartialFootprints[0].Polygons = append(path.PartialFootprints[0].Polygons, []moon.OccultationPathPoint{
		{Longitude: 30, Latitude: 0}, {Longitude: 31, Latitude: 0},
	})
	if _, err := geojson.MarshalPlanetOccultation(path); err == nil {
		t.Fatal("invalid footprint polygon was silently dropped")
	}
}

func TestMarshalFunctionsRejectIncompleteInput(t *testing.T) {
	if _, err := geojson.MarshalSolarEclipse(eclipse.SolarEclipsePartialFootprintsInfo{}, nil); err == nil {
		t.Fatal("empty solar eclipse input was accepted")
	}
	if _, err := geojson.MarshalLunarEclipse(eclipse.LunarEclipseInfo{}, 360); err == nil {
		t.Fatal("empty lunar eclipse input was accepted")
	}
	if _, err := geojson.MarshalStarOccultation(moon.StarOccultationPath{}); err == nil {
		t.Fatal("empty stellar occultation input was accepted")
	}
	if _, err := geojson.MarshalPlanetOccultation(moon.PlanetOccultationPath{}); err == nil {
		t.Fatal("empty planetary occultation input was accepted")
	}
}

func TestTimeMarkerOptionsRejectNegativeStep(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	center := occultationSamples(start, []float64{20, 30, 40}, []float64{2, 1, 0})
	north := occultationSamples(start, []float64{20, 30, 40}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{20, 30, 40}, []float64{-8, -9, -10})
	path := moon.StarOccultationPath{
		TargetID:      "HR 4799",
		Start:         north[0],
		Greatest:      center[1],
		End:           north[len(north)-1],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: north,
		SouthernLimit: south,
	}
	if _, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: -time.Minute}); err == nil {
		t.Fatal("negative time-marker step was accepted")
	}
	if _, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: time.Nanosecond}); err == nil {
		t.Fatal("sub-minute time-marker step was accepted")
	}

	excessiveEnd := path.CenterLine[0].Time.Add(24*time.Hour + 2*time.Minute)
	path.End.Time = excessiveEnd
	path.CenterLine[len(path.CenterLine)-1].Time = excessiveEnd
	path.NorthernLimit[len(path.NorthernLimit)-1].Time = excessiveEnd
	path.SouthernLimit[len(path.SouthernLimit)-1].Time = excessiveEnd
	if _, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: time.Minute}); err == nil {
		t.Fatal("excessive time-marker count was accepted")
	}
}

func TestTimeMarkerOptionsAreValidatedWithoutCenterLine(t *testing.T) {
	start := time.Date(2025, time.June, 5, 17, 45, 0, 0, time.UTC)
	north := occultationSamples(start, []float64{20, 30, 40}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{20, 30, 40}, []float64{-8, -9, -10})
	path := moon.StarOccultationPath{
		TargetID: "edge-only", Start: north[0], Greatest: north[1], End: north[2], Complete: true,
		NorthernLimit: north, SouthernLimit: south, Step: time.Hour,
	}
	if _, err := geojson.MarshalStarOccultationWithTimeMarkers(path, geojson.TimeMarkerOptions{Step: time.Nanosecond}); err == nil {
		t.Fatal("edge-only stellar path accepted sub-minute markers")
	}
	planet := moon.PlanetOccultationPath{
		Planet: moon.OccultationVenus, TargetID: "Venus", Start: north[0], Greatest: north[1], End: north[2],
		Complete: true, NorthernLimit: north, SouthernLimit: south, Step: time.Hour,
	}
	if _, err := geojson.MarshalPlanetOccultationWithTimeMarkers(planet, geojson.TimeMarkerOptions{Step: time.Nanosecond}); err == nil {
		t.Fatal("edge-only planetary path accepted sub-minute markers")
	}
}

func decodeCollection(t *testing.T, data []byte) decodedCollection {
	t.Helper()
	var collection decodedCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("decode GeoJSON: %v", err)
	}
	if collection.Type != "FeatureCollection" {
		t.Fatalf("collection type=%q, want FeatureCollection", collection.Type)
	}
	if len(collection.Features) == 0 {
		t.Fatal("GeoJSON contains no features")
	}
	for _, feature := range collection.Features {
		if feature.Type != "Feature" {
			t.Fatalf("feature type=%q, want Feature", feature.Type)
		}
		if _, ok := feature.Properties["event"]; !ok {
			t.Fatal("feature has no event property")
		}
		if _, ok := feature.Properties["role"]; !ok {
			t.Fatal("feature has no role property")
		}
	}
	return collection
}

func assertRoles(t *testing.T, collection decodedCollection, roles ...string) {
	t.Helper()
	for _, role := range roles {
		featureWithRole(t, collection, role)
	}
}

func featureWithRole(t *testing.T, collection decodedCollection, role string) decodedFeature {
	t.Helper()
	for _, feature := range collection.Features {
		if feature.Properties["role"] == role {
			return feature
		}
	}
	t.Fatalf("GeoJSON is missing role %q", role)
	return decodedFeature{}
}

func featuresWithRole(collection decodedCollection, role string) []decodedFeature {
	result := make([]decodedFeature, 0)
	for _, feature := range collection.Features {
		if feature.Properties["role"] == role {
			result = append(result, feature)
		}
	}
	return result
}

func assertCollectionCoordinates(t *testing.T, collection decodedCollection) {
	t.Helper()
	for _, feature := range collection.Features {
		var coordinates interface{}
		if err := json.Unmarshal(feature.Geometry.Coordinates, &coordinates); err != nil {
			t.Fatalf("decode %v coordinates: %v", feature.Properties["role"], err)
		}
		assertCoordinateTree(t, coordinates)
	}
}

func assertCoordinateTree(t *testing.T, value interface{}) {
	t.Helper()
	items, ok := value.([]interface{})
	if !ok {
		t.Fatalf("coordinate node has type %T", value)
	}
	if len(items) >= 2 {
		longitude, lonOK := items[0].(float64)
		latitude, latOK := items[1].(float64)
		if lonOK && latOK {
			if math.IsNaN(longitude) || math.IsInf(longitude, 0) || longitude < -180 || longitude > 180 {
				t.Fatalf("invalid longitude %.12f", longitude)
			}
			if math.IsNaN(latitude) || math.IsInf(latitude, 0) || latitude < -90 || latitude > 90 {
				t.Fatalf("invalid latitude %.12f", latitude)
			}
			return
		}
	}
	for _, item := range items {
		assertCoordinateTree(t, item)
	}
}

func assertClosedMultiPolygon(t *testing.T, feature decodedFeature) {
	t.Helper()
	if feature.Geometry.Type != "MultiPolygon" {
		t.Fatalf("%v geometry=%q, want MultiPolygon", feature.Properties["role"], feature.Geometry.Type)
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode %v polygon: %v", feature.Properties["role"], err)
	}
	if len(polygons) == 0 {
		t.Fatalf("%v has no polygons", feature.Properties["role"])
	}
	for _, polygon := range polygons {
		if len(polygon) == 0 || len(polygon[0]) < 4 {
			t.Fatalf("%v contains an incomplete ring", feature.Properties["role"])
		}
		ring := polygon[0]
		first, last := ring[0], ring[len(ring)-1]
		if first[0] != last[0] || first[1] != last[1] {
			t.Fatalf("%v ring is not closed", feature.Properties["role"])
		}
	}
}

func assertTimedLineAligned(t *testing.T, feature decodedFeature) {
	t.Helper()
	var lines [][][]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
		t.Fatalf("decode line coordinates: %v", err)
	}
	timeSegments, ok := feature.Properties["times"].([]interface{})
	if !ok || len(timeSegments) != len(lines) {
		t.Fatalf("times do not align with %d line segments: %#v", len(lines), feature.Properties["times"])
	}
	for index, rawSegment := range timeSegments {
		times, ok := rawSegment.([]interface{})
		if !ok || len(times) != len(lines[index]) {
			t.Fatalf("times segment %d does not align with %d coordinates", index, len(lines[index]))
		}
		for _, rawTime := range times {
			value, ok := rawTime.(string)
			if !ok {
				t.Fatalf("time has type %T", rawTime)
			}
			if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
				t.Fatalf("invalid RFC3339 time %q: %v", value, err)
			}
		}
	}
}

func occultationSamples(start time.Time, longitudes, latitudes []float64) []moon.OccultationPathPoint {
	result := make([]moon.OccultationPathPoint, len(longitudes))
	for index := range result {
		result[index] = moon.OccultationPathPoint{
			Time:         start.Add(time.Duration(index) * time.Hour),
			Longitude:    longitudes[index],
			Latitude:     latitudes[index],
			MoonAltitude: 30,
			WidthKM:      3000,
		}
	}
	return result
}

func sampleStarOccultationPath(start time.Time) moon.StarOccultationPath {
	center := occultationSamples(start, []float64{20, 30, 40}, []float64{2, 1, 0})
	north := occultationSamples(start, []float64{20, 30, 40}, []float64{12, 11, 10})
	south := occultationSamples(start, []float64{20, 30, 40}, []float64{-8, -9, -10})
	return moon.StarOccultationPath{
		TargetID: "HR 4799", Start: north[0], Greatest: center[1], End: north[len(north)-1],
		Complete: true, CenterLine: center, NorthernLimit: north, SouthernLimit: south, Step: time.Hour,
	}
}

func samplePlanetOccultationPath(start time.Time) moon.PlanetOccultationPath {
	star := sampleStarOccultationPath(start)
	return moon.PlanetOccultationPath{
		Planet: moon.OccultationSaturn, TargetID: "Saturn",
		Start: star.Start, Greatest: star.Greatest, End: star.End, Complete: true,
		CenterLine: star.CenterLine, NorthernLimit: star.NorthernLimit, SouthernLimit: star.SouthernLimit,
		Step: time.Hour,
	}
}

func sampleFootprint(at time.Time, west, south, east, north float64) moon.PlanetOccultationFootprint {
	return moon.PlanetOccultationFootprint{
		Time: at,
		Polygons: [][]moon.OccultationPathPoint{{
			{Longitude: west, Latitude: south},
			{Longitude: east, Latitude: south},
			{Longitude: east, Latitude: north},
			{Longitude: west, Latitude: north},
		}},
	}
}
