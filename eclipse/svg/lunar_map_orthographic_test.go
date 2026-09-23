package svg

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgmap"
)

// 月食正射图的两条契约：
//  1. 视点取食甚时刻的月下点（"食甚可见半球"的边界就是圆盘边缘），否则视点固定在 (0°,0°)，
//     会把半个可见半球让给与本次月食无关的区域；
//  2. 去掉圆盘边缘 2 px 内的固有亚像素带后，四类分区的着色必须与三刻地平几何一致。
func TestLunarEclipseMapSVGOrthographicCentering(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2029, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		time.Date(2011, 6, 15, 12, 0, 0, 0, time.UTC),
	} {
		// 本用例核对投影几何与四类可见区的物理判据，半影带另有覆盖性用例。
		doc, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
			Width: 960, Height: 640, Language: "zh", Location: time.UTC,
			Projection: EclipseMapProjectionOrthographic, DisablePenumbralPhase: true,
		})
		if !ok {
			t.Fatalf("%s: missing orthographic map", date.Format("2006-01-02"))
		}
		centreX, centreY, radius := lunarOrthographicDisc(t, doc)
		partition := lunarMapPartitionOf(t, doc)
		// 食甚可见半球的边界是视界大圆：它的投影环应当以圆盘中心为圆心。
		ringCentroidX, ringCentroidY, ok := lunarOrthographicRingCentroid(partition, "lunar-visible-maximum-shape")
		if !ok {
			t.Fatalf("%s: no maximum-visible ring", date.Format("2006-01-02"))
		}
		if offset := math.Hypot(ringCentroidX-centreX, ringCentroidY-centreY); offset > 1.5 {
			t.Errorf("%s: greatest-visible ring is %.3f px off the disc centre", date.Format("2006-01-02"), offset)
		}
		if err := lunarOrthographicInteriorMatchesPhysics(t, date, doc, partition, centreX, centreY, radius); err != nil {
			t.Errorf("%s: %v", date.Format("2006-01-02"), err)
		}
	}
}

// lunarOrthographicDisc 从图内读出圆盘圆心与半径。
func lunarOrthographicDisc(t *testing.T, doc string) (float64, float64, float64) {
	t.Helper()
	match := regexp.MustCompile(`<circle class="map-frame" cx="([-0-9.]+)" cy="([-0-9.]+)" r="([-0-9.]+)"`).FindStringSubmatch(doc)
	if match == nil {
		t.Fatal("disc frame not found")
	}
	x, _ := strconv.ParseFloat(match[1], 64)
	y, _ := strconv.ParseFloat(match[2], 64)
	radius, _ := strconv.ParseFloat(match[3], 64)
	return x, y, radius
}

// lunarOrthographicRingCentroid 给出投影环顶点的均值，用作"该环是否为以盘心为圆心的圆"的判据。
func lunarOrthographicRingCentroid(partition lunarPartition, shape string) (float64, float64, bool) {
	rings := partition.shapes[shape]
	if len(rings) == 0 || len(rings[0]) == 0 {
		return 0, 0, false
	}
	sumX, sumY := 0.0, 0.0
	for _, vertex := range rings[0] {
		sumX += vertex[0]
		sumY += vertex[1]
	}
	return sumX / float64(len(rings[0])), sumY / float64(len(rings[0])), true
}

// lunarOrthographicInteriorMatchesPhysics 逐格比对物理分类与着色，只豁免盘缘 2 px 内的固有亚像素带：
// 越靠近视界，地平线在屏幕上越被压扁，整条边界都缩进不到一个像素，任何多边形表示在那里都无分辨率可言。
func lunarOrthographicInteriorMatchesPhysics(
	t *testing.T,
	date time.Time,
	doc string,
	partition lunarPartition,
	centreX, centreY, radius float64,
) error {
	t.Helper()
	info, ok := eclipsecore.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing eclipse")
	}
	frame := eclipseMapFrame(960, 640, svgmap.ProjectionOrthographic, 142, 92)
	frame.CenterLongitude, frame.CenterLatitude = lunarEclipseOrthographicCenter(info.Maximum)
	counts := map[string]int{}
	mismatches := 0
	var first string
	for longitude := -179.5; longitude < 180; longitude += 1 {
		for latitude := -89.5; latitude < 90; latitude += 1 {
			x, y, projectable := frame.Project(longitude, latitude)
			if !projectable {
				continue
			}
			if math.Hypot(x-centreX, y-centreY) > radius-2 {
				continue
			}
			h1 := lunarPartitionAltitude(info.PenumbralStart, longitude, latitude)
			hm := lunarPartitionAltitude(info.Maximum, longitude, latitude)
			h4 := lunarPartitionAltitude(info.PenumbralEnd, longitude, latitude)
			if math.Abs(h1) < 0.05 || math.Abs(hm) < 0.05 || math.Abs(h4) < 0.05 {
				continue
			}
			expected := lunarOrthographicExpectedClass(h1, hm, h4)
			category, _ := partition.painted(x, y)
			counts[expected]++
			if category != expected {
				mismatches++
				if first == "" {
					first = fmt.Sprintf("lon=%s lat=%s h=(%s,%s,%s) expected=%s got=%s",
						lunarPartitionFormat(longitude), lunarPartitionFormat(latitude),
						lunarPartitionFormat(h1), lunarPartitionFormat(hm), lunarPartitionFormat(h4),
						expected, category)
				}
			}
		}
	}
	for _, class := range []string{lunarPartitionCategoryEntire, lunarPartitionCategoryMoonset, lunarPartitionCategoryMoonrise} {
		if counts[class] == 0 {
			t.Errorf("orthographic disc has no %s cell away from the rim", class)
		}
	}
	if mismatches != 0 {
		return fmt.Errorf("%d interior cells carry the wrong class, first %s", mismatches, first)
	}
	return nil
}

func lunarOrthographicExpectedClass(h1, hm, h4 float64) string {
	switch {
	case h1 > 0 && hm > 0 && h4 > 0:
		return lunarPartitionCategoryEntire
	case h1 > 0 && h4 > 0:
		return lunarPartitionCategoryMoonset
	case h1 > 0:
		return lunarPartitionCategoryMoonset
	case h4 > 0:
		return lunarPartitionCategoryMoonrise
	case hm > 0:
		return lunarPartitionCategoryMoonset
	}
	return lunarPartitionCategoryUnavailable
}
