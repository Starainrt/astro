package svg

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/ephemtest"
	"github.com/starainrt/astro/moon"
)

// 本文件钉住详细版式三入口的契约：确定性、画布边界、面板数值与选项校验。

func TestStarOccultationDetailedSVGDeterministicXML(t *testing.T) {
	path := occultationTestStarPath(t)
	options := OccultationDetailedSVGOptions{Width: 1000, Height: 1414, Location: time.UTC}
	first, err := StarOccultationDetailedSVG(path, hr4799StarCoordinate(), options)
	if err != nil {
		t.Fatalf("StarOccultationDetailedSVG() error = %v", err)
	}
	if len(first) == 0 {
		t.Fatal("StarOccultationDetailedSVG() returned an empty diagram")
	}
	second, err := StarOccultationDetailedSVG(path, hr4799StarCoordinate(), options)
	if err != nil {
		t.Fatalf("second StarOccultationDetailedSVG() error = %v", err)
	}
	if first != second {
		t.Fatal("detailed stellar SVG is not byte-identical across renders")
	}
	if err := validateXML(first); err != nil {
		t.Fatalf("detailed stellar SVG is not valid XML: %v", err)
	}
	for _, want := range []string{`class="occultation-detailed-panel"`, "月亮（地心坐标）", "天平动"} {
		if !strings.Contains(first, want) {
			t.Fatalf("detailed stellar SVG is missing %q", want)
		}
	}
}

func TestPlanetOccultationDetailedSVGDeterministicXML(t *testing.T) {
	path := samplePlanetOccultationPath()
	options := OccultationDetailedSVGOptions{Width: 1414, Height: 1000, Location: time.UTC}
	first, err := PlanetOccultationDetailedSVG(path, options)
	if err != nil {
		t.Fatalf("PlanetOccultationDetailedSVG() error = %v", err)
	}
	if len(first) == 0 {
		t.Fatal("PlanetOccultationDetailedSVG() returned an empty diagram")
	}
	second, err := PlanetOccultationDetailedSVG(path, options)
	if err != nil {
		t.Fatalf("second PlanetOccultationDetailedSVG() error = %v", err)
	}
	if first != second {
		t.Fatal("detailed planetary SVG is not byte-identical across renders")
	}
	if err := validateXML(first); err != nil {
		t.Fatalf("detailed planetary SVG is not valid XML: %v", err)
	}
	if !strings.Contains(first, "全掩始") || !strings.Contains(first, "全掩带宽") {
		t.Fatal("detailed planetary SVG dropped total-band content")
	}
}

func TestOccultationDetailedSVGKeepsContentInsideCanvas(t *testing.T) {
	sizes := [][2]int{{1000, 1414}, {1414, 1000}, {800, 600}}
	starPath := occultationTestStarPath(t)
	for _, size := range sizes {
		size := size
		t.Run(strconv.Itoa(size[0])+"x"+strconv.Itoa(size[1]), func(t *testing.T) {
			options := OccultationDetailedSVGOptions{Width: size[0], Height: size[1], Location: time.UTC}
			diagrams := map[string]string{}
			star, err := StarOccultationDetailedSVG(starPath, hr4799StarCoordinate(), options)
			if err != nil {
				t.Fatalf("stellar detailed render: %v", err)
			}
			diagrams["stellar"] = star
			planet, err := PlanetOccultationDetailedSVG(samplePlanetOccultationPath(), options)
			if err != nil {
				t.Fatalf("planetary detailed render: %v", err)
			}
			diagrams["planetary"] = planet
			for name, diagram := range diagrams {
				texts := occultationSVGTexts(t, diagram)
				if len(texts) == 0 {
					t.Fatalf("%s detailed SVG has no text", name)
				}
				if outside := occultationSVGTextsOutsideCanvas(texts, float64(size[0]), float64(size[1])); len(outside) > 0 {
					t.Fatalf("%s detailed SVG has %d texts outside %dx%d: %v", name, len(outside), size[0], size[1], outside)
				}
				for _, rect := range occultationSVGPanelRects(t, diagram) {
					if rect[0] < 0 || rect[1] < 0 || rect[0]+rect[2] > float64(size[0]) || rect[1]+rect[3] > float64(size[1]) {
						t.Fatalf("%s detailed SVG panel rect %.1f,%.1f %.1fx%.1f leaves %dx%d",
							name, rect[0], rect[1], rect[2], rect[3], size[0], size[1])
					}
				}
			}
		})
	}
}

func TestOccultationDetailedSVGRejectsCanvasTooSmall(t *testing.T) {
	starPath := sampleStarOccultationPath()
	planetPath := samplePlanetOccultationPath()
	tests := []struct {
		name          string
		width, height int
	}{
		{name: "reference overflow", width: 400, height: 300},
		{name: "narrow", width: 479, height: 1414},
		{name: "short", width: 1000, height: 319},
		{name: "single pixel", width: 1, height: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := OccultationDetailedSVGOptions{Width: test.width, Height: test.height}
			if _, err := StarOccultationDetailedSVG(starPath, hr4799StarCoordinate(), options); !errors.Is(err, ErrInvalidOccultationDetailedSVGOptions) {
				t.Fatalf("stellar detailed %dx%d error = %v, want ErrInvalidOccultationDetailedSVGOptions", test.width, test.height, err)
			}
			if _, err := PlanetOccultationDetailedSVG(planetPath, options); !errors.Is(err, ErrInvalidOccultationDetailedSVGOptions) {
				t.Fatalf("planetary detailed %dx%d error = %v, want ErrInvalidOccultationDetailedSVGOptions", test.width, test.height, err)
			}
		})
	}
}

func TestOccultationDetailedSVGRejectsInvalidPathOnce(t *testing.T) {
	starPath := sampleStarOccultationPath()
	starPath.Greatest.Longitude = math.NaN()
	_, err := StarOccultationDetailedSVG(starPath, hr4799StarCoordinate(), OccultationDetailedSVGOptions{})
	if !errors.Is(err, ErrInvalidStarOccultationPath) {
		t.Fatalf("stellar detailed invalid path error = %v, want ErrInvalidStarOccultationPath", err)
	}
	if count := strings.Count(err.Error(), "invalid stellar occultation path"); count != 1 {
		t.Fatalf("stellar detailed error = %q, sentinel prefix appears %d times", err, count)
	}
	planetPath := samplePlanetOccultationPath()
	planetPath.TotalBandFootprints = []moon.PlanetOccultationFootprint{{Time: planetPath.TotalStart.Time}}
	_, err = PlanetOccultationDetailedSVG(planetPath, OccultationDetailedSVGOptions{})
	if !errors.Is(err, ErrInvalidPlanetOccultationPath) {
		t.Fatalf("planetary detailed invalid path error = %v, want ErrInvalidPlanetOccultationPath", err)
	}
	if count := strings.Count(err.Error(), "invalid planetary occultation path"); count != 1 {
		t.Fatalf("planetary detailed error = %q, sentinel prefix appears %d times", err, count)
	}
}

func TestOccultationDetailedSVGPanelValuesGolden(t *testing.T) {
	options := OccultationDetailedSVGOptions{Width: 1000, Height: 1414, Location: time.UTC}
	starDiagram, err := StarOccultationDetailedSVG(sampleStarOccultationPath(), hr4799StarCoordinate(), options)
	if err != nil {
		t.Fatalf("StarOccultationDetailedSVG() error = %v", err)
	}
	star := occultationSVGDetailedPanelRows(t, starDiagram)
	semiValues := ephemtest.NewValues(t, "occultation_detailed_panel")
	for _, expected := range []struct {
		title, label, value string
	}{
		{"月亮（地心坐标）", "赤经 R.A.", "23h37m22.4s"},
		{"月亮（地心坐标）", "赤纬 Dec.", "-00°01'34.1\""},
		{"月亮（地心坐标）", "视半径 S.D.", "00°15'25.5\""},
		{"月亮（地心坐标）", "地平视差 H.P.", "00°56'37.7\""},
		{"目标天体", "目标", "HR 4799"},
		{"目标天体", "赤经 R.A.", "12h36m47.4s"},
		{"目标天体", "赤纬 Dec.", "-05°49'55.0\""},
		{"历表与常数", "ΔT", "69.2 s"},
		{"历表与常数", "月距", "387216 km"},
		{"历表与常数", "部分掩带宽", "3300.0 km"},
		{"天平动", "经天平动 l", "-5.86°"},
		{"天平动", "纬天平动 b", "-2.79°"},
		{"天平动", "自转轴位置角 c", "-21.92°"},
	} {
		got := occultationSVGPanelValue(t, star, expected.title, expected.label)
		// 月心地心坐标随模型变化，与行星坐标行一样按模型取期望值。
		if expected.title == "月亮（地心坐标）" && strings.HasPrefix(expected.label, "赤") {
			expected.value = semiValues.WantString("moon."+expected.label, expected.value, got)
		}
		if got != expected.value {
			t.Fatalf("stellar panel %q row %q = %q, want %q", expected.title, expected.label, got, expected.value)
		}
	}

	planetDiagram, err := PlanetOccultationDetailedSVG(samplePlanetOccultationPath(), options)
	if err != nil {
		t.Fatalf("PlanetOccultationDetailedSVG() error = %v", err)
	}
	planet := occultationSVGDetailedPanelRows(t, planetDiagram)
	for _, expected := range []struct {
		title, label, value string
	}{
		{"目标天体", "目标", "土星"},
		{"目标天体", "赤经 R.A.", "00h58m06.2s"},
		{"目标天体", "赤纬 Dec.", "+03°27'41.4\""},
		{"目标天体", "视半径 S.D.", "00°00'09.3\""},
		{"历表与常数", "ΔT", "69.2 s"},
		{"历表与常数", "部分掩带宽", "3300.0 km"},
		{"历表与常数", "全掩带宽", "3100.0 km"},
		{"接触时刻", "全掩始", "10:30:00"},
		{"接触时刻", "全掩终", "12:30:00"},
	} {
		got := occultationSVGPanelValue(t, planet, expected.title, expected.label)
		// 新模型视位置更准，坐标行按模型取期望值。
		if strings.HasPrefix(expected.label, "赤") {
			expected.value = semiValues.WantString("planet."+expected.label, expected.value, got)
		}
		if got != expected.value {
			t.Fatalf("planetary panel %q row %q = %q, want %q", expected.title, expected.label, got, expected.value)
		}
	}
}

// TestOccultationDetailedSVGMoonRowsFollowGeocentricEphemeris 卡住月心地心量的量级，
// 避免把 ΔT、视差常数抄错却仍然画出"看起来正常"的面板。
func TestOccultationDetailedSVGMoonRowsFollowGeocentricEphemeris(t *testing.T) {
	diagram, err := StarOccultationDetailedSVG(sampleStarOccultationPath(), hr4799StarCoordinate(), OccultationDetailedSVGOptions{
		Width: 1000, Height: 1414, Location: time.UTC,
	})
	if err != nil {
		t.Fatalf("StarOccultationDetailedSVG() error = %v", err)
	}
	panels := occultationSVGDetailedPanelRows(t, diagram)
	if deltaT := occultationSVGNumericPrefix(t, occultationSVGPanelValue(t, panels, "历表与常数", "ΔT")); deltaT < 40 || deltaT > 110 {
		t.Fatalf("panel ΔT = %.1f s, want a 2026 value between 40 and 110 s", deltaT)
	}
	if distance := occultationSVGNumericPrefix(t, occultationSVGPanelValue(t, panels, "历表与常数", "月距")); distance < 350000 || distance > 410000 {
		t.Fatalf("panel Moon distance = %.0f km, want a physical geocentric distance", distance)
	}
	if parallax := occultationSVGSexagesimal(t, occultationSVGPanelValue(t, panels, "月亮（地心坐标）", "地平视差 H.P.")); parallax < 0.9 || parallax > 1.02 {
		t.Fatalf("panel horizontal parallax = %.4f°, want about 1° from the geocentric distance", parallax)
	}
	if semidiameter := occultationSVGSexagesimal(t, occultationSVGPanelValue(t, panels, "月亮（地心坐标）", "视半径 S.D.")); semidiameter < 0.23 || semidiameter > 0.28 {
		t.Fatalf("panel semidiameter = %.4f°, want about 0.25°", semidiameter)
	}
	if rightAscension := occultationSVGSexagesimal(t, occultationSVGPanelValue(t, panels, "月亮（地心坐标）", "赤经 R.A.")); rightAscension < 350 || rightAscension > 360 {
		t.Fatalf("panel Moon R.A. = %.3f°, want the same geocentric direction as the target star", rightAscension)
	}
}
