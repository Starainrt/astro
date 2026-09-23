package svg

import (
	"crypto/md5"
	"encoding/hex"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
)

// 半影阶段开关的契约：零值输出与不区分半影阶段的基线逐字节一致，打开后补画 U1–U4 地平边界、
// 单独着色只看得见半影的月出月落带，且每条边界仍落在它自己的接触时刻上。
// Contract of the penumbral-phase switch: DisablePenumbralPhase reproduces the penumbral-agnostic baseline
// byte for byte, while the enabled value adds the U1-U4 horizons and the penumbra-only bands, with
// every boundary still sitting on its own contact instant.

var lunarPenumbralPhaseBaseline = []struct {
	name   string
	digest string
	length int
}{
	{"2026-03-03 全食 960x640 zh CST", "4e83ea42b18c703bdde6715f572169d5", 243483},
	{"2026-03-03 全食 960x640 en CST", "21696a72ec3bc77a21752e2d675a230b", 243544},
	{"2026-03-03 全食 1200x800 zh UTC", "e5a72a8f0b73449f85b4e433f6054069", 247413},
	{"2026-03-03 全食 1200x800 en UTC", "1a0386abae194bdbee0bd5c0cb7a5a15", 247482},
	{"2025-03-14 全食 960x640 zh CST", "84c39bb9f893678c1cdf354ba8acf850", 250237},
	{"2025-03-14 全食 1200x800 zh UTC", "dad36f434ba2bfd7369a430b6f00e567", 255213},
	{"2026-03-03 北极投影 zh", "8ee2f75f0f7c617b136576635d908cb4", 159000},
	{"2026-03-03 南极投影 zh", "931e502692cdc5b0d09e8f91629f1b2f", 84934},
	{"2016-08-18 半影 1200x800 zh UTC", "6d9c60787cc1b703abdd13c4b75a2a92", 241387},
	{"2029-01-01 全食 960x640 en CST", "95c0a9d618bda4c0510fa827e104ceaa", 231567},
	{"2026-03-03 详细版 zh", "31795334a05621f3a5a0beff1b63b5bd", 270844},
	{"2026-03-03 详细版 en", "c5ee0c76f875c3c5fe4a7ff29a90a653", 270992},
	{"2029-01-01 详细版 zh", "b0521aa165454262df4344168c9100e9", 257156},
	{"2016-08-18 详细版 en", "0c5fdc6c04f993504ae970c2e8861ccb", 261488},
}

func TestLunarEclipseMapSVGDisablePenumbralPhaseMatchesBaseline(t *testing.T) {
	render := lunarPenumbralPhaseBaselineRenderers()
	if len(render) != len(lunarPenumbralPhaseBaseline) {
		t.Fatalf("baseline table has %d cases, renderers %d", len(lunarPenumbralPhaseBaseline), len(render))
	}
	for _, want := range lunarPenumbralPhaseBaseline {
		draw, ok := render[want.name]
		if !ok {
			t.Fatalf("baseline case %q has no renderer", want.name)
		}
		doc, rendered := draw()
		if !rendered {
			t.Fatalf("%s: chart missing", want.name)
		}
		sum := md5.Sum([]byte(doc))
		if got := hex.EncodeToString(sum[:]); got != want.digest || len(doc) != want.length {
			t.Fatalf("%s: md5=%s bytes=%d, want md5=%s bytes=%d",
				want.name, got, len(doc), want.digest, want.length)
		}
	}
}

func lunarPenumbralPhaseBaselineRenderers() map[string]func() (string, bool) {
	cst := time.FixedZone("CST", 8*3600)
	utc := time.UTC
	return map[string]func() (string, bool){
		"2026-03-03 全食 960x640 zh CST": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, cst),
				LunarEclipseMapSVGOptions{Width: 960, Height: 640, Language: "zh", Location: cst, DisablePenumbralPhase: true})
		},
		"2026-03-03 全食 960x640 en CST": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, cst),
				LunarEclipseMapSVGOptions{Width: 960, Height: 640, Language: "en", Location: cst, DisablePenumbralPhase: true})
		},
		"2026-03-03 全食 1200x800 zh UTC": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, utc),
				LunarEclipseMapSVGOptions{Width: 1200, Height: 800, Language: "zh", Location: utc, DisablePenumbralPhase: true})
		},
		"2026-03-03 全食 1200x800 en UTC": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, utc),
				LunarEclipseMapSVGOptions{Width: 1200, Height: 800, Language: "en", Location: utc, DisablePenumbralPhase: true})
		},
		"2025-03-14 全食 960x640 zh CST": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2025, 3, 14, 0, 0, 0, 0, cst),
				LunarEclipseMapSVGOptions{Width: 960, Height: 640, Language: "zh", Location: cst, DisablePenumbralPhase: true})
		},
		"2025-03-14 全食 1200x800 zh UTC": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2025, 3, 14, 0, 0, 0, 0, utc),
				LunarEclipseMapSVGOptions{Width: 1200, Height: 800, Language: "zh", Location: utc, DisablePenumbralPhase: true})
		},
		"2026-03-03 北极投影 zh": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, utc),
				LunarEclipseMapSVGOptions{Language: "zh", Location: utc, Projection: EclipseMapProjectionNorthPolar, DisablePenumbralPhase: true})
		},
		"2026-03-03 南极投影 zh": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, utc),
				LunarEclipseMapSVGOptions{Language: "zh", Location: utc, Projection: EclipseMapProjectionSouthPolar, DisablePenumbralPhase: true})
		},
		"2016-08-18 半影 1200x800 zh UTC": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2016, 8, 18, 0, 0, 0, 0, utc),
				LunarEclipseMapSVGOptions{Width: 1200, Height: 800, Language: "zh", Location: utc, DisablePenumbralPhase: true})
		},
		"2029-01-01 全食 960x640 en CST": func() (string, bool) {
			return LunarEclipseMapSVG(time.Date(2029, 1, 1, 0, 0, 0, 0, cst),
				LunarEclipseMapSVGOptions{Width: 960, Height: 640, Language: "en", Location: cst, DisablePenumbralPhase: true})
		},
		"2026-03-03 详细版 zh": func() (string, bool) {
			return LunarEclipseDetailedSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, cst),
				LunarEclipseDetailedSVGOptions{Language: "zh", Location: cst, DisablePenumbralPhase: true})
		},
		"2026-03-03 详细版 en": func() (string, bool) {
			return LunarEclipseDetailedSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, cst),
				LunarEclipseDetailedSVGOptions{Language: "en", Location: cst, DisablePenumbralPhase: true})
		},
		"2029-01-01 详细版 zh": func() (string, bool) {
			return LunarEclipseDetailedSVG(time.Date(2029, 1, 1, 0, 0, 0, 0, cst),
				LunarEclipseDetailedSVGOptions{Language: "zh", Location: cst, DisablePenumbralPhase: true})
		},
		"2016-08-18 详细版 en": func() (string, bool) {
			return LunarEclipseDetailedSVG(time.Date(2016, 8, 18, 12, 0, 0, 0, utc),
				LunarEclipseDetailedSVGOptions{Language: "en", Location: cst, DisablePenumbralPhase: true})
		},
	}
}

func TestLunarEclipseMapSVGPenumbralPhaseAddsUmbralContacts(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 3, 3, 0, 0, 0, 0, cst)
	options := LunarEclipseMapSVGOptions{Language: "zh", Location: cst}
	on, ok := LunarEclipseMapSVG(date, options)
	if !ok {
		t.Fatal("missing penumbral-phase lunar-eclipse map")
	}
	options.DisablePenumbralPhase = true
	off, ok := LunarEclipseMapSVG(date, options)
	if !ok {
		t.Fatal("missing lunar-eclipse map")
	}
	for _, want := range []string{
		`class="u1-horizon"`, `class="u2-horizon"`, `class="u3-horizon"`, `class="u4-horizon"`,
		`class="penumbra-only-region penumbra-moonset-region"`, `class="penumbra-only-region penumbra-moonrise-region"`,
		`mask id="lunar-not-umbral-visible-mask"`, `clip-path="url(#lunar-visible-umbral)"`,
		`id="lunar-visible-umbral-1"`,
		`id="lunar-visible-u1-shape"`, `id="lunar-visible-u4-shape"`,
		"半影月出", "半影月落", "本影阶段 U1",
	} {
		if !strings.Contains(on, want) {
			t.Fatalf("penumbral-phase map missing %q", want)
		}
		if strings.Contains(off, want) {
			t.Fatalf("default map must not contain %q", want)
		}
	}
	if err := validateEclipseMapXML(on); err != nil {
		t.Fatalf("penumbral-phase map is not valid XML: %v", err)
	}
	for _, id := range []string{"lunar-visible-p1-shape", "lunar-visible-p4-shape", "lunar-visible-maximum-shape"} {
		shape := regexp.MustCompile(`id="` + id + `" d="([^"]*)"`)
		a, b := shape.FindStringSubmatch(off), shape.FindStringSubmatch(on)
		if a == nil || b == nil || a[1] != b[1] {
			t.Fatalf("shape %s changed when the penumbral phase is enabled", id)
		}
	}
	english, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
		Language: "en", Location: cst})
	if !ok || !strings.Contains(english, "Penumbra moonrise") || !strings.Contains(english, "Penumbra moonset") ||
		!strings.Contains(english, "Umbral phases U1") {
		t.Fatal("English penumbral-phase map misses its legend entry or contact row")
	}
	// 纯半影月食没有本影接触，按 NASA 口径不该凭空生成本影边界。
	onlyPenumbral, ok := LunarEclipseMapSVG(time.Date(2016, 8, 18, 0, 0, 0, 0, time.UTC),
		LunarEclipseMapSVGOptions{Language: "zh", Location: time.UTC})
	if !ok {
		t.Fatal("missing penumbral-only lunar-eclipse map")
	}
	for _, unwanted := range []string{"u1-horizon", "penumbra-only-region", "半影月出", "半影月落"} {
		if strings.Contains(onlyPenumbral, unwanted) {
			t.Fatalf("penumbral-only eclipse must not gain %q", unwanted)
		}
	}
	onlyPenumbralDefault, ok := LunarEclipseMapSVG(time.Date(2016, 8, 18, 0, 0, 0, 0, time.UTC),
		LunarEclipseMapSVGOptions{Language: "zh", Location: time.UTC, DisablePenumbralPhase: true})
	if !ok || onlyPenumbralDefault != onlyPenumbral {
		t.Fatal("penumbral-only eclipse must render identically with and without the switch")
	}
}

// 同判：U1/U4 形状等价于「该接触时刻月球中心在地平线上」，半影带等价于对应的三段高度判据，
// 每条地平边界线都落在自己的接触时刻上（P1 线不能落在 U1 时刻）。
func TestLunarEclipseMapSVGPenumbraBandsMatchHorizonCriterion(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
	} {
		info, ok := eclipsecore.LunarEclipseOnDate(date)
		if !ok || !info.HasPartial {
			t.Fatalf("%s: missing umbral lunar eclipse", date.Format("2006-01-02"))
		}
		doc, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
			Width: 1200, Height: 800, Language: "zh", Location: time.UTC})
		if !ok {
			t.Fatalf("%s: missing penumbral-phase map", date.Format("2006-01-02"))
		}
		frame := lunarPenumbralTestFrame(t, doc)
		p1 := lunarPenumbralTestShape(t, doc, "lunar-visible-p1-shape")
		p4 := lunarPenumbralTestShape(t, doc, "lunar-visible-p4-shape")
		u1 := lunarPenumbralTestShape(t, doc, "lunar-visible-u1-shape")
		u4 := lunarPenumbralTestShape(t, doc, "lunar-visible-u4-shape")

		for _, line := range []struct {
			class string
			at    time.Time
			other time.Time
		}{
			{"p1-horizon", info.PenumbralStart, info.PartialStart},
			{"p4-horizon", info.PenumbralEnd, info.PartialEnd},
			{"u1-horizon", info.PartialStart, info.PenumbralStart},
			{"u4-horizon", info.PartialEnd, info.PenumbralEnd},
		} {
			maxOnContact := 0.0
			var otherAltitudes []float64
			vertices := lunarPenumbralTestLine(t, doc, line.class)
			if len(vertices) < 30 {
				t.Fatalf("%s: %s has %d vertices", date.Format("2006-01-02"), line.class, len(vertices))
			}
			for _, vertex := range vertices {
				longitude, latitude := lunarPenumbralTestUnproject(frame, vertex[0], vertex[1])
				maxOnContact = math.Max(maxOnContact, math.Abs(lunarPenumbralTestAltitude(line.at, longitude, latitude)))
				otherAltitudes = append(otherAltitudes, math.Abs(lunarPenumbralTestAltitude(line.other, longitude, latitude)))
			}
			if maxOnContact > 0.01 {
				t.Fatalf("%s: %s deviates %.4f deg from its own contact", date.Format("2006-01-02"), line.class, maxOnContact)
			}
			// 两条地平线必然相交于两点，交点是整条线唯一贴近另一组接触的地方；中位数说明其余部分分得很开。
			sort.Float64s(otherAltitudes)
			median := otherAltitudes[len(otherAltitudes)/2]
			if median < 2 {
				t.Fatalf("%s: %s sits only %.3f deg (median) from the other contact", date.Format("2006-01-02"), line.class, median)
			}
			t.Logf("%s %s: 与自身接触最大偏差 %.5f 度，与另一组接触 |alt| 中位数 %.3f 度（最小 %.3f）",
				date.Format("2006-01-02"), line.class, maxOnContact, median, otherAltitudes[0])
		}

		judged, mismatches := 0, 0
		for longitude := -179.0; longitude < 180; longitude += 2 {
			for latitude := -89.0; latitude < 90; latitude += 2 {
				x, y, projectable := frame.Project(longitude, latitude)
				if !projectable {
					continue
				}
				altitudes := [4]float64{
					lunarPenumbralTestAltitude(info.PenumbralStart, longitude, latitude),
					lunarPenumbralTestAltitude(info.PartialStart, longitude, latitude),
					lunarPenumbralTestAltitude(info.PartialEnd, longitude, latitude),
					lunarPenumbralTestAltitude(info.PenumbralEnd, longitude, latitude),
				}
				nearHorizon := false
				for _, altitude := range altitudes {
					if math.Abs(altitude) < 0.05 {
						nearHorizon = true
					}
				}
				if nearHorizon {
					continue
				}
				judged++
				if lunarPenumbralTestInside(u1, x, y) != (altitudes[1] > 0) {
					t.Errorf("%s: U1 shape disagrees with the U1 horizon at (%g,%g)", date.Format("2006-01-02"), longitude, latitude)
				}
				if lunarPenumbralTestInside(u4, x, y) != (altitudes[2] > 0) {
					t.Errorf("%s: U4 shape disagrees with the U4 horizon at (%g,%g)", date.Format("2006-01-02"), longitude, latitude)
				}
				// SVG 组合规则：月落带 = P1 可见、P4 与 U1 不可见；月出带 = P4 可见、P1 与 U4 不可见。
				svgMoonset := lunarPenumbralTestInside(p1, x, y) && !lunarPenumbralTestInside(p4, x, y) && !lunarPenumbralTestInside(u1, x, y)
				svgMoonrise := lunarPenumbralTestInside(p4, x, y) && !lunarPenumbralTestInside(p1, x, y) && !lunarPenumbralTestInside(u4, x, y)
				wantMoonset := altitudes[0] > 0 && altitudes[1] <= 0 && altitudes[3] <= 0
				wantMoonrise := altitudes[3] > 0 && altitudes[2] <= 0 && altitudes[0] <= 0
				if svgMoonset != wantMoonset {
					mismatches++
				}
				if svgMoonrise != wantMoonrise {
					mismatches++
				}
			}
		}
		if judged < 5000 {
			t.Fatalf("%s: too few judged samples (%d)", date.Format("2006-01-02"), judged)
		}
		if mismatches != 0 {
			t.Fatalf("%s: penumbra band boundaries disagree with the horizon criterion at %d of %d judged sites",
				date.Format("2006-01-02"), mismatches, judged)
		}
	}
}

func lunarPenumbralTestAltitude(at time.Time, longitude, latitude float64) float64 {
	return basic.HMoonHeight(basic.Date2JD(at.UTC()), longitude, latitude, 0)
}

func lunarPenumbralTestFrame(t *testing.T, doc string) svgFrameBox {
	t.Helper()
	match := regexp.MustCompile(`<clipPath id="lunar-map-clip"><rect x="([-0-9.]+)" y="([-0-9.]+)" width="([-0-9.]+)" height="([-0-9.]+)"`).FindStringSubmatch(doc)
	if match == nil {
		// 圆盘版式（极区/正射）画的是圆，用 map-frame 圆的外接方形当图框。
		disk := regexp.MustCompile(`<circle class="map-frame" cx="([-0-9.]+)" cy="([-0-9.]+)" r="([-0-9.]+)"`).FindStringSubmatch(doc)
		if disk == nil {
			t.Fatal("map frame not found")
		}
		cx, _ := strconv.ParseFloat(disk[1], 64)
		cy, _ := strconv.ParseFloat(disk[2], 64)
		radius, _ := strconv.ParseFloat(disk[3], 64)
		return svgFrameBox{x: cx - radius, y: cy - radius, width: 2 * radius, height: 2 * radius}
	}
	values := make([]float64, 4)
	for index := range values {
		values[index], _ = strconv.ParseFloat(match[index+1], 64)
	}
	return svgFrameBox{x: values[0], y: values[1], width: values[2], height: values[3]}
}

type svgFrameBox struct{ x, y, width, height float64 }

func (frame svgFrameBox) Project(longitude, latitude float64) (float64, float64, bool) {
	return frame.x + (longitude+180)/360*frame.width, frame.y + (90-latitude)/180*frame.height, true
}

func lunarPenumbralTestUnproject(frame svgFrameBox, x, y float64) (float64, float64) {
	return (x-frame.x)/frame.width*360 - 180, 90 - (y-frame.y)/frame.height*180
}

func lunarPenumbralTestShape(t *testing.T, doc, id string) [][][2]float64 {
	t.Helper()
	match := regexp.MustCompile(`id="` + id + `" d="([^"]*)"`).FindStringSubmatch(doc)
	if match == nil {
		t.Fatalf("shape %s not found", id)
	}
	return lunarPenumbralTestRings(t, match[1])
}

func lunarPenumbralTestLine(t *testing.T, doc, class string) [][2]float64 {
	t.Helper()
	match := regexp.MustCompile(`class="` + class + `" d="([^"]*)"`).FindStringSubmatch(doc)
	if match == nil {
		t.Fatalf("boundary %s not found", class)
	}
	rings := lunarPenumbralTestRings(t, match[1])
	if len(rings) == 0 {
		t.Fatalf("boundary %s has no vertices", class)
	}
	return rings[0]
}

func lunarPenumbralTestRings(t *testing.T, path string) [][][2]float64 {
	t.Helper()
	var rings [][][2]float64
	for _, chunk := range strings.Split(path, "Z") {
		fields := strings.Fields(chunk)
		var ring [][2]float64
		for index := 0; index < len(fields); {
			if fields[index] == "M" || fields[index] == "L" {
				x, _ := strconv.ParseFloat(fields[index+1], 64)
				y, _ := strconv.ParseFloat(fields[index+2], 64)
				ring = append(ring, [2]float64{x, y})
				index += 3
				continue
			}
			index++
		}
		if len(ring) >= 3 {
			rings = append(rings, ring)
		}
	}
	return rings
}

func lunarPenumbralTestInside(rings [][][2]float64, x, y float64) bool {
	for _, ring := range rings {
		if pointInLunarVisibilityPolygon(x, y, ring) {
			return true
		}
	}
	return false
}

// 仅见半影的两条带必须排除"本影区间里月亮曾在地平上"的站点：只挖 U1/U4 两刻会漏掉两刻之间的窗口，
// 按经度插值又会把近乎子午线方向的边界切进真实区域（1932-03-22 西经 75.5° 一带 0.5° 经度内上边界
// 从 +19.9° 掉到 −27.2°）。掩膜按 15 分钟档位取样整球可见半球，残余的掠射窗口深度约 0.05°
// （约 0.15 像素）；深度更浅、只持续几分钟的窗口只由站点 API 与 GeoJSON 精确判定。
func TestLunarEclipseMapSVGPenumbralBandsExcludeUmbralWindow(t *testing.T) {
	date := time.Date(1932, 3, 22, 0, 0, 0, 0, time.UTC)
	doc, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
		Width: 1200, Height: 800, Language: "zh", Location: time.UTC})
	if !ok {
		t.Fatal("missing lunar-eclipse map")
	}
	partition := lunarMapPartitionOf(t, doc)
	for _, witness := range []struct {
		name     string
		lon, lat float64
		want     string
	}{
		{"U1 已在地平上（+0.0041 度）", -91.75, -88.75, lunarPartitionCategoryMoonrise},
		{"U1 高度 +0.21 度（陡边界，1° 列会切进去）", -75.5, 6.5, lunarPartitionCategoryMoonset},
		{"本影极值 -2.25 度（宽余量）", -68.75, -70.15, lunarPartitionCategoryPenumbra},
	} {
		x, y, projectable := partition.frame.Project(witness.lon, witness.lat)
		if !projectable {
			t.Fatalf("%s: 投影失败", witness.name)
		}
		category, _ := partition.painted(x, y)
		if category != witness.want {
			t.Fatalf("%s (%.4f,%.4f) category=%q want %q", witness.name, witness.lon, witness.lat, category, witness.want)
		}
	}
}
