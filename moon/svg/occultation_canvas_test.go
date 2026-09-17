package svg

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

// 本文件钉住全球掩带图的画布下限与"所有文本都在画布内"：480x360、600x400 曾因图例压页脚、经纬刻度出框而被放行。

func TestOccultationMapRejectsCanvasBelowMinimum(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
	}{
		{name: "reference too small", width: 480, height: 360},
		{name: "legend footer collision", width: 600, height: 400},
		{name: "width just below", width: 639, height: 480},
		{name: "height just below", width: 640, height: 479},
	}
	starPath := occultationTestStarPath(t)
	planetPath := occultationTestSaturnPath(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := StarOccultationSVGOptions{Width: test.width, Height: test.height}
			if _, err := StarOccultationPathSVG(starPath, options); !errors.Is(err, ErrInvalidStarOccultationSVGOptions) {
				t.Fatalf("stellar %dx%d error = %v, want ErrInvalidStarOccultationSVGOptions", test.width, test.height, err)
			}
			if _, err := PlanetOccultationPathSVG(planetPath, options); !errors.Is(err, ErrInvalidPlanetOccultationSVGOptions) {
				t.Fatalf("planetary %dx%d error = %v, want ErrInvalidPlanetOccultationSVGOptions", test.width, test.height, err)
			}
		})
	}
}

func TestOccultationMapKeepsEveryTextInsideCanvas(t *testing.T) {
	sizes := [][2]int{{640, 480}, {800, 600}, {1200, 800}, {1000, 1414}}
	projections := []MapProjection{MapProjectionAuto, MapProjectionNorthPolar, MapProjectionOrthographic}
	starPath := occultationTestStarPath(t)
	planetPath := occultationTestSaturnPath(t)
	for _, size := range sizes {
		for _, projection := range projections {
			name := strconv.Itoa(size[0]) + "x" + strconv.Itoa(size[1]) + "-" + string(projection)
			t.Run(name, func(t *testing.T) {
				options := StarOccultationSVGOptions{
					Width: size[0], Height: size[1], Location: time.UTC, Projection: projection,
				}
				diagrams := map[string]string{}
				star, err := StarOccultationPathSVG(starPath, options)
				if err != nil {
					t.Fatalf("stellar render: %v", err)
				}
				diagrams["stellar"] = star
				planet, err := PlanetOccultationPathSVG(planetPath, options)
				if err != nil {
					t.Fatalf("planetary render: %v", err)
				}
				diagrams["planetary"] = planet
				for label, diagram := range diagrams {
					texts := occultationSVGTexts(t, diagram)
					if len(texts) == 0 {
						t.Fatalf("%s SVG has no text", label)
					}
					if outside := occultationSVGTextsOutsideCanvas(texts, float64(size[0]), float64(size[1])); len(outside) > 0 {
						t.Fatalf("%s SVG has %d texts outside %s: %v", label, len(outside), name, outside)
					}
				}
			})
		}
	}
}

func TestOccultationMapLegendDoesNotOverlapFooter(t *testing.T) {
	sizes := [][2]int{{640, 480}, {800, 600}, {1200, 800}}
	starPath := occultationTestStarPath(t)
	planetPath := occultationTestSaturnPath(t)
	for _, size := range sizes {
		t.Run(strconv.Itoa(size[0])+"x"+strconv.Itoa(size[1]), func(t *testing.T) {
			options := StarOccultationSVGOptions{Width: size[0], Height: size[1], Location: time.UTC}
			for _, item := range []struct {
				label string
				path  func() (string, error)
			}{
				{label: "stellar", path: func() (string, error) { return StarOccultationPathSVG(starPath, options) }},
				{label: "planetary", path: func() (string, error) { return PlanetOccultationPathSVG(planetPath, options) }},
			} {
				diagram, err := item.path()
				if err != nil {
					t.Fatalf("%s render: %v", item.label, err)
				}
				texts := occultationSVGTexts(t, diagram)
				legend := []occultationSVGText{}
				footer := []occultationSVGText{}
				for _, text := range texts {
					if text.class == "occultation-legend-text" {
						legend = append(legend, text)
					}
					if text.parent == "svg" && text.class == "" && text.fontSize == 11 && text.x == 40 {
						footer = append(footer, text)
					}
				}
				if len(legend) < 3 || len(footer) == 0 {
					t.Fatalf("%s SVG legend=%d footer=%d, want both present", item.label, len(legend), len(footer))
				}
				for _, legendText := range legend {
					for _, footerText := range footer {
						if occultationSVGTextBox(legendText)[3] > occultationSVGTextBox(footerText)[1] {
							t.Fatalf("%s legend %q overlaps footer %q at %dx%d",
								item.label, legendText.value, footerText.value, size[0], size[1])
						}
					}
				}
			}
		})
	}
}
