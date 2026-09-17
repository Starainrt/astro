package svg

import (
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

// 本文件钉住紧凑掩带合并缓存的契约：同输入复用只读结果、不同几何不会串味、产物逐字节确定。

func occultationTestCompactSweep(t *testing.T) occultationFootprintSweep {
	t.Helper()
	path := occultationTestSaturnPath(t)
	if len(path.PartialBandFootprints) == 0 {
		t.Fatal("fixture carries no compact band footprints")
	}
	return occultationFootprintSweep{
		footprints:         path.PartialBandFootprints,
		contours:           path.PartialBandContours,
		visibilityContours: path.PartialVisibilityContours,
		northern:           path.NorthernLimit,
		southern:           path.SouthernLimit,
		curves:             path.RiseSetCurves,
		compactBand:        true,
		kind:               partialOccultationFootprintSweep,
	}
}

func TestOccultationBandCacheReusesIdenticalGeometry(t *testing.T) {
	sweep := occultationTestCompactSweep(t)
	first, firstAuthoritative, err := sweep.mergedBandPolygons()
	if err != nil {
		t.Fatalf("first mergedBandPolygons() error = %v", err)
	}
	if len(first) == 0 {
		t.Fatal("compact band merge produced no polygon")
	}
	second, secondAuthoritative, err := sweep.mergedBandPolygons()
	if err != nil {
		t.Fatalf("second mergedBandPolygons() error = %v", err)
	}
	if len(second) != len(first) || &second[0][0] != &first[0][0] {
		t.Fatal("identical band geometry was merged again instead of reusing the cached polygons")
	}
	if secondAuthoritative != firstAuthoritative {
		t.Fatalf("cached authoritative flag = %v, want %v", secondAuthoritative, firstAuthoritative)
	}
}

func TestOccultationBandCacheSeparatesDifferentGeometry(t *testing.T) {
	sweep := occultationTestCompactSweep(t)
	first, _, err := sweep.mergedBandPolygons()
	if err != nil {
		t.Fatalf("first mergedBandPolygons() error = %v", err)
	}
	mutated := sweep
	mutated.footprints = append([]moon.PlanetOccultationFootprint(nil), sweep.footprints...)
	last := len(mutated.footprints) - 1
	mutated.footprints[last].Polygons = append([][]moon.OccultationPathPoint(nil), mutated.footprints[last].Polygons...)
	polygon := append([]moon.OccultationPathPoint(nil), mutated.footprints[last].Polygons[0]...)
	polygon[0].Longitude += 1.5
	mutated.footprints[last].Polygons[0] = polygon
	second, _, err := mutated.mergedBandPolygons()
	if err != nil {
		t.Fatalf("mutated mergedBandPolygons() error = %v", err)
	}
	if len(second) > 0 && len(first) > 0 && &second[0][0] == &first[0][0] {
		t.Fatal("a shifted footprint reused the cached band polygons")
	}
}

func TestCompactBandRenderIsDeterministic(t *testing.T) {
	path := occultationTestSaturnPath(t)
	options := StarOccultationSVGOptions{Width: 1200, Height: 800, Location: time.UTC}
	first, err := PlanetOccultationPathSVG(path, options)
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	second, err := PlanetOccultationPathSVG(path, options)
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if first != second {
		t.Fatal("compact-band SVG changed between renders after the merge was cached")
	}
	if len(first) == 0 {
		t.Fatal("compact-band render is empty")
	}
}
