package occultationgeo

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func differentialRing(center geodata.GeoPoint, radius float64, points int, jitter *rand.Rand) []geodata.GeoPoint {
	ring := make([]geodata.GeoPoint, 0, points+1)
	for index := 0; index <= points; index++ {
		angle := 2 * math.Pi * float64(index) / float64(points)
		scale := radius
		if jitter != nil {
			scale *= 0.85 + 0.3*jitter.Float64()
		}
		ring = append(ring, geodata.GeoPoint{
			Longitude: center.Longitude + scale*math.Cos(angle),
			Latitude:  center.Latitude + scale*math.Sin(angle),
		})
	}
	return ring
}

func differentialCopyPolygons(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	result := make([][]geodata.GeoPoint, len(polygons))
	for index, polygon := range polygons {
		result[index] = append([]geodata.GeoPoint(nil), polygon...)
	}
	return result
}

// 包围盒剪枝只跳过所有顶点对都超过 60 km 的组合，必须与全扫参考实现逐点一致。
func TestMergeTouchingVisiblePolygonsMatchesReference(t *testing.T) {
	random := rand.New(rand.NewSource(20260915))
	bridged := 0
	for test := 0; test < 24; test++ {
		polygons := make([][]geodata.GeoPoint, 0, 12)
		longitude := -40.0
		latitude := 20 + 30*random.Float64()
		for index := 0; index < 8+random.Intn(5); index++ {
			radius := 0.1 + 0.4*random.Float64()
			polygons = append(polygons, differentialRing(
				geodata.GeoPoint{Longitude: longitude, Latitude: latitude + 0.5*random.Float64()},
				radius, 6+random.Intn(12), random))
			longitude += radius + 0.4 + random.Float64()
		}
		got := mergeTouchingVisiblePolygons(differentialCopyPolygons(polygons))
		want := mergeTouchingVisiblePolygonsReference(differentialCopyPolygons(polygons))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("case %d: box pruning changed the merge result: got=%d want=%d",
				test, len(got), len(want))
		}
		if len(got) < len(polygons) {
			bridged++
		}
	}
	if bridged == 0 {
		t.Fatal("random inputs never exercised the bridging path")
	}
}

// 最大分量显式保留不得改变任何输入的过滤结果。
func TestRemoveTinyPolygonComponentsMatchesReference(t *testing.T) {
	random := rand.New(rand.NewSource(4242))
	for test := 0; test < 32; test++ {
		count := 1 + random.Intn(6)
		polygons := make([][]geodata.GeoPoint, 0, count)
		for index := 0; index < count; index++ {
			scale := math.Pow(10, -6*random.Float64())
			polygons = append(polygons, differentialRing(
				geodata.GeoPoint{
					Longitude: -180 + 360*random.Float64(),
					Latitude:  -80 + 160*random.Float64(),
				},
				scale*(0.001+random.Float64()), 3+random.Intn(6), nil))
		}
		if test%4 == 0 {
			polygons = append(polygons,
				[]geodata.GeoPoint{{Longitude: 0, Latitude: 0}},
				[]geodata.GeoPoint{
					{Longitude: 179.9, Latitude: 0}, {Longitude: -179.9, Latitude: 0},
					{Longitude: -179.9, Latitude: 1}, {Longitude: 179.9, Latitude: 1},
				})
		}
		got := RemoveTinyPolygonComponents(differentialCopyPolygons(polygons))
		want := removeTinyPolygonComponentsReference(differentialCopyPolygons(polygons))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("case %d: got %d components, want %d", test, len(got), len(want))
		}
	}
}

// 单侧换支的抑制区间以"下一次任一侧换支"为界，重构不得改变输出。
func TestPairedBoundaryPolygonsMatchesReference(t *testing.T) {
	random := rand.New(rand.NewSource(90210))
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	for test := 0; test < 24; test++ {
		count := 2 + random.Intn(8)
		first := make([]basic.OccultationPathPoint, 0, count)
		second := make([]basic.OccultationPathPoint, 0, count)
		longitude := 0.0
		latitude := -30 + 60*random.Float64()
		for index := 0; index < count; index++ {
			if random.Float64() < 0.25 {
				longitude += 30 + 90*random.Float64()
			} else {
				longitude += 0.2 + random.Float64()
			}
			when := start.Add(time.Duration(index) * time.Minute)
			first = append(first, basic.OccultationPathPoint{
				Time: when, Longitude: longitude, Latitude: latitude,
			})
			second = append(second, basic.OccultationPathPoint{
				Time: when, Longitude: longitude, Latitude: latitude - 1 - random.Float64(),
			})
		}
		got := PairedBoundaryPolygons(first, second)
		want := pairedBoundaryPolygonsReference(first, second)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("case %d: got %d cells, want %d", test, len(got), len(want))
		}
	}
}

// ≤5 km 的同相位分支是冗余早退：两种写法在全部距离区间必须给出相同判定。
func TestSamePhaseHorizonConnectorAllowedMatchesReferenceRule(t *testing.T) {
	reference := func(distanceKM float64) bool {
		if distanceKM <= horizonConnectorSamePhaseEndpointDistanceKM {
			return true
		}
		return distanceKM <= horizonConnectorSamePhaseBranchDistanceKM
	}
	for _, distanceKM := range []float64{0, 0.5, 5, 5.001, 120, 1199.9, 1200, 1200.1, 5000} {
		first := horizonConnectorEndpoint{
			point: basic.OccultationPathPoint{}, curveIndex: 3, segmentIndex: 0, atStart: true,
		}
		last := horizonConnectorEndpoint{
			point: basic.OccultationPathPoint{
				Longitude: distanceKM / (EarthRadiusKM * math.Pi / 180),
			},
			curveIndex: 3, segmentIndex: 1, atStart: true,
		}
		measured := DistanceKM(first.point, last.point)
		if math.Abs(measured-distanceKM) > 0.02*math.Max(1, distanceKM) {
			t.Fatalf("test distance %.3f km measured as %.3f km", distanceKM, measured)
		}
		if got, want := samePhaseHorizonConnectorAllowed(first, last), reference(distanceKM); got != want {
			t.Fatalf("distance=%.3f km: got=%v want=%v", distanceKM, got, want)
		}
	}
}

// 相位环候选顺序会被 map 迭代打乱，并集必须与顺序无关。
func TestUnionPolygonsPermutationInvariant(t *testing.T) {
	random := rand.New(rand.NewSource(777))
	for test := 0; test < 8; test++ {
		polygons := make([][]geodata.GeoPoint, 0, 6)
		for index := 0; index < 6; index++ {
			polygons = append(polygons, differentialRing(geodata.GeoPoint{
				Longitude: -20 + 15*random.Float64(),
				Latitude:  -10 + 20*random.Float64(),
			}, 2+3*random.Float64(), 8+random.Intn(8), random))
		}
		base, err := geodata.UnionPolygons(differentialCopyPolygons(polygons))
		if err != nil {
			t.Fatalf("case %d: %v", test, err)
		}
		for permutation := 0; permutation < 6; permutation++ {
			shuffled := differentialCopyPolygons(polygons)
			random.Shuffle(len(shuffled), func(first, second int) {
				shuffled[first], shuffled[second] = shuffled[second], shuffled[first]
			})
			other, err := geodata.UnionPolygons(shuffled)
			if err != nil {
				t.Fatalf("case %d permutation %d: %v", test, permutation, err)
			}
			if len(other) != len(base) {
				t.Fatalf("case %d permutation %d: components %d vs %d", test, permutation, len(other), len(base))
			}
			miss := math.Max(
				geodata.SphericalPolygonsPathMissDistanceKM(base, other, true),
				geodata.SphericalPolygonsPathMissDistanceKM(other, base, true),
			)
			if miss > 1e-9 {
				t.Fatalf("case %d permutation %d: unions differ by %.6f km", test, permutation, miss)
			}
		}
	}
}

func differentialOpenFootprintRun(centers []geodata.GeoPoint, halfHeight float64, points int) []basic.OccultationFootprint {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	footprints := make([]basic.OccultationFootprint, 0, len(centers))
	for index, center := range centers {
		when := start.Add(time.Duration(index) * 5 * time.Minute)
		boundary := make([]basic.OccultationPathPoint, 0, points)
		for step := 0; step < points; step++ {
			fraction := float64(step) / float64(points-1)
			boundary = append(boundary, basic.OccultationPathPoint{
				Time:      when,
				Longitude: center.Longitude + 0.02*math.Sin(2*math.Pi*fraction),
				Latitude:  center.Latitude - halfHeight + 2*halfHeight*fraction,
			})
		}
		footprints = append(footprints, basic.OccultationFootprint{
			Time:       when,
			Boundaries: [][]basic.OccultationPathPoint{boundary},
		})
	}
	return footprints
}

func differentialPolygonMiss(first, second [][]geodata.GeoPoint) float64 {
	return math.Max(
		geodata.SphericalPolygonsPathMissDistanceKM(first, second, true),
		geodata.SphericalPolygonsPathMissDistanceKM(second, first, true),
	)
}

// 单调端点轨迹轮廓与逐条 ribbon 球面并集必须给出同一外包络。
func TestOpenSweepMonotoneMatchesRibbonUnion(t *testing.T) {
	random := rand.New(rand.NewSource(31337))
	for test := 0; test < 6; test++ {
		centers := make([]geodata.GeoPoint, 0, 9)
		longitude := -20 + 10*random.Float64()
		latitude := 30 + 20*random.Float64()
		for index := 0; index < 9; index++ {
			centers = append(centers, geodata.GeoPoint{Longitude: longitude, Latitude: latitude})
			longitude += 0.4 + random.Float64()
			latitude += 0.2*random.Float64() - 0.1
		}
		footprints := differentialOpenFootprintRun(centers, 0.8, 9)
		got, err := footprintOpenSweepPolygons(footprints)
		if err != nil {
			t.Fatalf("case %d: %v", test, err)
		}
		want, err := footprintOpenSweepPolygonsWithTransitionsReference(footprints, true)
		if err != nil {
			t.Fatalf("case %d reference: %v", test, err)
		}
		if len(got) != len(want) {
			t.Fatalf("case %d: components %d vs reference %d", test, len(got), len(want))
		}
		if miss := differentialPolygonMiss(got, want); miss > 1 {
			t.Fatalf("case %d: monotone outline differs from ribbon union by %.3f km", test, miss)
		}
	}
}

// 真实极区事件的单调轮廓与通用并集也必须落在同一包络内。
func TestOpenSweepMonotoneMatchesRibbonUnionSaturn20250105(t *testing.T) {
	path := saturn20250105Path(t)
	sweep, err := footprintOpenSweepPolygons(path.PartialBandFootprints)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := footprintOpenSweepPolygonsWithTransitionsReference(path.PartialBandFootprints, true)
	if err != nil {
		t.Fatal(err)
	}
	miss := differentialPolygonMiss(sweep, reference)
	t.Logf("Saturn 2025-01-05 partial open sweep: monotone=%d reference=%d mutual miss=%.3f km",
		len(sweep), len(reference), miss)
	if miss > 25 {
		t.Fatalf("monotone outline leaves the ribbon union by %.3f km", miss)
	}
}

// 包含索引复用只在顶点修复路径上与参考实现一致；逐边修复路径允许保留更好的结果。
func TestConstrainPolygonsWithinMatchesReference(t *testing.T) {
	random := rand.New(rand.NewSource(5150))
	for test := 0; test < 12; test++ {
		parent := [][]geodata.GeoPoint{
			differentialRing(geodata.GeoPoint{Longitude: 0, Latitude: 0}, 5, 16, nil),
		}
		childRing := differentialRing(geodata.GeoPoint{Longitude: 0, Latitude: 0}, 5, 16, nil)
		for index := range childRing {
			childRing[index].Longitude += 0.005 * (random.Float64() - 0.5)
			childRing[index].Latitude += 0.005 * (random.Float64() - 0.5)
		}
		child := [][]geodata.GeoPoint{childRing}
		if miss := geodata.SphericalPolygonsPathMissDistanceKM(parent, child, true); miss > 3 {
			continue
		}
		got := ConstrainPolygonsWithin(parent, differentialCopyPolygons(child))
		want := constrainPolygonsWithinReference(parent, differentialCopyPolygons(child))
		if reflect.DeepEqual(got, want) {
			continue
		}
		gotMiss := geodata.SphericalPolygonsPathMissDistanceKM(parent, got, true)
		wantMiss := geodata.SphericalPolygonsPathMissDistanceKM(parent, want, true)
		if gotMiss > wantMiss {
			t.Fatalf("case %d: repaired miss %.3f km worse than reference %.3f km", test, gotMiss, wantMiss)
		}
	}
}
