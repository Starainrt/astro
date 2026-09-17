package geodata

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestOpenBoundarySweepContainsIntermediateArcs(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{{Longitude: -2, Latitude: 0}, {Longitude: 2, Latitude: 0}}}},
		{Boundaries: [][]GeoPoint{{
			{Longitude: -1.5, Latitude: 0.5},
			{Longitude: 0, Latitude: 3},
			{Longitude: 1.5, Latitude: 0.5},
		}}},
		{Boundaries: [][]GeoPoint{{{Longitude: -1, Latitude: 1}, {Longitude: 1, Latitude: 1}}}},
	}
	polygons, err := OpenBoundarySweep(samples)
	if err != nil {
		t.Fatalf("OpenBoundarySweep: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("polygon count = %d, want 1", len(polygons))
	}
	for _, point := range []GeoPoint{
		{Longitude: -2, Latitude: 0},
		{Longitude: 0, Latitude: 3},
		{Longitude: 2, Latitude: 0},
	} {
		if !sweepPolygonContainsOrTouches(polygons[0], point) {
			t.Fatalf("sweep polygon does not contain boundary point %+v", point)
		}
	}
}

func TestOpenBoundarySweepSeparatesGroupsAcrossClosedSample(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{{Longitude: -20, Latitude: 0}, {Longitude: -10, Latitude: 0}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: -19, Latitude: 1}, {Longitude: -9, Latitude: 1}}}},
		{Closed: true, Boundaries: [][]GeoPoint{{
			{Longitude: 0, Latitude: 0}, {Longitude: 1, Latitude: 0}, {Longitude: 0, Latitude: 1},
		}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 10, Latitude: 0}, {Longitude: 20, Latitude: 0}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 11, Latitude: 1}, {Longitude: 21, Latitude: 1}}}},
	}
	polygons, err := OpenBoundarySweep(samples)
	if err != nil {
		t.Fatalf("OpenBoundarySweep: %v", err)
	}
	if len(polygons) != 2 {
		t.Fatalf("polygon count=%d, want two groups separated by the closed sample", len(polygons))
	}
}

func TestMonotoneOpenBoundarySweepUsesEndpointTracks(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{{Longitude: -2, Latitude: 0}, {Longitude: 2, Latitude: 0}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: -1.5, Latitude: 1}, {Longitude: 1.5, Latitude: 1}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: -1, Latitude: 2}, {Longitude: 1, Latitude: 2}}}},
	}
	polygons, err := MonotoneOpenBoundarySweep(samples)
	if err != nil {
		t.Fatalf("MonotoneOpenBoundarySweep: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("polygon count=%d, want one monotone outline", len(polygons))
	}
	for _, point := range []GeoPoint{
		{Longitude: -2, Latitude: 0},
		{Longitude: 2, Latitude: 0},
		{Longitude: -1.5, Latitude: 1},
		{Longitude: 1.5, Latitude: 1},
		{Longitude: -1, Latitude: 2},
		{Longitude: 1, Latitude: 2},
	} {
		found := false
		for _, candidate := range polygons[0] {
			if SameGeoPoint(candidate, point) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("monotone outline is missing endpoint-track point %+v", point)
		}
	}
}

func TestMonotoneOpenBoundarySweepRejectsIntermediateBulge(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{{Longitude: -2, Latitude: 0}, {Longitude: 2, Latitude: 0}}}},
		{Boundaries: [][]GeoPoint{{
			{Longitude: -1.5, Latitude: 0.5},
			{Longitude: 0, Latitude: 3},
			{Longitude: 1.5, Latitude: 0.5},
		}}},
		{Boundaries: [][]GeoPoint{{{Longitude: -1, Latitude: 1}, {Longitude: 1, Latitude: 1}}}},
	}
	if _, err := MonotoneOpenBoundarySweep(samples); err == nil {
		t.Fatal("monotone sweep accepted an intermediate arc outside its endpoint tracks")
	}
}

func TestMonotoneOpenBoundarySweepRejectsEndpointTrackReversal(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{{Longitude: 0, Latitude: 0}, {Longitude: 2, Latitude: 0}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 1, Latitude: 0.5}, {Longitude: 3, Latitude: 0.5}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 0.5, Latitude: 1}, {Longitude: 2.5, Latitude: 1}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 0, Latitude: 1.5}, {Longitude: 2, Latitude: 1.5}}}},
	}
	if _, err := MonotoneOpenBoundarySweep(samples); !errors.Is(err, ErrOpenBoundaryEndpointTrackReversal) {
		t.Fatalf("monotone sweep error=%v, want endpoint-track reversal", err)
	}
}

func TestMonotoneOpenBoundarySweepAllowsSubKilometreEndpointNoise(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{{Longitude: 0, Latitude: 0}, {Longitude: 2, Latitude: 0}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 0.005, Latitude: 0.5}, {Longitude: 2.005, Latitude: 0.5}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 0.001, Latitude: 1}, {Longitude: 2.001, Latitude: 1}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 0.01, Latitude: 1.5}, {Longitude: 2.01, Latitude: 1.5}}}},
	}
	if _, err := MonotoneOpenBoundarySweep(samples); err != nil {
		t.Fatalf("monotone sweep rejected sub-kilometre endpoint noise: %v", err)
	}
}

func TestOpenBoundaryEndpointOutlinesRetainTracksAcrossIntermediateBulge(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{{Longitude: 0, Latitude: 0}, {Longitude: 4, Latitude: 0}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 0, Latitude: 1}, {Longitude: 2, Latitude: 4}, {Longitude: 4, Latitude: 1}}}},
		{Boundaries: [][]GeoPoint{{{Longitude: 0, Latitude: 2}, {Longitude: 4, Latitude: 2}}}},
	}

	polygons, err := OpenBoundaryEndpointOutlines(samples)
	if err != nil {
		t.Fatalf("OpenBoundaryEndpointOutlines: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("polygon count=%d, want 1", len(polygons))
	}
	for _, point := range []GeoPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 0, Latitude: 1},
		{Longitude: 0, Latitude: 2},
		{Longitude: 4, Latitude: 0},
		{Longitude: 4, Latitude: 1},
		{Longitude: 4, Latitude: 2},
	} {
		if !sweepPolygonContainsOrTouches(polygons[0], point) {
			t.Fatalf("endpoint outline does not contain track point %+v", point)
		}
	}
	if sweepPolygonContainsOrTouches(polygons[0], GeoPoint{Longitude: 2, Latitude: 4}) {
		t.Fatal("endpoint-only outline unexpectedly contains the intermediate bulge")
	}
}

func TestOpenBoundaryEndpointOutlinesRejectSingleSampleGroups(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{{Longitude: 0, Latitude: 0}, {Longitude: 1, Latitude: 1}, {Longitude: 2, Latitude: 0}}}},
		{Closed: true},
		{Boundaries: [][]GeoPoint{{{Longitude: 10, Latitude: 0}, {Longitude: 11, Latitude: 1}, {Longitude: 12, Latitude: 0}}}},
	}
	if _, err := OpenBoundaryEndpointOutlines(samples); err == nil {
		t.Fatal("endpoint outlines accepted groups without an endpoint track")
	}
}

func TestDecimateOpenBoundarySweepSamplesPreservesGroupsAndEndpoints(t *testing.T) {
	var samples []OpenBoundarySweepSample
	for group := 0; group < 2; group++ {
		for index := 0; index < 10; index++ {
			longitude := float64(group*100 + index)
			samples = append(samples, OpenBoundarySweepSample{Boundaries: [][]GeoPoint{{
				{Longitude: longitude, Latitude: 0},
				{Longitude: longitude + 0.1, Latitude: 0.1},
				{Longitude: longitude + 0.2, Latitude: 0.2},
			}}})
		}
		if group == 0 {
			samples = append(samples, OpenBoundarySweepSample{Closed: true})
		}
	}
	decimated := DecimateOpenBoundarySweepSamples(samples, 4, 20)
	if len(decimated) != 9 || !decimated[4].Closed {
		t.Fatalf("decimated samples=%d separator=%v, want two groups of four", len(decimated), decimated[4].Closed)
	}
	for _, index := range []int{0, 3, 5, 8} {
		if len(decimated[index].Boundaries) != 1 || len(decimated[index].Boundaries[0]) != 2 {
			t.Fatalf("sample %d boundary was not spatially decimated: %#v", index, decimated[index].Boundaries)
		}
	}
	for index, wantLongitude := range map[int]float64{0: 0, 3: 9, 5: 100, 8: 109} {
		if got := decimated[index].Boundaries[0][0].Longitude; got != wantLongitude {
			t.Fatalf("sample %d first longitude=%.1f, want %.1f", index, got, wantLongitude)
		}
	}
}

func TestOpenBoundarySweepInnerCapsCloseOnlyNearSeparator(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{
			{Longitude: -2, Latitude: 0}, {Longitude: 0, Latitude: 2}, {Longitude: 2, Latitude: 0},
		}}},
		{Closed: true},
		{Boundaries: [][]GeoPoint{{
			{Longitude: 10, Latitude: 0}, {Longitude: 12, Latitude: 2}, {Longitude: 14, Latitude: 0},
		}}},
	}
	if caps := OpenBoundarySweepInnerCaps(samples, 500); len(caps) != 2 {
		t.Fatalf("inner caps=%d, want both sides of the closed separator", len(caps))
	}
	if caps := OpenBoundarySweepInnerCaps(samples, 100); len(caps) != 0 {
		t.Fatalf("inner caps=%d, want large horizon gaps left open", len(caps))
	}
}

func TestOpenBoundarySweepStaysContinuousAcrossAntimeridian(t *testing.T) {
	samples := []OpenBoundarySweepSample{
		{Boundaries: [][]GeoPoint{{
			{Longitude: 170, Latitude: 82},
			{Longitude: -170, Latitude: 82},
		}}},
		{Boundaries: [][]GeoPoint{{
			{Longitude: 172, Latitude: 84},
			{Longitude: -168, Latitude: 84},
		}}},
		{Boundaries: [][]GeoPoint{{
			{Longitude: 175, Latitude: 86},
			{Longitude: -165, Latitude: 86},
		}}},
	}
	polygons, err := OpenBoundarySweep(samples)
	if err != nil {
		t.Fatalf("OpenBoundarySweep across antimeridian: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("antimeridian sweep polygon count=%d, want one", len(polygons))
	}
	for _, sample := range samples {
		for _, point := range sample.Boundaries[0] {
			if !sphericalPolygonContainsOrTouches(polygons[0], point) {
				t.Fatalf("sweep excludes sampled endpoint %+v", point)
			}
		}
	}
}

func TestSweepPointInPolygonIncludesBoundary(t *testing.T) {
	polygon := []GeoPoint{
		{Longitude: 120, Latitude: -72},
		{Longitude: 135, Latitude: -72},
		{Longitude: 135, Latitude: -68},
		{Longitude: 120, Latitude: -68},
	}
	for _, point := range []GeoPoint{
		{Longitude: 120, Latitude: -72},
		{Longitude: 127.5, Latitude: -72},
		{Longitude: 135, Latitude: -70},
		{Longitude: 127.5, Latitude: -70},
	} {
		if !sweepPointInPolygon(polygon, point) {
			t.Fatalf("boundary/interior point %#v was rejected", point)
		}
	}
	if sweepPointInPolygon(polygon, GeoPoint{Longitude: 127.5, Latitude: -73}) {
		t.Fatal("outside point was accepted")
	}
}

func sweepPolygonContainsOrTouches(polygon []GeoPoint, point GeoPoint) bool {
	inside := false
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		first, second := polygon[previous], polygon[current]
		cross := (point.Longitude-first.Longitude)*(second.Latitude-first.Latitude) -
			(point.Latitude-first.Latitude)*(second.Longitude-first.Longitude)
		if math.Abs(cross) <= 1e-9 && point.Longitude >= math.Min(first.Longitude, second.Longitude)-1e-9 &&
			point.Longitude <= math.Max(first.Longitude, second.Longitude)+1e-9 &&
			point.Latitude >= math.Min(first.Latitude, second.Latitude)-1e-9 &&
			point.Latitude <= math.Max(first.Latitude, second.Latitude)+1e-9 {
			return true
		}
		if (first.Latitude > point.Latitude) != (second.Latitude > point.Latitude) &&
			point.Longitude < (second.Longitude-first.Longitude)*(point.Latitude-first.Latitude)/
				(second.Latitude-first.Latitude)+first.Longitude {
			inside = !inside
		}
	}
	return inside
}

func TestSweepBridgeTouchingPolygonsBridgesNearPair(t *testing.T) {
	polygons := [][]GeoPoint{
		{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
		{{0, 0.00005}, {1, 0.00005}, {1, 1.00005}, {0, 1.00005}},
	}
	result, err := sweepBridgeTouchingPolygons(polygons)
	if err != nil {
		t.Fatalf("sweepBridgeTouchingPolygons: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("ring count=%d, want one bridged ring", len(result))
	}
	for _, point := range []GeoPoint{{Longitude: 0.5, Latitude: 0.5}, {Longitude: 0.5, Latitude: 0.00002}} {
		if !sweepPointInPolygon(result[0], point) {
			t.Fatalf("bridged ring misses %+v", point)
		}
	}
}

func TestSweepBridgeTouchingPolygonsSeparatesDistantRings(t *testing.T) {
	polygons := [][]GeoPoint{
		{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
		{{0, 0.001}, {1, 0.001}, {1, 1.001}, {0, 1.001}},
	}
	result, err := sweepBridgeTouchingPolygons(polygons)
	if err != nil {
		t.Fatalf("sweepBridgeTouchingPolygons: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("ring count=%d, want two rings a kilometre apart", len(result))
	}
}

func TestSweepBridgeTouchingPolygonsScansManyRingsQuickly(t *testing.T) {
	rings := make([][]GeoPoint, 0, 120)
	for row := 0; row < 10; row++ {
		for col := 0; col < 12; col++ {
			rings = append(rings, SphericalCircle(GeoPoint{
				Longitude: 15 * float64(col),
				Latitude:  10 + 5*float64(row),
			}, 0.5, 128))
		}
	}
	rings[119] = SphericalCircle(GeoPoint{Longitude: 150.0005, Latitude: 55}, 0.5, 128)
	start := time.Now()
	result, err := sweepBridgeTouchingPolygons(rings)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("sweepBridgeTouchingPolygons: %v", err)
	}
	if len(result) != len(rings) {
		t.Fatalf("ring count=%d, want %d", len(result), len(rings))
	}
	if elapsed > 2*time.Second {
		t.Fatalf("scanning 120 rings took %v", elapsed)
	}
}
