// 被测契约：通用天象搜索与各行星专用事件一致，并能找到大合与行星近星。
package basic

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

func TestSkyAspectMatchesMarsOpposition(t *testing.T) {
	start := 2460000.5
	// 专用函数返回 UT。
	want := UTC2TT(NextMarsOppositionWithInstance(start, nil))
	got := SkyAspectCrossingsWithInstance(SkyMars, SkySun, 180, start, start+800, nil)
	if len(got) == 0 || math.Abs(got[0].JDE-want)*86400 > 1 {
		t.Fatalf("opposition %v, want %.6f", got, want)
	}
}

// 专用函数按赤经留且返回 UT，通用搜索按黄经留，二者相差数小时属口径不同；这里只核对黄经速度确实变号。
func TestSkyStationsFlipLongitudeRate(t *testing.T) {
	got := SkyStationsWithInstance(SkyMars, 2460000.5, 2460800.5, nil)
	if len(got) != 2 {
		t.Fatalf("stations %v", got)
	}
	for _, e := range got {
		before, _ := SkyBodyApparentLoBoWithInstance(SkyMars, e.JDE-1, nil)
		at, _ := SkyBodyApparentLoBoWithInstance(SkyMars, e.JDE, nil)
		after, _ := SkyBodyApparentLoBoWithInstance(SkyMars, e.JDE+1, nil)
		if geodata.NormalizeLongitude(at-before)*geodata.NormalizeLongitude(after-at) >= 0 {
			t.Errorf("no rate flip at %.5f", e.JDE)
		}
		ut := UTC2TT(NextMarsProgradeToRetrogradeWithInstance(e.JDE-30, nil))
		if math.Abs(ut-e.JDE) > 5 && math.Abs(UTC2TT(NextMarsRetrogradeToProgradeWithInstance(e.JDE-30, nil))-e.JDE) > 5 {
			t.Errorf("station %.5f far from RA stations", e.JDE)
		}
	}
}

func TestSkyGreatConjunction2020(t *testing.T) {
	got := SkySeparationMinimaWithInstance(SkyJupiter, SkySaturn, 1, 2459100.5, 2459300.5, nil)
	if len(got) != 1 {
		t.Fatalf("minima %v", got)
	}
	when := JD2Date(TT2UTC(got[0].JDE))
	if when.Year() != 2020 || when.Month() != time.December || when.Day() < 21 || when.Day() > 22 || got[0].Value > 0.11 || got[0].Value < 0.09 {
		t.Errorf("great conjunction at %v separation %.4f°", when, got[0].Value)
	}
}

func TestSkyStarApproachRegulus(t *testing.T) {
	regulus := StarCoordinate{ID: "Regulus", RA: 152.09296, Dec: 11.96721, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000}
	got := SkyStarSeparationMinimaWithInstance(SkyVenus, regulus, 1.5, 2460400.5, 2460600.5, nil)
	if len(got) == 0 {
		t.Fatal("no Venus–Regulus approach")
	}
	for _, e := range got {
		if e.Value > 1.5 || e.Value < 0 {
			t.Errorf("approach %.4f° at %.5f", e.Value, e.JDE)
		}
	}
}

func TestSkySeparationMinimaFindsNarrowWindow(t *testing.T) {
	wide := SkySeparationMinimaWithInstance(SkyJupiter, SkySaturn, 1, 2459190, 2459220, nil)
	if len(wide) == 0 {
		t.Fatal("wide window returned no minimum")
	}
	center := wide[0].JDE
	for _, halfWidth := range []float64{0.1, 0.25, 0.49, 0.99, 1.5, 3} {
		got := SkySeparationMinimaWithInstance(SkyJupiter, SkySaturn, 1, center-halfWidth, center+halfWidth, nil)
		if len(got) == 0 {
			t.Errorf("halfWidth %v: no minimum", halfWidth)
			continue
		}
		if math.Abs(got[0].JDE-center) > 1e-4 {
			t.Errorf("halfWidth %v: JDE = %.6f, want %.6f", halfWidth, got[0].JDE, center)
		}
	}
}

func TestSkySeparationMinimaFindsEdgeOfWindow(t *testing.T) {
	wide := SkySeparationMinimaWithInstance(SkyJupiter, SkySaturn, 1, 2459190, 2459220, nil)
	if len(wide) == 0 {
		t.Fatal("wide window returned no minimum")
	}
	center := wide[0].JDE
	windows := [][2]float64{
		{center - 1.0/1440, center + 3.0/24},
		{center - 3.0/24, center + 1.0/1440},
		{center - 1.0/24, center + 1.0/24},
	}
	for _, window := range windows {
		got := SkySeparationMinimaWithInstance(SkyJupiter, SkySaturn, 1, window[0], window[1], nil)
		if len(got) != 1 || math.Abs(got[0].JDE-center) > 1e-4 {
			t.Errorf("window [%+.4f, %+.4f]: got %+v, want one minimum at %.6f", window[0]-center, window[1]-center, got, center)
		}
	}
	if got := SkySeparationMinimaWithInstance(SkyJupiter, SkySaturn, 1, center+0.5, center+3, nil); len(got) != 0 {
		t.Errorf("window after the minimum: got %+v, want none", got)
	}
}

func TestSkyScansTerminateOnDegenerateWindows(t *testing.T) {
	start := 2459205.26539
	windows := [][2]float64{
		{start, math.Nextafter(start, math.Inf(1))},
		{start, start + 26e-6/86400},
		{start, math.Inf(1)},
		{math.Inf(-1), start},
		{math.NaN(), start},
	}
	for _, window := range windows {
		SkySeparationMinimaWithInstance(SkyJupiter, SkySaturn, 1, window[0], window[1], nil)
		SkyLongitudeCrossingsWithInstance(SkySun, 280, window[0], window[1], nil)
		SkyAspectCrossingsWithInstance(SkyJupiter, SkySaturn, 0, window[0], window[1], nil)
	}
}

func TestSkyGridRejectsLimitInclusive(t *testing.T) {
	if _, _, ok := skyGrid(0, float64(skyGridMaxIntervals), 1, 1); ok {
		t.Error("grid at limit accepted")
	}
	if count, _, ok := skyGrid(0, float64(skyGridMaxIntervals-1), 1, 1); !ok || count != skyGridMaxIntervals-1 {
		t.Errorf("grid below limit: count=%d ok=%v", count, ok)
	}
}

func TestSkyLongitudeCrossingsIncludesClosedEndpoints(t *testing.T) {
	const jde = 2451545.0
	lo, _ := SkyBodyApparentLoBoWithInstance(SkySun, jde, nil)
	for _, window := range [][2]float64{{jde - 2, jde}, {jde, jde + 2}} {
		got := SkyLongitudeCrossingsWithInstance(SkySun, lo, window[0], window[1], nil)
		if len(got) != 1 || math.Abs(got[0].JDE-jde) > 1e-5 {
			t.Errorf("window [%v, %v]: got %+v, want one crossing at %v", window[0], window[1], got, jde)
		}
	}
}
