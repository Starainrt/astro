package svg

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

// 本节钉住写出前折线简化的契约：简化几何与原几何逐点偏差不超过容差，环数不变，
// 文档字节与顶点数显著下降；渲染速度只由同文件的基准度量，单元测试不做墙钟断言。

func TestSimplifyOccultationPolylineCollapsesCollinearRun(t *testing.T) {
	points := make([]occultationScreenPoint, 0, 101)
	for index := 0; index <= 100; index++ {
		points = append(points, occultationScreenPoint{x: float64(index), y: 0})
	}
	simplified := simplifyOccultationPolyline(points, 0.4)
	if len(simplified) != 2 {
		t.Fatalf("collinear run simplified to %d points, want 2", len(simplified))
	}
	if simplified[0] != points[0] || simplified[1] != points[len(points)-1] {
		t.Fatal("polyline simplification dropped an endpoint")
	}
}

func TestSimplifyOccultationRingKeepsClosureAndAnchors(t *testing.T) {
	ring := make([]occultationScreenPoint, 0, 73)
	for index := 0; index < 72; index++ {
		angle := 2 * math.Pi * float64(index) / 72
		ring = append(ring, occultationScreenPoint{x: 200 + 80*math.Cos(angle), y: 200 + 80*math.Sin(angle)})
	}
	for _, input := range [][]occultationScreenPoint{ring, append(append([]occultationScreenPoint(nil), ring...), ring[0])} {
		simplified := simplifyOccultationRing(input, 0.4)
		if len(simplified) >= len(input) {
			t.Fatalf("circle simplification kept %d of %d vertices", len(simplified), len(input))
		}
		if simplified[0] != input[0] {
			t.Fatal("ring simplification dropped the first vertex")
		}
		if closed := input[0] == input[len(input)-1]; closed != (simplified[0] == simplified[len(simplified)-1]) {
			t.Fatalf("ring closure changed: input closed=%v output closed=%v", closed, simplified[0] == simplified[len(simplified)-1])
		}
		if deviation := occultationRingMaxDeviation([][]occultationScreenPoint{input}, [][]occultationScreenPoint{simplified}); deviation > 0.4+1e-9 {
			t.Fatalf("circle simplification deviation = %.6f px, want <= 0.4", deviation)
		}
	}
}

func TestSimplifyOccultationRingKeepsDegenerateRing(t *testing.T) {
	ring := []occultationScreenPoint{{x: 10, y: 10}, {x: 10.05, y: 10}, {x: 10, y: 10.05}}
	simplified := simplifyOccultationRing(ring, 0.4)
	if len(simplified) != len(ring) {
		t.Fatalf("sub-tolerance ring simplified to %d points, want %d", len(simplified), len(ring))
	}
}

func TestStarOccultationBandSimplificationWithinTolerance(t *testing.T) {
	path := starOccultationSimplificationPath(t)
	raw := occultationRenderWithTolerance(t, path, 0)
	simplified := occultationRenderWithTolerance(t, path, occultationGeometryTolerancePixels)
	rawRings := occultationSVGPathRings(t, raw, "occultation-band")
	rings := occultationSVGPathRings(t, simplified, "occultation-band")
	if len(rawRings) == 0 {
		t.Fatal("stellar reference document has no occultation-band rings")
	}
	if len(rawRings) != len(rings) {
		t.Fatalf("band ring count changed from %d to %d", len(rawRings), len(rings))
	}
	deviation := occultationRingMaxDeviation(rawRings, rings)
	if deviation > occultationGeometryTolerancePixels+0.01 {
		t.Fatalf("band deviation = %.4f px, tolerance = %.4f px", deviation, occultationGeometryTolerancePixels)
	}
}

func TestPlanetOccultationBandSimplificationWithinTolerance(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	render := func(tolerance float64) string {
		t.Helper()
		previous := occultationGeometryTolerancePixels
		occultationGeometryTolerancePixels = tolerance
		defer func() { occultationGeometryTolerancePixels = previous }()
		return occultationRender(t, func() (string, error) {
			return PlanetOccultationPathSVG(paths[0], PlanetOccultationSVGOptions{Width: 1200, Height: 800})
		})
	}
	raw := render(0)
	simplified := render(occultationGeometryTolerancePixels)
	rawVertices, simplifiedVertices := 0, 0
	for _, className := range []string{"occultation-band", "total-occultation-band"} {
		rawRings := occultationSVGPathRings(t, raw, className)
		rings := occultationSVGPathRings(t, simplified, className)
		if len(rawRings) == 0 || len(rawRings) != len(rings) {
			t.Fatalf("%s ring count changed from %d to %d", className, len(rawRings), len(rings))
		}
		if deviation := occultationRingMaxDeviation(rawRings, rings); deviation > occultationGeometryTolerancePixels+0.01 {
			t.Fatalf("%s deviation = %.4f px, tolerance = %.4f px", className, deviation, occultationGeometryTolerancePixels)
		}
		rawVertices += occultationPathVertexCount(planetOccultationSVGPathData(t, raw, className))
		simplifiedVertices += occultationPathVertexCount(planetOccultationSVGPathData(t, simplified, className))
	}
	if len(simplified)*4 > len(raw)*3 {
		t.Fatalf("planetary document bytes %d -> %d, want at least a 25%% reduction", len(raw), len(simplified))
	}
	if simplifiedVertices*100 > rawVertices*40 {
		t.Fatalf("planetary band vertices %d -> %d, want at least a 60%% reduction", rawVertices, simplifiedVertices)
	}
	t.Logf("planetary document %d -> %d bytes, band vertices %d -> %d",
		len(raw), len(simplified), rawVertices, simplifiedVertices)
}

func TestStarOccultationSVGSimplificationShrinksOutput(t *testing.T) {
	path := starOccultationSimplificationPath(t)
	raw := occultationRenderWithTolerance(t, path, 0)
	simplified := occultationRenderWithTolerance(t, path, occultationGeometryTolerancePixels)
	rawVertices := occultationBandVertexCount(t, raw)
	simplifiedVertices := occultationBandVertexCount(t, simplified)
	if rawVertices == 0 || simplifiedVertices == 0 {
		t.Fatalf("band vertex count raw=%d simplified=%d", rawVertices, simplifiedVertices)
	}
	if len(simplified)*4 > len(raw)*3 {
		t.Fatalf("document bytes %d -> %d, want at least a 25%% reduction", len(raw), len(simplified))
	}
	if simplifiedVertices*100 > rawVertices*35 {
		t.Fatalf("band vertices %d -> %d, want at least a 65%% reduction", rawVertices, simplifiedVertices)
	}
	t.Logf("document %d -> %d bytes, band vertices %d -> %d",
		len(raw), len(simplified), rawVertices, simplifiedVertices)
}

func TestStarOccultationBandSimplificationPreservesRingTopology(t *testing.T) {
	path := starOccultationSimplificationPath(t)
	raw := occultationRenderWithTolerance(t, path, 0)
	simplified := occultationRenderWithTolerance(t, path, occultationGeometryTolerancePixels)
	rawRings := occultationSVGPathRings(t, raw, "occultation-band")
	rings := occultationSVGPathRings(t, simplified, "occultation-band")
	if len(rawRings) != len(rings) {
		t.Fatalf("band ring count %d -> %d, want unchanged", len(rawRings), len(rings))
	}
	for index := range rawRings {
		if len(rings[index]) < 3 {
			t.Fatalf("ring %d has %d vertices after simplification, want at least 3", index, len(rings[index]))
		}
		if len(rings[index]) > len(rawRings[index]) {
			t.Fatalf("ring %d grew from %d to %d vertices", index, len(rawRings[index]), len(rings[index]))
		}
		if math.Signbit(occultationRingSignedArea(rawRings[index])) != math.Signbit(occultationRingSignedArea(rings[index])) {
			t.Fatalf("ring %d winding changed by simplification", index)
		}
		for _, point := range rings[index] {
			if !occultationRingContainsVertex(rawRings[index], point) {
				t.Fatalf("ring %d vertex %.3f,%.3f is not an original vertex", index, point.x, point.y)
			}
		}
	}
}

func occultationRingSignedArea(ring []occultationScreenPoint) float64 {
	area := 0.0
	for index, point := range ring {
		next := ring[(index+1)%len(ring)]
		area += point.x*next.y - next.x*point.y
	}
	return area / 2
}

func occultationRingContainsVertex(ring []occultationScreenPoint, target occultationScreenPoint) bool {
	for _, point := range ring {
		if point == target {
			return true
		}
	}
	return false
}

func starOccultationSimplificationPath(t testing.TB) moon.StarOccultationPath {
	t.Helper()
	location := time.FixedZone("CST", 8*3600)
	paths, err := moon.FindStarOccultationPaths(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		hr4799StarCoordinate(),
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	return paths[0]
}

func occultationRenderWithTolerance(
	t testing.TB,
	path moon.StarOccultationPath,
	tolerance float64,
) string {
	t.Helper()
	previous := occultationGeometryTolerancePixels
	occultationGeometryTolerancePixels = tolerance
	defer func() { occultationGeometryTolerancePixels = previous }()
	return occultationRender(t, func() (string, error) {
		return StarOccultationPathSVG(path, StarOccultationSVGOptions{Width: 720, Height: 520})
	})
}

func occultationRender(t testing.TB, render func() (string, error)) string {
	t.Helper()
	diagram, err := render()
	if err != nil {
		t.Fatalf("occultation render: %v", err)
	}
	return diagram
}

func occultationBandVertexCount(t testing.TB, diagram string) int {
	t.Helper()
	return occultationPathVertexCount(planetOccultationSVGPathData(t, diagram, "occultation-band"))
}

func occultationPathVertexCount(pathData string) int {
	count := 0
	for _, field := range strings.Fields(pathData) {
		if field == "M" || field == "L" {
			count++
		}
	}
	return count
}

func occultationSVGPathRings(t testing.TB, diagram, className string) [][]occultationScreenPoint {
	t.Helper()
	fields := strings.Fields(planetOccultationSVGPathData(t, diagram, className))
	rings := make([][]occultationScreenPoint, 0, 8)
	for index := 0; index < len(fields); {
		switch fields[index] {
		case "M":
			rings = append(rings, nil)
			fallthrough
		case "L":
			if index+2 >= len(fields) {
				t.Fatalf("truncated path data in %q", className)
			}
			x, xErr := strconv.ParseFloat(fields[index+1], 64)
			y, yErr := strconv.ParseFloat(fields[index+2], 64)
			if xErr != nil || yErr != nil {
				t.Fatalf("parse %q coordinate %q/%q: %v %v", className, fields[index+1], fields[index+2], xErr, yErr)
			}
			rings[len(rings)-1] = append(rings[len(rings)-1], occultationScreenPoint{x: x, y: y})
			index += 3
		case "Z":
			index++
		default:
			t.Fatalf("unexpected path command %q in %q", fields[index], className)
		}
	}
	return rings
}

func occultationRingMaxDeviation(raw, simplified [][]occultationScreenPoint) float64 {
	maximum := 0.0
	for index := range raw {
		for _, point := range raw[index] {
			maximum = math.Max(maximum, occultationRingDeviation(point, simplified[index]))
		}
	}
	return maximum
}

func occultationRingDeviation(point occultationScreenPoint, ring []occultationScreenPoint) float64 {
	minimum := math.Inf(1)
	for index := range ring {
		minimum = math.Min(minimum, occultationSegmentDistance(point, ring[index], ring[(index+1)%len(ring)]))
	}
	return minimum
}

func occultationSegmentDistance(point, start, end occultationScreenPoint) float64 {
	dx, dy := end.x-start.x, end.y-start.y
	lengthSquared := dx*dx + dy*dy
	position := 0.0
	if lengthSquared > 0 {
		position = ((point.x-start.x)*dx + (point.y-start.y)*dy) / lengthSquared
	}
	if position < 0 {
		position = 0
	} else if position > 1 {
		position = 1
	}
	return math.Hypot(point.x-(start.x+position*dx), point.y-(start.y+position*dy))
}
