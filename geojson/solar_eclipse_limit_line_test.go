package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// 本文件钉住限线口径：north-limit / south-limit 的标签由每个顶点自身的投影侧决定，
// times 是该顶点在中心线上的投影时刻，不按顶点顺序均匀铺开。
const limitSideSlackKM = 25.0

type derivedLimitLine struct {
	coordinates [][2]float64
	times       []time.Time
}

func limitSolarEclipseFixture(
	t *testing.T,
	text string,
) ([]byte, eclipse.SolarEclipsePath) {
	t.Helper()
	date := grazingBandDate(t, text)
	info, ok := eclipse.SolarEclipseOnDate(date)
	if !ok || !info.HasCentral {
		t.Fatalf("expected a central solar eclipse on %s", text)
	}
	partialOptions, pathOptions := grazingBandOptions(info, "overview")
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, partialOptions)
	if !ok {
		t.Fatal("missing partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, pathOptions)
	if !ok {
		t.Fatal("missing central path")
	}
	raw, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	return raw, central
}

func derivedLimitLineFrom(t *testing.T, raw []byte, role string) derivedLimitLine {
	t.Helper()
	feature := featureWithRole(t, decodeCollection(t, raw), role)
	var line [][2]float64
	if err := json.Unmarshal(feature.Geometry.Coordinates, &line); err != nil {
		var lines [][][2]float64
		if multiErr := json.Unmarshal(feature.Geometry.Coordinates, &lines); multiErr != nil {
			t.Fatalf("%s coordinates: %v", role, err)
		}
		line = nil
		for index, segment := range lines {
			if index > 0 && !segmentsMeetAtSeam(lines[index-1], segment) {
				t.Fatalf("%s is split into %d segments", role, len(lines))
			}
			line = append(line, segment...)
		}
	}
	encoded, ok := feature.Properties["times"].([]interface{})
	if !ok {
		t.Fatalf("%s has no times array", role)
	}
	if len(encoded) > 0 {
		if _, nestedOK := encoded[0].([]interface{}); nestedOK {
			var flat []interface{}
			for _, segment := range encoded {
				nested, _ := segment.([]interface{})
				flat = append(flat, nested...)
			}
			encoded = flat
		}
	}
	if len(encoded) != len(line) {
		t.Fatalf("%s times=%d coordinates=%d, want one time per vertex", role, len(encoded), len(line))
	}
	times := make([]time.Time, len(encoded))
	for index, value := range encoded {
		text, _ := value.(string)
		stamp, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			t.Fatalf("%s time %d: %v", role, index, err)
		}
		times[index] = stamp
	}
	return derivedLimitLine{coordinates: line, times: times}
}

// segmentsMeetAtSeam 报告前段终点与后段起点是否为 ±180° 上的同一点。
func segmentsMeetAtSeam(first, second [][2]float64) bool {
	if len(first) == 0 || len(second) == 0 {
		return false
	}
	end, start := first[len(first)-1], second[0]
	return math.Abs(end[0]) == 180 && end[0] == -start[0] && math.Abs(end[1]-start[1]) < 1e-9
}

// projectOnCenterLine 独立复算顶点在中心线上的投影纬度与插值时刻。
func projectOnCenterLine(
	point [2]float64,
	centerLine []eclipse.SolarEclipsePathPoint,
) (float64, time.Time) {
	bestDistance := math.Inf(1)
	bestLatitude := centerLine[0].Latitude
	bestTime := centerLine[0].Time
	scale := math.Cos(point[1] * math.Pi / 180)
	for index := 0; index+1 < len(centerLine); index++ {
		first, second := centerLine[index], centerLine[index+1]
		ax := math.Remainder(first.Longitude-point[0], 360) * scale
		ay := first.Latitude - point[1]
		bx := math.Remainder(second.Longitude-point[0], 360) * scale
		by := second.Latitude - point[1]
		dx, dy := bx-ax, by-ay
		length := dx*dx + dy*dy
		fraction := 0.0
		if length > 0 {
			fraction = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/length))
		}
		distance := math.Hypot(ax+fraction*dx, ay+fraction*dy)
		if distance >= bestDistance {
			continue
		}
		bestDistance = distance
		bestLatitude = first.Latitude + fraction*(second.Latitude-first.Latitude)
		bestTime = first.Time.Add(time.Duration(float64(second.Time.Sub(first.Time)) * fraction))
	}
	return bestLatitude, bestTime
}

func TestSolarEclipseLimitSidesCarryTheirOwnProjectionSide(t *testing.T) {
	for _, text := range []string{"2003-05-31", "1874-10-10", "2185-07-26", "4862-09-28"} {
		t.Run(text, func(t *testing.T) {
			raw, central := limitSolarEclipseFixture(t, text)
			for _, side := range []struct {
				role  string
				north bool
			}{{"north-limit", true}, {"south-limit", false}} {
				line := derivedLimitLineFrom(t, raw, side.role)
				for index, point := range line.coordinates {
					latitude, _ := projectOnCenterLine(point, central.CenterLine)
					offset := (point[1] - latitude) * 111.3
					if side.north && offset < -limitSideSlackKM {
						t.Fatalf("%s vertex %d sits %.1f km south of the center line", side.role, index, -offset)
					}
					if !side.north && offset > limitSideSlackKM {
						t.Fatalf("%s vertex %d sits %.1f km north of the center line", side.role, index, offset)
					}
				}
			}
		})
	}
}

// bandRingVertices 收集 central-band 环上的顶点，用于判定限线是否由带边界派生。
func bandRingVertices(t *testing.T, raw []byte) map[[2]float64]bool {
	t.Helper()
	vertices := map[[2]float64]bool{}
	for _, ring := range grazingBandRings(t, raw) {
		for _, point := range ring {
			vertices[[2]float64{point.Longitude, point.Latitude}] = true
		}
	}
	return vertices
}

func limitFollowsBandRing(line derivedLimitLine, vertices map[[2]float64]bool) bool {
	if len(line.coordinates) < 3 || len(vertices) == 0 {
		return false
	}
	matched := 0
	for _, point := range line.coordinates {
		if vertices[point] {
			matched++
		}
	}
	return matched*5 >= len(line.coordinates)*4
}

func TestSolarEclipseDerivedLimitTimesAreProjectionTimes(t *testing.T) {
	for _, text := range []string{"2003-05-31"} {
		t.Run(text, func(t *testing.T) {
			raw, central := limitSolarEclipseFixture(t, text)
			ringVertices := bandRingVertices(t, raw)
			derived := 0
			for _, role := range []string{"north-limit", "south-limit"} {
				line := derivedLimitLineFrom(t, raw, role)
				if !limitFollowsBandRing(line, ringVertices) {
					continue
				}
				derived++
				for index, point := range line.coordinates {
					_, projected := projectOnCenterLine(point, central.CenterLine)
					if delta := line.times[index].Sub(projected); delta > 2*time.Second || delta < -2*time.Second {
						t.Fatalf("%s vertex %d time=%v, projected=%v",
							role, index, line.times[index].UTC(), projected.UTC())
					}
				}
				minimum, maximum := time.Duration(math.MaxInt64), time.Duration(0)
				for index := 1; index < len(line.times); index++ {
					step := line.times[index].Sub(line.times[index-1])
					if step < minimum {
						minimum = step
					}
					if step > maximum {
						maximum = step
					}
				}
				if maximum < 3*minimum {
					t.Fatalf("%s times look uniformly spread: min=%v max=%v", role, minimum, maximum)
				}
			}
			if derived == 0 {
				t.Fatal("no derived limit in the fixture set; the projection-time contract is untested")
			}
		})
	}
}
