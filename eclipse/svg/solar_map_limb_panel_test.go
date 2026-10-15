package svg

// 全球日食图开月缘时，面板数值与画出的界线同源；剖面失败时整张图保持圆月面。

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/moon/limb"
)

type failingEclipseLimb struct{}

func (failingEclipseLimb) LimbRadiusKM(_, _, _ float64) (float64, bool) { return 0, false }

func TestSolarEclipseMapLunarLimbPanelMatchesLimbResults(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	profile := limb.Default()
	circular, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{Location: time.UTC})
	if !ok {
		t.Fatal("circular map missing")
	}
	corrected, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{Location: time.UTC, LunarLimb: profile})
	if !ok {
		t.Fatal("limb map missing")
	}

	module := eclipsecore.NewWithInstance(nil).WithLunarLimb(profile)
	global, _ := module.SolarEclipseOnDateNASABulletinSplitK(date)
	if !global.LunarLimbApplied {
		t.Fatal("global limb result not flagged")
	}
	// 与地图同一组路径选项，带宽才和画出的南北界同源。
	path, ok := module.SolarEclipseCentralPathNASABulletinSplitK(date, eclipsecore.SolarEclipsePathOptions{
		Step: 2 * time.Minute, TargetSpacingKM: 150, SkipCentralBand: true,
	})
	if !ok || !path.PathWidthDefined {
		t.Fatal("limb central path missing")
	}
	limbWidth := fmt.Sprintf("%.1f km", path.Greatest.WidthKM)
	circularGlobal, _ := eclipsecore.SolarEclipseOnDateNASABulletinSplitK(date)
	circularWidth := fmt.Sprintf("%.1f km", circularGlobal.PathWidthKM)
	if limbWidth == circularWidth {
		t.Fatalf("limb width %s equals circular width; fixture no longer distinguishes them", limbWidth)
	}
	if !strings.Contains(corrected, limbWidth) || strings.Contains(corrected, circularWidth) {
		t.Fatalf("limb map must show width %s and not %s", limbWidth, circularWidth)
	}
	if !strings.Contains(circular, circularWidth) {
		t.Fatalf("circular map must show width %s", circularWidth)
	}

	local, ok := module.GeometricLocalSolarEclipseOnDateNASABulletinSplitK(date, global.GreatestLongitude, global.GreatestLatitude, 0)
	if !ok || !local.LunarLimbApplied {
		t.Fatal("limb local panel result missing or not flagged")
	}
	if magnitude := fmt.Sprintf("%.4f", local.Magnitude); !strings.Contains(corrected, magnitude) {
		t.Fatalf("limb map must show the limb local magnitude %s", magnitude)
	}
	begin := global.PartialBeginOnEarth.UTC().Format("15:04:05")
	if begin == circularGlobal.PartialBeginOnEarth.UTC().Format("15:04:05") {
		t.Fatal("limb P1 rounds to the circular value; pick a fixture where it moves")
	}
	if !strings.Contains(corrected, begin) || strings.Contains(circular, begin) {
		t.Fatalf("limb P1 %s must appear only in the limb map", begin)
	}
	if !strings.Contains(corrected, "按月缘剖面修正全球始末、食甚点站心值、中心带界线、可见域边界") {
		t.Fatal("limb map does not declare every corrected layer")
	}
}

// selectiveEclipseLimb 只在位置角 [low, high) 内给出剖面值并报告半径范围；外切扫描靠范围剪枝停在日心方向附近，需要全方向的判据退回圆月面。
type selectiveEclipseLimb struct{ low, high float64 }

func (s selectiveEclipseLimb) LimbRadiusKM(longitude, latitude, angle float64) (float64, bool) {
	if angle < s.low || angle >= s.high {
		return 0, false
	}
	return limb.Default().LimbRadiusKM(longitude, latitude, angle)
}

func (selectiveEclipseLimb) RadiusRangeKM() (float64, float64) { return limb.Default().RadiusRangeKM() }

func (selectiveEclipseLimb) AngleStepDeg() float64 { return limb.Default().AngleStepDeg() }

func TestSolarEclipseMapFooterDeclaresOnlyCorrectedLayers(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	circular, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{Location: time.UTC})
	if !ok {
		t.Fatal("circular map missing")
	}
	// 位置角 45°～90° 只覆盖偏食外边界的接触点：全球始末、站心面板与中心界线都退回圆月面。
	selective, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{Location: time.UTC, LunarLimb: selectiveEclipseLimb{45, 90}})
	if !ok {
		t.Fatal("selective-limb map missing")
	}
	if !strings.Contains(selective, "按月缘剖面修正可见域边界（食分线与等时线仍按圆月面）") {
		t.Fatal("selective map must declare only the visibility boundary")
	}
	for _, layer := range []string{"全球始末", "食甚点站心值", "中心带界线"} {
		if strings.Contains(selective, layer) {
			t.Errorf("selective map declares uncorrected layer %s", layer)
		}
	}
	circularGlobal, _ := eclipsecore.SolarEclipseOnDateNASABulletinSplitK(date)
	if width := fmt.Sprintf("%.1f km", circularGlobal.PathWidthKM); !strings.Contains(selective, width) {
		t.Errorf("selective map must keep the circular width %s while the limits stay circular", width)
	}
	if strings.Contains(circular, "按月缘剖面修正") {
		t.Error("circular map declares a limb correction")
	}
}

func TestEclipseSVGsWithFailingLimbStayCircular(t *testing.T) {
	solarDate := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	lunarDate := time.Date(2025, time.September, 7, 0, 0, 0, 0, time.UTC)
	failing := failingEclipseLimb{}
	pairs := map[string][2]func() (string, bool){
		"local solar": {
			func() (string, bool) {
				return LocalSolarEclipseSVG(solarDate, -96.8, 32.78, 0, LocalSolarEclipseSVGOptions{Location: time.UTC})
			},
			func() (string, bool) {
				return LocalSolarEclipseSVG(solarDate, -96.8, 32.78, 0, LocalSolarEclipseSVGOptions{Location: time.UTC, LunarLimb: failing})
			},
		},
		"solar map": {
			func() (string, bool) {
				return SolarEclipseMapSVG(solarDate, SolarEclipseMapSVGOptions{Location: time.UTC})
			},
			func() (string, bool) {
				return SolarEclipseMapSVG(solarDate, SolarEclipseMapSVGOptions{Location: time.UTC, LunarLimb: failing})
			},
		},
		"lunar": {
			func() (string, bool) { return LunarEclipseSVG(lunarDate, LunarEclipseSVGOptions{Location: time.UTC}) },
			func() (string, bool) {
				return LunarEclipseSVG(lunarDate, LunarEclipseSVGOptions{Location: time.UTC, LunarLimb: failing})
			},
		},
		"lunar map": {
			func() (string, bool) {
				return LunarEclipseMapSVG(lunarDate, LunarEclipseMapSVGOptions{Location: time.UTC})
			},
			func() (string, bool) {
				return LunarEclipseMapSVG(lunarDate, LunarEclipseMapSVGOptions{Location: time.UTC, LunarLimb: failing})
			},
		},
		"lunar detailed": {
			func() (string, bool) {
				return LunarEclipseDetailedSVG(lunarDate, LunarEclipseDetailedSVGOptions{Location: time.UTC})
			},
			func() (string, bool) {
				return LunarEclipseDetailedSVG(lunarDate, LunarEclipseDetailedSVGOptions{Location: time.UTC, LunarLimb: failing})
			},
		},
	}
	for name, pair := range pairs {
		circular, ok := pair[0]()
		if !ok {
			t.Fatalf("%s: circular chart missing", name)
		}
		failed, ok := pair[1]()
		if !ok {
			t.Fatalf("%s: failing-limb chart missing", name)
		}
		if failed != circular {
			t.Errorf("%s: failing limb profile changed the chart", name)
		}
	}
}

// TestSolarEclipseMapContactMarkerMatchesCorrectedTime 检查全球图的 P1/P4 标记与页脚时刻同源：
// 月缘修正过时刻时，标记位置必须取修正后的接触点，而不是圆月面接触点。
func TestSolarEclipseMapContactMarkerMatchesCorrectedTime(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	module := eclipsecore.NewWithInstance(nil).WithLunarLimb(limb.Default())
	global, ok := module.SolarEclipseOnDateNASABulletinSplitK(date)
	if !ok {
		t.Fatal("global result missing")
	}
	if !global.LunarLimbApplied || !global.PartialBeginContact.Defined {
		t.Fatal("the default limb profile must correct the global contacts")
	}
	circular, ok := eclipsecore.SolarEclipsePartialFootprintsNASABulletinSplitK(date, eclipsecore.SolarEclipsePartialFootprintOptions{})
	if !ok {
		t.Fatal("circular footprints missing")
	}
	// 修正后的接触点必须真的离开圆月面接触点，否则这条回归就测不到东西。
	moved := math.Hypot(
		(global.PartialBeginContact.Longitude-circular.P1.Longitude)*math.Cos(circular.P1.Latitude*math.Pi/180),
		global.PartialBeginContact.Latitude-circular.P1.Latitude)
	if moved < 1e-4 {
		t.Fatalf("corrected P1 did not move (%.6f deg)", moved)
	}
	if delta := global.PartialBeginOnEarth.Sub(circular.P1.Time).Seconds(); math.Abs(delta) < 0.5 {
		t.Fatalf("corrected P1 time moved by only %.3f s", delta)
	}
}
