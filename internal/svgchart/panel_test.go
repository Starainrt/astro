package svgchart

import (
	"math"
	"strings"
	"testing"
)

func TestEstimatedTextWidthUsesSeparateAsciiAndWideRuneUnits(t *testing.T) {
	if got := EstimatedTextWidth("", 10); got != 0 {
		t.Fatalf("empty text width = %v, want 0", got)
	}
	if got := EstimatedTextWidth("AB", 10); math.Abs(got-12.4) > 1e-9 {
		t.Fatalf("two ASCII runes = %v, want 12.4", got)
	}
	if got := EstimatedTextWidth("中文", 10); math.Abs(got-21) > 1e-9 {
		t.Fatalf("two wide runes = %v, want 21", got)
	}
	above, below := EstimatedTextExtents(10)
	if math.Abs(above-8.5) > 1e-9 || math.Abs(below-3) > 1e-9 {
		t.Fatalf("extents = %v/%v, want 8.5/3", above, below)
	}
}

func TestPanelRowFontSizeShrinksUntilLabelAndValueFit(t *testing.T) {
	row := PanelRow{Label: "地平视差 H.P.", Value: "00°16'21.4\""}
	if got := PanelRowFontSize(320, row); got != panelRowLabelSize {
		t.Fatalf("wide box font = %v, want the base size", got)
	}
	narrow := 170.0
	size := PanelRowFontSize(narrow, row)
	if size >= panelRowLabelSize || size < panelMinimumFontSize {
		t.Fatalf("narrow box font = %v, want a size in [%v,%v)", size, panelMinimumFontSize, panelRowLabelSize)
	}
	content := narrow - 2*panelBoxPadding
	pair := EstimatedTextWidth(row.Label, size) + EstimatedTextWidth(row.Value, size)
	if pair+panelLabelValueGap > content {
		t.Fatalf("shrunk label %v + value %v + gap does not fit %v", pair, panelLabelValueGap, content)
	}
}

func TestPanelBoxOverlapsDetectsCrowdedRows(t *testing.T) {
	rows := []PanelRow{
		{Label: "赤经 R.A.", Value: "23h21m52.2s"},
		{Label: "赤纬 Dec.", Value: "-05°59'26.5\""},
		{Label: "掩甚", Value: "04:28:23", Detail: "90.9407°E, 11.7283°N"},
		{Label: "视半径 S.D.", Value: "00°16'21.4\""},
	}
	if PanelBoxOverlaps(0, 0, 230, 200, "月亮（地心坐标）", rows) {
		t.Fatal("tall box reported overlapping rows")
	}
	if !PanelBoxOverlaps(0, 0, 230, 34, "月亮（地心坐标）", rows) {
		t.Fatal("box too short for its rows reported no overlap")
	}
	if !PanelBoxOverlaps(0, 0, 120, 200, "月亮（地心坐标）", rows) {
		t.Fatal("box too narrow for label and value reported no overlap")
	}
	if PanelBoxOverlaps(0, 0, 230, 200, "标题", nil) {
		t.Fatal("empty block reported an overlap")
	}
}

func TestWritePanelBoxKeepsEveryTextInsideTheBlockWidth(t *testing.T) {
	row := PanelRow{Label: "地平视差 H.P.", Value: "00°16'21.4\""}
	var builder strings.Builder
	WritePanelBox(&builder, "panel", 0, 0, 170, 200, "月亮（地心坐标）", []PanelRow{row})
	document := builder.String()
	labelX := strings.Index(document, `x="12.000"`)
	if labelX < 0 {
		t.Fatalf("label is not drawn at the block padding: %s", document)
	}
	if !strings.Contains(document, `text-anchor="end"`) {
		t.Fatalf("value is not right-anchored: %s", document)
	}
}
