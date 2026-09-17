package geojson_test

// greatest-time-line 逐支路导出，time 与 jde 取自等时线本身，几何与支路逐点一致。

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func cloneSolarGreatestTimeContours(
	source []eclipse.SolarEclipseGreatestTimeContour,
) []eclipse.SolarEclipseGreatestTimeContour {
	result := make([]eclipse.SolarEclipseGreatestTimeContour, len(source))
	for index, contour := range source {
		segments := make([][]eclipse.SolarEclipsePathPoint, len(contour.Segments))
		for segmentIndex, segment := range contour.Segments {
			segments[segmentIndex] = append([]eclipse.SolarEclipsePathPoint(nil), segment...)
		}
		contour.Segments = segments
		result[index] = contour
	}
	return result
}

func TestMarshalSolarEclipseGreatestTimeLines(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	stepped := eclipse.SolarEclipsePartialFootprintOptions{
		Step: 20 * time.Minute, BoundaryPoints: 24, DisableRiseSet: true,
		GreatestTimeStep: 30 * time.Minute,
	}
	valued := stepped
	valued.GreatestTimeStep = 0
	valued.GreatestTimeValues = []time.Time{
		time.Date(2009, time.July, 22, 9, 0, 0, 0, location),
		time.Date(2009, time.July, 22, 10, 0, 0, 0, location),
		time.Date(2009, time.July, 22, 11, 0, 0, 0, location),
	}
	for _, fixture := range []struct {
		name            string
		date            time.Time
		options         eclipse.SolarEclipsePartialFootprintOptions
		central         bool
		wantMultiBranch bool
	}{
		{
			name: "stepped-2021", date: time.Date(2021, time.June, 10, 0, 0, 0, 0, time.UTC),
			options: stepped, wantMultiBranch: true,
		},
		{
			name: "valued-2009-central", date: time.Date(2009, time.July, 22, 0, 0, 0, 0, location),
			options: valued, central: true,
		},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprintsNASABulletinSplitK(fixture.date, fixture.options)
			if !ok || len(partial.GreatestTimeContours) == 0 {
				t.Fatal("expected greatest-time contours")
			}
			multiBranch := false
			for _, contour := range partial.GreatestTimeContours {
				if len(contour.Segments) > 1 {
					multiBranch = true
				}
			}
			if multiBranch != fixture.wantMultiBranch {
				t.Fatalf("fixture multi-branch=%v, want %v", multiBranch, fixture.wantMultiBranch)
			}
			var central *eclipse.SolarEclipsePath
			if fixture.central {
				path, pathOK := eclipse.SolarEclipseCentralPath(
					fixture.date, eclipse.SolarEclipsePathOptions{Step: 10 * time.Minute},
				)
				if !pathOK {
					t.Fatal("expected central path")
				}
				central = &path
			}
			data, err := geojson.MarshalSolarEclipse(partial, central)
			if err != nil {
				t.Fatalf("MarshalSolarEclipse: %v", err)
			}
			lines := featuresWithRole(decodeCollection(t, data), "greatest-time-line")
			branches := 0
			for _, contour := range partial.GreatestTimeContours {
				branches += len(contour.Segments)
			}
			if len(lines) != branches {
				t.Fatalf("greatest-time-line features=%d, branches=%d", len(lines), branches)
			}
			lineIndex := 0
			for _, contour := range partial.GreatestTimeContours {
				for _, segment := range contour.Segments {
					line := lines[lineIndex]
					lineIndex++
					if line.Geometry.Type != "MultiLineString" {
						t.Fatalf("line %d geometry=%q, want MultiLineString", lineIndex-1, line.Geometry.Type)
					}
					if got, want := line.Properties["time"], contour.Time.UTC().Format(time.RFC3339Nano); got != want {
						t.Fatalf("line %d time=%v, want %v", lineIndex-1, got, want)
					}
					if got, ok := line.Properties["jde"].(float64); !ok || got != contour.JDE {
						t.Fatalf("line %d jde=%v, want %v", lineIndex-1, line.Properties["jde"], contour.JDE)
					}
					assertTimedLineAligned(t, line)
					var coordinates [][][]float64
					if err := json.Unmarshal(line.Geometry.Coordinates, &coordinates); err != nil {
						t.Fatalf("decode line %d: %v", lineIndex-1, err)
					}
					if len(coordinates) != 1 || len(coordinates[0]) != len(segment) {
						t.Fatalf("line %d coordinate lines=%d points=%d, branch points=%d",
							lineIndex-1, len(coordinates), coordinateCount(coordinates), len(segment))
					}
					for pointIndex, point := range segment {
						if coordinates[0][pointIndex][0] != point.Longitude ||
							coordinates[0][pointIndex][1] != point.Latitude {
							t.Fatalf("line %d point %d=(%v, %v), branch point=(%v, %v)",
								lineIndex-1, pointIndex, coordinates[0][pointIndex][0],
								coordinates[0][pointIndex][1], point.Longitude, point.Latitude)
						}
					}
				}
			}
		})
	}
}

func coordinateCount(coordinates [][][]float64) int {
	count := 0
	for _, line := range coordinates {
		count += len(line)
	}
	return count
}

func TestMarshalSolarEclipseRejectsInvalidGreatestTimeContours(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	base, ok := eclipse.SolarEclipsePartialFootprintsNASABulletinSplitK(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 20 * time.Minute, BoundaryPoints: 24, DisableRiseSet: true,
		GreatestTimeValues: []time.Time{
			time.Date(2009, time.July, 22, 1, 0, 0, 0, time.UTC),
			time.Date(2009, time.July, 22, 2, 0, 0, 0, time.UTC),
		},
	})
	if !ok || len(base.GreatestTimeContours) == 0 || len(base.GreatestTimeContours[0].Segments) == 0 {
		t.Fatal("expected greatest-time contours")
	}
	for _, testCase := range []struct {
		name   string
		mutate func([]eclipse.SolarEclipseGreatestTimeContour)
	}{
		{name: "nan-longitude", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].Segments[0][0].Longitude = math.NaN()
		}},
		{name: "infinite-latitude", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].Segments[0][0].Latitude = math.Inf(1)
		}},
		{name: "longitude-out-of-range", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].Segments[0][0].Longitude = 180.5
		}},
		{name: "latitude-out-of-range", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].Segments[0][0].Latitude = -90.5
		}},
		{name: "single-point-branch", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].Segments[0] = contours[0].Segments[0][:1]
		}},
		{name: "zero-jde", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].JDE = 0
		}},
		{name: "nan-jde", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].JDE = math.NaN()
		}},
		{name: "zero-time", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].Time = time.Time{}
		}},
		{name: "zero-point-time", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].Segments[0][0].Time = time.Time{}
		}},
		{name: "no-branches", mutate: func(contours []eclipse.SolarEclipseGreatestTimeContour) {
			contours[0].Segments = nil
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			partial := base
			partial.GreatestTimeContours = cloneSolarGreatestTimeContours(base.GreatestTimeContours)
			testCase.mutate(partial.GreatestTimeContours)
			if _, err := geojson.MarshalSolarEclipse(partial, nil); err == nil {
				t.Fatal("invalid greatest-time contour was accepted")
			} else if !strings.Contains(err.Error(), "greatest-time contour") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMarshalSolarEclipseWithoutGreatestTimeLinesIsUnchanged(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	options := eclipse.SolarEclipsePartialFootprintOptions{
		Step: 20 * time.Minute, BoundaryPoints: 24, DisableRiseSet: true,
	}
	plain, ok := eclipse.SolarEclipsePartialFootprintsNASABulletinSplitK(date, options)
	if !ok {
		t.Fatal("expected solar partial footprints")
	}
	first, err := geojson.MarshalSolarEclipse(plain, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	second, err := geojson.MarshalSolarEclipse(plain, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("default solar eclipse export is not byte stable")
	}
	if bytes.Contains(first, []byte("greatest-time-line")) {
		t.Fatal("default solar eclipse export contains greatest-time lines")
	}
	requested := options
	requested.GreatestTimeValues = []time.Time{
		time.Date(2009, time.July, 22, 1, 0, 0, 0, time.UTC),
		time.Date(2009, time.July, 22, 2, 0, 0, 0, time.UTC),
	}
	withLines, ok := eclipse.SolarEclipsePartialFootprintsNASABulletinSplitK(date, requested)
	if !ok || len(withLines.GreatestTimeContours) == 0 {
		t.Fatal("expected greatest-time contours")
	}
	cleared := withLines
	cleared.GreatestTimeContours = nil
	third, err := geojson.MarshalSolarEclipse(cleared, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	if !bytes.Equal(first, third) {
		t.Fatal("clearing greatest-time contours changes the export")
	}
}
