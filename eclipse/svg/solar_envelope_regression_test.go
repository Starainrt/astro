package svg

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgmap"
)

func TestSolarSVGUsesAuthoritativeCentralBand(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*3600)
	for _, fixture := range []struct {
		date     time.Time
		lon, lat float64
	}{
		{time.Date(2012, 5, 21, 0, 0, 0, 0, zone), 120.4913, 27.4779},
		{time.Date(2056, 7, 13, 0, 0, 0, 0, zone), -64.3737, -5.4978},
		{time.Date(2164, 3, 23, 0, 0, 0, 0, zone), -120.575, 48.75},
		{time.Date(2026, 8, 12, 0, 0, 0, 0, zone), -5.991755201, 45.070085755},
	} {
		t.Run(fixture.date.Format("2006-01-02"), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(fixture.date, eclipse.SolarEclipsePartialFootprintOptions{Step: 2 * time.Minute, BoundaryPoints: 96})
			if !ok {
				t.Fatal("missing eclipse")
			}
			central, ok := eclipse.SolarEclipseCentralPath(fixture.date, eclipse.SolarEclipsePathOptions{Step: 2 * time.Minute})
			if !ok {
				t.Fatal("missing central path")
			}
			frame := svgmap.Frame{Width: 36000, Height: 18000, Projection: svgmap.ProjectionEquirectangular}
			var builder strings.Builder
			writeSolarEclipseCentralPath(&builder, central, partial.CentralBandFootprints, partial.CentralBandSegments, frame, partial.Eclipse.Type)
			if !strings.Contains(builder.String(), `class="central-eclipse-band" data-source="besselian-critical-envelope"`) {
				t.Fatal("old central fill used")
			}
			x, y, _ := frame.Project(fixture.lon, fixture.lat)
			inside := false
			for _, ring := range svgClosedPathRings(t, builder.String()) {
				inside = inside || pointInLunarVisibilityPolygon(x, y, ring)
			}
			if !inside {
				t.Fatal("rendered central band omits visible site")
			}
			full, ok := SolarEclipseMapSVG(fixture.date, SolarEclipseMapSVGOptions{Projection: EclipseMapProjectionEquirectangular})
			if !ok || !strings.Contains(full, `class="central-eclipse-band" data-source="besselian-critical-envelope"`) {
				t.Fatal("public SVG does not use core envelope")
			}
		})
	}
}

func svgClosedPathRings(t *testing.T, value string) [][][2]float64 {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(value))
	var rings [][][2]float64
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		tag, ok := token.(xml.StartElement)
		if !ok || tag.Name.Local != "path" {
			continue
		}
		for _, attribute := range tag.Attr {
			if attribute.Name.Local != "d" || !strings.Contains(attribute.Value, "Z") {
				continue
			}
			fields := strings.Fields(attribute.Value)
			var ring [][2]float64
			for index := 0; index < len(fields); {
				if fields[index] == "Z" {
					rings = append(rings, ring)
					ring = nil
					index++
					continue
				}
				if (fields[index] != "M" && fields[index] != "L") || index+2 >= len(fields) {
					t.Fatalf("unexpected path %q", attribute.Value)
				}
				x, err := strconv.ParseFloat(fields[index+1], 64)
				if err != nil {
					t.Fatal(err)
				}
				y, err := strconv.ParseFloat(fields[index+2], 64)
				if err != nil {
					t.Fatal(err)
				}
				ring = append(ring, [2]float64{x, y})
				index += 3
			}
		}
	}
	return rings
}

func TestSolarPolarSVGUsesAuthoritativeCentralBand(t *testing.T) {
	for _, day := range [][3]int{
		{2003, 11, 23}, {2021, 12, 4}, {2039, 12, 15}, {2061, 10, 13}, {2981, 10, 19},
	} {
		date := time.Date(day[0], time.Month(day[1]), day[2], 0, 0, 0, 0, time.UTC)
		t.Run(date.Format("2006-01-02"), func(t *testing.T) {
			for _, projection := range []EclipseMapProjection{EclipseMapProjectionEquirectangular, EclipseMapProjectionSouthPolar} {
				value, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{Projection: projection})
				if !ok || !strings.Contains(value, `class="central-eclipse-band" data-source="besselian-critical-envelope"`) {
					t.Fatalf("projection %s discarded the core central envelope", projection)
				}
			}
		})
	}
}
