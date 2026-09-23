package svg

import (
	"encoding/xml"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
)

// 月食全球可见图的分区契约：四类区域对全球网格穷尽且互斥，"不可见"只覆盖 P1、食甚、P4 三刻都在
// 地平下的地点，"全程可见"恰好等于 P1、食甚、P4 三刻都在地平上的地点。
// Partition contract: the four regions tile the globe exhaustively and exclusively.

const (
	lunarPartitionCategoryEntire      = "entire"
	lunarPartitionCategoryMoonset     = "moonset"
	lunarPartitionCategoryMoonrise    = "moonrise"
	lunarPartitionCategoryUnavailable = "unavailable"
	lunarPartitionCategoryPenumbra    = "penumbra-only"
)

type lunarPartitionLayer struct {
	category string
	shape    string
	masks    []string
	clips    []string
}

type lunarPartition struct {
	frame  svgFrameBox
	shapes map[string][][][2]float64
	masks  map[string][]string
	clips  map[string][]string
	layers []lunarPartitionLayer
}

func lunarMapPartitionOf(t *testing.T, doc string) lunarPartition {
	t.Helper()
	partition := lunarPartition{
		frame:  lunarPenumbralTestFrame(t, doc),
		shapes: map[string][][][2]float64{},
		masks:  map[string][]string{},
		clips:  map[string][]string{},
	}
	type frame struct {
		masks []string
		clips []string
	}
	var stack []frame
	inRegions := false
	var container, containerID string
	decoder := xml.NewDecoder(strings.NewReader(doc))
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch typed := token.(type) {
		case xml.StartElement:
			class := lunarPartitionAttr(typed, "class")
			reference := strings.TrimPrefix(lunarPartitionAttr(typed, "href"), "#")
			mask := lunarPartitionURLID(lunarPartitionAttr(typed, "mask"))
			clip := lunarPartitionURLID(lunarPartitionAttr(typed, "clip-path"))
			switch typed.Name.Local {
			case "path":
				if id := lunarPartitionAttr(typed, "id"); id != "" {
					partition.shapes[id] = lunarPenumbralTestRings(t, lunarPartitionAttr(typed, "d"))
				}
			case "mask", "clipPath":
				container, containerID = typed.Name.Local, lunarPartitionAttr(typed, "id")
			case "use":
				switch container {
				case "mask":
					if reference != "" {
						partition.masks[containerID] = append(partition.masks[containerID], reference)
					}
				case "clipPath":
					if reference != "" {
						partition.clips[containerID] = append(partition.clips[containerID], reference)
					}
				}
			case "rect", "circle":
				if container == "clipPath" && containerID != "" && partition.clips[containerID] == nil {
					partition.clips[containerID] = []string{}
				}
			case "g":
				if strings.Contains(class, "lunar-visibility-regions") {
					inRegions = true
				}
				inherited := frame{}
				if len(stack) > 0 {
					inherited = stack[len(stack)-1]
				}
				if mask != "" {
					inherited.masks = append(append([]string(nil), inherited.masks...), mask)
				}
				if clip != "" {
					inherited.clips = append(append([]string(nil), inherited.clips...), clip)
				}
				stack = append(stack, inherited)
			}
			if inRegions && (typed.Name.Local == "use" || typed.Name.Local == "rect" || typed.Name.Local == "circle") {
				current := frame{}
				if len(stack) > 0 {
					current = stack[len(stack)-1]
				}
				if mask != "" {
					current.masks = append(append([]string(nil), current.masks...), mask)
				}
				if clip != "" {
					current.clips = append(append([]string(nil), current.clips...), clip)
				}
				partition.layers = append(partition.layers, lunarPartitionLayer{
					category: lunarPartitionCategory(class, typed.Name.Local),
					shape:    reference,
					masks:    current.masks,
					clips:    current.clips,
				})
			}
		case xml.EndElement:
			if typed.Name.Local == "mask" || typed.Name.Local == "clipPath" {
				container, containerID = "", ""
			}
			if typed.Name.Local == "g" {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
				if len(stack) == 0 {
					inRegions = false
				}
			}
		}
	}
	if len(partition.layers) == 0 {
		t.Fatal("visibility region layers not found")
	}
	return partition
}

func lunarPartitionAttr(element xml.StartElement, name string) string {
	for _, attribute := range element.Attr {
		if attribute.Name.Local == name {
			return attribute.Value
		}
	}
	return ""
}

func lunarPartitionURLID(value string) string {
	value = strings.TrimSuffix(strings.TrimPrefix(value, "url(#"), ")")
	return value
}

func lunarPartitionCategory(class, element string) string {
	switch {
	case strings.Contains(class, "entire-eclipse-region"):
		return lunarPartitionCategoryEntire
	case strings.Contains(class, "penumbra-only-region"):
		return lunarPartitionCategoryPenumbra
	case strings.Contains(class, "moonset-region"):
		return lunarPartitionCategoryMoonset
	case strings.Contains(class, "moonrise-region"):
		return lunarPartitionCategoryMoonrise
	case strings.Contains(class, "eclipse-unavailable-region"):
		return lunarPartitionCategoryUnavailable
	}
	return element + ":" + class
}

// covers 逐点复算一层是否落在该点：形状、外层 clip-path（多形状取并）与嵌套 mask 都要通过，
// mask 里的黑色形状按挖空处理。
func (partition lunarPartition) covers(layer lunarPartitionLayer, x, y float64) bool {
	if layer.shape != "" && !partition.inside(layer.shape, x, y) {
		return false
	}
	// clipPath 里的多个形状取并集：落在其中任意一个之内就算通过。
	for _, clip := range layer.clips {
		shapes := partition.clips[clip]
		if len(shapes) == 0 {
			continue
		}
		inside := false
		for _, shape := range shapes {
			if partition.inside(shape, x, y) {
				inside = true
				break
			}
		}
		if !inside {
			return false
		}
	}
	for _, maskID := range layer.masks {
		for _, excluded := range partition.masks[maskID] {
			if partition.inside(excluded, x, y) {
				return false
			}
		}
	}
	return true
}

func (partition lunarPartition) inside(shapeID string, x, y float64) bool {
	return lunarPenumbralTestInside(partition.shapes[shapeID], x, y)
}

// painted 返回覆盖该点的最上层区域（SVG 后画的盖住先画的）与覆盖层数。
func (partition lunarPartition) painted(x, y float64) (string, int) {
	category, count := "", 0
	for _, layer := range partition.layers {
		if partition.covers(layer, x, y) {
			category, count = layer.category, count+1
		}
	}
	return category, count
}

// partitionGrid 遍历 1°×1° 全球网格（64800 点，格心取 ±0.5°）。
func (partition lunarPartition) grid(visit func(longitude, latitude float64, x, y float64, category string, covered int)) {
	for longitude := -179.5; longitude < 180; longitude += 1 {
		for latitude := -89.5; latitude < 90; latitude += 1 {
			x, y, projectable := partition.frame.Project(longitude, latitude)
			if !projectable {
				continue
			}
			category, count := partition.painted(x, y)
			visit(longitude, latitude, x, y, category, count)
		}
	}
}

func lunarPartitionAltitude(at time.Time, longitude, latitude float64) float64 {
	return basic.HMoonHeight(basic.Date2JD(at.UTC()), longitude, latitude, 0)
}

func TestLunarEclipseMapSVGVisibilityRegionsTileTheGlobe(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2029, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		time.Date(2011, 6, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2016, 8, 18, 0, 0, 0, 0, time.UTC),
	} {
		// 只检查四类底图的分区：半影带另有覆盖性用例。
		doc, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
			Width: 960, Height: 640, Language: "zh", Location: time.UTC, DisablePenumbralPhase: true})
		if !ok {
			t.Fatalf("%s: missing map", date.Format("2006-01-02"))
		}
		partition := lunarMapPartitionOf(t, doc)
		counts := map[string]int{}
		uncovered, covered, multiple := 0, 64800, 0
		var firstUncovered string
		partition.grid(func(longitude, latitude, x, y float64, category string, layers int) {
			counts[category]++
			if category == "" {
				uncovered++
				if firstUncovered == "" {
					firstUncovered = "(" + lunarPartitionFormat(longitude) + "," + lunarPartitionFormat(latitude) + ")"
				}
				return
			}
			if layers > 1 {
				multiple++
			}
		})
		t.Logf("%s %s: %v", date.Format("2006-01-02"), countsString(counts), "uncovered="+lunarPartitionFormat(float64(uncovered)))
		if uncovered != 0 {
			t.Errorf("%s: %d of %d grid points carry no region colour, first at %s",
				date.Format("2006-01-02"), uncovered, covered, firstUncovered)
		}
		if multiple != 0 {
			t.Errorf("%s: %d grid points are painted by more than one region layer", date.Format("2006-01-02"), multiple)
		}
		for _, category := range []string{lunarPartitionCategoryEntire, lunarPartitionCategoryMoonset,
			lunarPartitionCategoryMoonrise, lunarPartitionCategoryUnavailable} {
			if counts[category] == 0 {
				t.Errorf("%s: region %s covers no grid point", date.Format("2006-01-02"), category)
			}
		}
	}
}

// 打开半影阶段后，"仅见半影"带与带食月出/带食月落必须互斥：两种半透明底色叠在一起会混出图例里
// 没有的颜色，所以底带要被同源的 U1/U4 形状裁掉；同时裁切不能切出没有着色的缝。
func TestLunarEclipseMapSVGPenumbralPhaseKeepsCoverage(t *testing.T) {
	date := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	doc, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
		Width: 960, Height: 640, Language: "zh", Location: time.UTC})
	if !ok {
		t.Fatal("missing penumbral-phase map")
	}
	partition := lunarMapPartitionOf(t, doc)
	uncovered, penumbraOnly, doubleTinted := 0, 0, 0
	partition.grid(func(longitude, latitude, x, y float64, category string, layers int) {
		if category == "" {
			uncovered++
			return
		}
		if category != lunarPartitionCategoryPenumbra {
			return
		}
		penumbraOnly++
		for _, layer := range partition.layers {
			if layer.category == lunarPartitionCategoryPenumbra || !partition.covers(layer, x, y) {
				continue
			}
			if layer.category == lunarPartitionCategoryMoonset || layer.category == lunarPartitionCategoryMoonrise {
				doubleTinted++
				return
			}
		}
	})
	if uncovered != 0 {
		t.Errorf("penumbral-phase map leaves %d grid points without a region colour", uncovered)
	}
	if penumbraOnly == 0 || doubleTinted != 0 {
		t.Errorf("penumbra-only bands cover %d points, %d of them still tinted by a moonrise/moonset band", penumbraOnly, doubleTinted)
	}
}

// 与 NASA/EclipseWise 世界图同口径：月球在食甚时刻已在地平上的地点，无论月出月落落在食内哪一段，
// 都属于可见区，绝不能画成"不可见"；反之不可见区必须三刻都在地平下。
func TestLunarEclipseMapSVGNotVisibleFollowsHorizon(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2029, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		time.Date(2011, 6, 15, 12, 0, 0, 0, time.UTC),
	} {
		info, ok := eclipsecore.LunarEclipseOnDate(date)
		if !ok {
			t.Fatalf("%s: missing eclipse", date.Format("2006-01-02"))
		}
		doc, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
			Width: 960, Height: 640, Language: "zh", Location: time.UTC})
		if !ok {
			t.Fatalf("%s: missing map", date.Format("2006-01-02"))
		}
		partition := lunarMapPartitionOf(t, doc)
		// 0.02° 容差内的地点落在边界上，任何一类都不算错。
		wronglyInvisible, upAtMaximumInvisible, wronglyVisible := 0, 0, 0
		var sample string
		partition.grid(func(longitude, latitude, x, y float64, category string, layers int) {
			start := lunarPartitionAltitude(info.PenumbralStart, longitude, latitude)
			maximum := lunarPartitionAltitude(info.Maximum, longitude, latitude)
			end := lunarPartitionAltitude(info.PenumbralEnd, longitude, latitude)
			if math.Abs(start) < 0.02 || math.Abs(maximum) < 0.02 || math.Abs(end) < 0.02 {
				return
			}
			invisible := category == lunarPartitionCategoryUnavailable
			wantInvisible := start < 0 && maximum < 0 && end < 0
			if invisible && maximum > 0 {
				upAtMaximumInvisible++
				if sample == "" {
					sample = "(" + lunarPartitionFormat(longitude) + "," + lunarPartitionFormat(latitude) + ")"
				}
			}
			if invisible != wantInvisible {
				wronglyInvisible++
			}
			if !invisible && start > 0 && maximum > 0 && end > 0 && category != lunarPartitionCategoryEntire {
				wronglyVisible++
			}
		})
		t.Logf("%s: 食甚在地平上却画成不可见 %d 点%s；与三刻地平判据不符 %d 点；全程在地平上却不在全程可见区 %d 点",
			date.Format("2006-01-02"), upAtMaximumInvisible, sample, wronglyInvisible, wronglyVisible)
		if upAtMaximumInvisible != 0 {
			t.Errorf("%s: %d grid points whose Moon is above the horizon at greatest eclipse %s are painted as not visible",
				date.Format("2006-01-02"), upAtMaximumInvisible, sample)
		}
		if wronglyInvisible != 0 || wronglyVisible != 0 {
			t.Errorf("%s: visibility regions disagree with the horizon at P1/greatest/P4 for %d + %d grid points",
				date.Format("2006-01-02"), wronglyInvisible, wronglyVisible)
		}
	}
}

// "全程可见"必须同时满足 P1、食甚、P4 三刻月亮都在地平上：两端可见推不出中途可见，
// 高纬下中天会让月亮在食甚前后落到地平下，这类点归入带食月落。
// Visible throughout requires the Moon above the horizon at P1, greatest and P4 alike:
// visible at both contacts does not imply visible in between.
func TestLunarEclipseMapSVGEntireRegionRequiresGreatestVisibility(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2029, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2011, 6, 15, 12, 0, 0, 0, time.UTC),
	} {
		doc, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
			Width: 960, Height: 640, Language: "zh", Location: time.UTC})
		if !ok {
			t.Fatalf("%s: missing map", date.Format("2006-01-02"))
		}
		info, ok := eclipsecore.LunarEclipseOnDate(date)
		if !ok {
			t.Fatalf("%s: missing eclipse", date.Format("2006-01-02"))
		}
		partition := lunarMapPartitionOf(t, doc)
		holes, holeArea, missing, mislabelled, partialBand := 0, 0.0, 0, 0, 0
		partition.grid(func(longitude, latitude, x, y float64, category string, layers int) {
			inP1 := partition.inside("lunar-visible-p1-shape", x, y)
			inP4 := partition.inside("lunar-visible-p4-shape", x, y)
			inMaximum := partition.inside("lunar-visible-maximum-shape", x, y)
			cell := math.Cos(latitude * math.Pi / 180)
			switch {
			case inP1 && inP4 && inMaximum && category != lunarPartitionCategoryEntire:
				holes++
				holeArea += cell
			case !(inP1 && inP4 && inMaximum) && category == lunarPartitionCategoryEntire:
				missing++
			case inP1 && inP4 && !inMaximum && category == lunarPartitionCategoryMoonset:
				partialBand++
			}
			// 几何复核：三刻中任一刻月亮在地平下（留 0.05° 余量）就不得涂成全程可见。
			if category == lunarPartitionCategoryEntire {
				if lunarPartitionAltitude(info.PenumbralStart, longitude, latitude) < -0.05 ||
					lunarPartitionAltitude(info.Maximum, longitude, latitude) < -0.05 ||
					lunarPartitionAltitude(info.PenumbralEnd, longitude, latitude) < -0.05 {
					mislabelled++
				}
			}
		})
		t.Logf("%s: p1∩p4∩max 内未着全程可见色 %d 点（%.1f 平方度），全程可见区越界 %d 点，中间不可见带 %d 点，误标 %d 点",
			date.Format("2006-01-02"), holes, holeArea, missing, partialBand, mislabelled)
		if holes != 0 || missing != 0 || mislabelled != 0 {
			t.Errorf("%s: entire-eclipse region is not exactly p1∩p4∩max (holes=%d, extra=%d, mislabelled=%d)",
				date.Format("2006-01-02"), holes, missing, mislabelled)
		}
	}
}

func lunarPartitionFormat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func countsString(counts map[string]int) string {
	keys := []string{lunarPartitionCategoryEntire, lunarPartitionCategoryMoonset,
		lunarPartitionCategoryMoonrise, lunarPartitionCategoryUnavailable, "", lunarPartitionCategoryPenumbra}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if counts[key] == 0 {
			continue
		}
		name := key
		if name == "" {
			name = "uncovered"
		}
		parts = append(parts, name+"="+lunarPartitionFormat(float64(counts[key])))
	}
	return strings.Join(parts, " ")
}
