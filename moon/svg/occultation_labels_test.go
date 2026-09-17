package svg

import (
	"testing"
	"time"
)

// 本文件量化三族地图标签（等时线、中心线时刻、事件标记）的互相避让：
// 产物里任一族的 <text> 占位都不许与另一族相交。

type occultationLabelFixture struct {
	name   string
	width  int
	height int
	// labelStep 为 0 时用默认 30 分钟。
	labelStep time.Duration
	star      bool
}

func TestOccultationMapLabelsDoNotOverlap(t *testing.T) {
	fixtures := []occultationLabelFixture{
		{name: "stellar 1200x800", width: 1200, height: 800, labelStep: 10 * time.Minute, star: true},
		{name: "stellar 720x520", width: 720, height: 520, labelStep: 15 * time.Minute, star: true},
		{name: "stellar 640x480", width: 640, height: 480, star: true},
		{name: "planetary compact 1200x800", width: 1200, height: 800, labelStep: 10 * time.Minute},
		{name: "planetary compact 720x520", width: 720, height: 520, labelStep: 15 * time.Minute},
	}
	totalLabels := 0
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			options := StarOccultationSVGOptions{
				Width: fixture.width, Height: fixture.height, Location: time.UTC, TimeLabelStep: fixture.labelStep,
			}
			var diagram string
			var err error
			if fixture.star {
				diagram, err = StarOccultationPathSVG(occultationTestStarPath(t), options)
			} else {
				diagram, err = PlanetOccultationPathSVG(occultationTestSaturnPath(t), options)
			}
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			labels := occultationSVGLabelTexts(occultationSVGTexts(t, diagram))
			if len(labels) < 3 {
				t.Fatalf("fixture produced only %d labels; the overlap check would be vacuous", len(labels))
			}
			if outside := occultationSVGTextsOutsideCanvas(labels, float64(fixture.width), float64(fixture.height)); len(outside) > 0 {
				t.Fatalf("map labels outside the canvas: %v", outside)
			}
			if pairs := occultationSVGTextOverlapPairs(labels); pairs != 0 {
				t.Fatalf("%d label overlap pairs among %d labels", pairs, len(labels))
			}
			totalLabels += len(labels)
		})
	}
	if totalLabels < 40 {
		t.Fatalf("label fixtures only produced %d labels, want a dense label set", totalLabels)
	}
	t.Logf("checked %d labels across %d fixtures", totalLabels, len(fixtures))
}

func TestPlanetOccultationMapLabelsAvoidTotalBandMarkers(t *testing.T) {
	diagram, err := PlanetOccultationPathSVG(occultationTestMarsPath(t), StarOccultationSVGOptions{
		Width: 1200, Height: 800, Location: time.UTC, TimeLabelStep: 10 * time.Minute,
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	labels := occultationSVGLabelTexts(occultationSVGTexts(t, diagram))
	if len(labels) < 5 {
		t.Fatalf("total-band fixture produced only %d labels", len(labels))
	}
	if pairs := occultationSVGTextOverlapPairs(labels); pairs != 0 {
		t.Fatalf("%d label overlap pairs among %d labels", pairs, len(labels))
	}
	markers := 0
	for _, label := range labels {
		if label.parentClass == "event-marker event-greatest" {
			markers++
		}
	}
	if markers != 1 {
		t.Fatalf("greatest-event label count = %d, want one", markers)
	}
}
