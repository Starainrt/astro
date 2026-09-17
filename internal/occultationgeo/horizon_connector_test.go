package occultationgeo

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func TestHorizonConnectorTrimsSamplesOutsideExactEndpoints(t *testing.T) {
	for _, offset := range []float64{0, 177, -182} {
		var points []basic.OccultationPathPoint
		for _, longitude := range []float64{1, 0, 2, 3, 5, 4} {
			points = append(points, basic.OccultationPathPoint{Longitude: math.Remainder(longitude+offset, 360), Latitude: 0})
		}
		trimmed := trimHorizonConnectorSamples(points)
		want := []basic.OccultationPathPoint{points[0], points[2], points[3], points[5]}
		if len(trimmed) != len(want) {
			t.Fatalf("offset=%g retained %d points, want %d", offset, len(trimmed), len(want))
		}
		for i := range want {
			if trimmed[i] != want[i] {
				t.Fatalf("offset=%g vertex %d changed: got %+v want %+v", offset, i, trimmed[i], want[i])
			}
		}
	}
}

func TestHorizonConnectorSegmentsUsesHorizonClosureAndPhaseEndpoints(t *testing.T) {
	eventTime := time.Date(2025, time.July, 29, 0, 0, 0, 0, time.UTC)
	contactStart := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 0}
	contactEnd := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 2}
	phaseStart := basic.OccultationPathPoint{Time: eventTime.Add(time.Second), Longitude: -0.1, Latitude: 2}
	phaseGreatest := basic.OccultationPathPoint{Time: eventTime.Add(-time.Second), Longitude: -0.1, Latitude: 0}
	footprint := basic.OccultationFootprint{
		Boundaries: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 4, Latitude: 1},
			contactEnd,
		}},
		Polygons: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 4, Latitude: 1},
			contactEnd,
			{Time: eventTime, Longitude: 1, Latitude: 2},
			{Time: eventTime, Longitude: 1, Latitude: 1},
			{Time: eventTime, Longitude: 1, Latitude: 0},
			contactStart,
		}},
	}
	curves := []basic.OccultationRiseSetCurve{
		{
			Phase: basic.RiseSetPhaseStart, Direction: basic.RiseSetDirectionRise,
			Segments: [][]basic.OccultationPathPoint{{phaseStart, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: 3}}},
		},
		{
			Phase: basic.RiseSetPhaseGreatest, Direction: basic.RiseSetDirectionRise,
			Segments: [][]basic.OccultationPathPoint{{phaseGreatest, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: -1}}},
		},
	}

	connectors := HorizonConnectorSegments([]basic.OccultationFootprint{footprint}, curves)
	if len(connectors) != 1 {
		t.Fatalf("connectors=%d, want one", len(connectors))
	}
	points := connectors[0].Points
	if len(points) <= 4 {
		t.Fatalf("connector points=%d, want a densified horizon closure", len(points))
	}
	if points[0] != phaseStart || points[len(points)-1] != phaseGreatest {
		t.Fatalf("connector endpoints=(%+v, %+v), want phase endpoints (%+v, %+v)",
			points[0], points[len(points)-1], phaseStart, phaseGreatest)
	}
	for _, point := range points {
		if point.Longitude == 4 {
			t.Fatal("connector contains the contact-limb detour")
		}
	}
}

func TestSmoothHorizonConnectorJunctionsRemovesShortEndpointCorner(t *testing.T) {
	points := []basic.OccultationPathPoint{
		{Longitude: -49.9687087, Latitude: 22.7599834},
		{Longitude: -49.7995044, Latitude: 22.6531000},
		{Longitude: -49.8520050, Latitude: 22.4586340},
		{Longitude: -49.9043580, Latitude: 22.2641500},
	}
	cleaned := smoothHorizonConnectorJunctions(points)
	if len(cleaned) != len(points)-1 {
		t.Fatalf("cleaned point count=%d, want %d: %#v", len(cleaned), len(points)-1, cleaned)
	}
	if cleaned[0] != points[0] || cleaned[len(cleaned)-1] != points[len(points)-1] {
		t.Fatalf("connector endpoints changed: first=%+v last=%+v", cleaned[0], cleaned[len(cleaned)-1])
	}
}

func TestSmoothHorizonConnectorJunctionsKeepsBroadBend(t *testing.T) {
	points := []basic.OccultationPathPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 0.2, Latitude: 0.1},
		{Longitude: 1.0, Latitude: 0.8},
		{Longitude: 2.0, Latitude: 1.8},
	}
	cleaned := smoothHorizonConnectorJunctions(points)
	if len(cleaned) != len(points) {
		t.Fatalf("broad physical bend was changed: got %d points, want %d", len(cleaned), len(points))
	}
}

func TestHorizonConnectorSegmentsSmoothsExactEndpointJunctions(t *testing.T) {
	cases := []struct {
		name   string
		planet basic.OccultationPlanet
		start  time.Time
		target time.Time
	}{
		{"venus-1027", basic.OccultationVenus, time.Date(1027, time.March, 13, 0, 0, 0, 0, time.UTC), time.Date(1027, time.March, 13, 22, 11, 50, 833970457, time.UTC)},
		{"saturn-2629", basic.OccultationSaturn, time.Date(2629, time.December, 30, 0, 0, 0, 0, time.UTC), time.Date(2629, time.December, 30, 18, 28, 18, 614121973, time.UTC)},
	}
	for _, sample := range cases {
		t.Run(sample.name, func(t *testing.T) {
			paths, err := basic.FindPlanetOccultationPaths(sample.start.Add(-12*time.Hour), sample.start.Add(36*time.Hour), sample.planet, basic.OccultationPathOptions{
				Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
				DisableFootprints: true, Algorithm: basic.OccultationPathAlgorithmExact,
			})
			if err != nil || len(paths) == 0 {
				t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v", len(paths), err)
			}
			path := paths[0]
			for _, candidate := range paths[1:] {
				if absDuration(candidate.Greatest.Time.Sub(sample.target)) < absDuration(path.Greatest.Time.Sub(sample.target)) {
					path = candidate
				}
			}
			curvesByBand := []struct {
				name       string
				footprints []basic.OccultationFootprint
				curves     []basic.OccultationRiseSetCurve
			}{
				{"partial", path.PartialBandFootprints, path.RiseSetCurves},
				{"total", path.TotalBandFootprints, path.TotalRiseSetCurves},
			}
			for _, band := range curvesByBand {
				for connectorIndex, connector := range HorizonConnectorSegments(band.footprints, band.curves) {
					for index := 1; index+1 < len(connector.Points); index++ {
						first, middle, last := connector.Points[index-1], connector.Points[index], connector.Points[index+1]
						turn := 180 - occultationTurnAngleDegrees(
							geodata.GeoPoint{Longitude: first.Longitude, Latitude: first.Latitude},
							geodata.GeoPoint{Longitude: middle.Longitude, Latitude: middle.Latitude},
							geodata.GeoPoint{Longitude: last.Longitude, Latitude: last.Latitude},
						)
						deviation := occultationProjectedPointLineDistanceKM(
							geodata.GeoPoint{Longitude: middle.Longitude, Latitude: middle.Latitude},
							geodata.GeoPoint{Longitude: first.Longitude, Latitude: first.Latitude},
							geodata.GeoPoint{Longitude: last.Longitude, Latitude: last.Latitude},
						)
						if turn >= 55 && deviation >= 5 && deviation <= 25 &&
							DistanceKM(first, middle) <= 75 && DistanceKM(middle, last) <= 75 && DistanceKM(first, last) <= 55 {
							t.Fatalf("%s connector %d retains short endpoint corner at %d: turn=%.1f deviation=%.1f points=%+v/%+v/%+v", band.name, connectorIndex, index, turn, deviation, first, middle, last)
						}
					}
				}
			}
		})
	}
}

func TestHorizonConnectorSegmentsSkipsSamePhaseFoldBranches(t *testing.T) {
	eventTime := time.Date(2025, time.July, 29, 0, 0, 0, 0, time.UTC)
	contactStart := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 0}
	contactEnd := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 1}
	terminal := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 0.5}
	footprint := basic.OccultationFootprint{
		Boundaries: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 0.5},
			contactEnd,
		}},
		Polygons: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 0.5},
			contactEnd,
			{Time: eventTime, Longitude: 1, Latitude: 1},
			{Time: eventTime, Longitude: 1, Latitude: 0},
			contactStart,
		}},
	}
	curves := []basic.OccultationRiseSetCurve{{
		Phase: basic.RiseSetPhaseEnd, Direction: basic.RiseSetDirectionRise,
		Segments: [][]basic.OccultationPathPoint{
			{terminal, {Time: eventTime.Add(time.Minute), Longitude: 1, Latitude: 1}},
			{terminal, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: 1}},
		},
	}}

	connectors := HorizonConnectorSegments([]basic.OccultationFootprint{footprint}, curves)
	if len(connectors) != 0 {
		t.Fatalf("connectors=%d, want no duplicate terminal closure", len(connectors))
	}
}

func TestHorizonConnectorSegmentsSkipsCoincidentDifferentPhaseEndpoints(t *testing.T) {
	eventTime := time.Date(2025, time.July, 29, 0, 0, 0, 0, time.UTC)
	contactStart := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 0}
	contactEnd := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 1}
	shared := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 0.5}
	footprint := basic.OccultationFootprint{
		Boundaries: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 0.5},
			contactEnd,
		}},
		Polygons: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 0.5},
			contactEnd,
			{Time: eventTime, Longitude: 1, Latitude: 1},
			shared,
			{Time: eventTime, Longitude: 1, Latitude: 0},
			contactStart,
		}},
	}
	curves := []basic.OccultationRiseSetCurve{
		{Phase: basic.RiseSetPhaseStart, Direction: basic.RiseSetDirectionRise,
			Segments: [][]basic.OccultationPathPoint{{shared, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: 1}}}},
		{Phase: basic.RiseSetPhaseGreatest, Direction: basic.RiseSetDirectionRise,
			Segments: [][]basic.OccultationPathPoint{{shared, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: -1}}}},
	}

	if connectors := HorizonConnectorSegments([]basic.OccultationFootprint{footprint}, curves); len(connectors) != 0 {
		t.Fatalf("connectors=%d, want no zero-length phase pinch loop", len(connectors))
	}
}

func TestHorizonConnectorSegmentsClosesSamePhaseSeparatedBranchOpening(t *testing.T) {
	eventTime := time.Date(2025, time.July, 29, 0, 0, 0, 0, time.UTC)
	contactStart := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 0}
	contactEnd := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 3}
	branchStart := basic.OccultationPathPoint{Time: eventTime, Longitude: -0.05, Latitude: 0}
	branchEnd := basic.OccultationPathPoint{Time: eventTime, Longitude: -0.05, Latitude: 3}
	footprint := basic.OccultationFootprint{
		Boundaries: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 1.5},
			contactEnd,
		}},
		Polygons: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 1.5},
			contactEnd,
			{Time: eventTime, Longitude: 1, Latitude: 3},
			{Time: eventTime, Longitude: 1, Latitude: 0},
			contactStart,
		}},
	}
	curves := []basic.OccultationRiseSetCurve{{
		Phase: basic.RiseSetPhaseStart, Direction: basic.RiseSetDirectionRise,
		Segments: [][]basic.OccultationPathPoint{
			{branchStart, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: -1}},
			{branchEnd, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: 4}},
		},
	}}

	connectors := HorizonConnectorSegments([]basic.OccultationFootprint{footprint}, curves)
	if len(connectors) != 1 {
		t.Fatalf("connectors=%d, want one same-phase branch-opening closure", len(connectors))
	}
	points := connectors[0].Points
	if len(points) <= 4 {
		t.Fatalf("connector points=%d, want a densified horizon closure", len(points))
	}
	if !sameTestOccultationPoint(points[0], branchEnd) || !sameTestOccultationPoint(points[len(points)-1], branchStart) {
		t.Fatalf("connector endpoints=(%+v, %+v), want same-phase branch endpoints (%+v, %+v)",
			points[0], points[len(points)-1], branchEnd, branchStart)
	}
}

func TestHorizonConnectorSegmentsClosesShortTemporalHorizonGap(t *testing.T) {
	startTime := time.Date(2023, time.September, 21, 10, 1, 49, 0, time.UTC)
	endTime := startTime.Add(time.Minute)
	start := basic.OccultationPathPoint{Time: startTime, Longitude: 10, Latitude: 20}
	end := basic.OccultationPathPoint{Time: endTime, Longitude: 10.2, Latitude: 20.1}
	farStart := basic.OccultationPathPoint{Time: startTime, Longitude: 30, Latitude: 20}
	farEnd := basic.OccultationPathPoint{Time: endTime, Longitude: 30, Latitude: 20}
	footprints := []basic.OccultationFootprint{
		{
			Time:       startTime,
			Boundaries: [][]basic.OccultationPathPoint{{start, farStart}},
			Polygons: [][]basic.OccultationPathPoint{{
				start, farStart,
				{Time: startTime, Longitude: 30, Latitude: 19},
				{Time: startTime, Longitude: 10, Latitude: 19}, start,
			}},
		},
		{
			Time:       endTime,
			Boundaries: [][]basic.OccultationPathPoint{{end, farEnd}},
			Polygons: [][]basic.OccultationPathPoint{{
				end, farEnd,
				{Time: endTime, Longitude: 30, Latitude: 19},
				{Time: endTime, Longitude: 10.2, Latitude: 19}, end,
			}},
		},
	}
	curves := []basic.OccultationRiseSetCurve{
		{
			Phase: basic.RiseSetPhaseStart, Direction: basic.RiseSetDirectionSet,
			Segments: [][]basic.OccultationPathPoint{{
				{Time: startTime.Add(-10 * time.Minute), Longitude: 0, Latitude: 30}, start,
			}},
		},
		{
			Phase: basic.RiseSetPhaseEnd, Direction: basic.RiseSetDirectionSet,
			Segments: [][]basic.OccultationPathPoint{{
				end, {Time: endTime.Add(10 * time.Minute), Longitude: 0, Latitude: 30},
			}},
		},
	}

	if connectors := HorizonConnectorSegments(footprints, curves); len(connectors) != 0 {
		t.Fatalf("finite-disk connectors=%d, want no point-source temporal connector", len(connectors))
	}
	connectors := StarHorizonConnectorSegments(footprints, curves)
	if len(connectors) != 1 {
		t.Fatalf("connectors=%d, want one temporal horizon connector", len(connectors))
	}
	points := connectors[0].Points
	if len(points) < 2 || points[0] != start || points[len(points)-1] != end {
		t.Fatalf("connector endpoints=%+v -> %+v, want %+v -> %+v",
			points[0], points[len(points)-1], start, end)
	}
	for index := 1; index < len(points); index++ {
		if !points[index].Time.After(points[index-1].Time) {
			t.Fatalf("connector time is not strictly increasing at %d", index)
		}
	}
}

func TestHorizonConnectorSegmentsMars20250729OneMinuteRiseSetStepKeepsOpeningClosure(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationMars,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	connectors := HorizonConnectorSegments(paths[0].PartialBandFootprints, paths[0].RiseSetCurves)
	foundOpening := false
	for _, connector := range connectors {
		if connector.Direction != basic.RiseSetDirectionRise || len(connector.Points) < 2 {
			continue
		}
		first := connector.Points[0]
		last := connector.Points[len(connector.Points)-1]
		if !mars20250729OpeningFoldEndpoint(first) || !mars20250729OpeningFoldEndpoint(last) {
			continue
		}
		chord := DistanceKM(first, last)
		if chord < 250 || chord > 650 {
			t.Fatalf("opening connector chord %.1f km, want the missing same-phase branch gap", chord)
		}
		foundOpening = true
		break
	}
	if !foundOpening {
		t.Fatal("2025-07-29 Mars one-minute rise/set curves are missing the opening horizon closure")
	}
}

func sameTestOccultationPoint(first, second basic.OccultationPathPoint) bool {
	return first.Time.Equal(second.Time) &&
		first.Longitude == second.Longitude &&
		first.Latitude == second.Latitude
}

func mars20250729OpeningFoldEndpoint(point basic.OccultationPathPoint) bool {
	return point.Longitude >= -128.5 && point.Longitude <= -126 &&
		point.Latitude >= -48.5 && point.Latitude <= -42
}
