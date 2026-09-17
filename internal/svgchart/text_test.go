package svgchart

import (
	"strings"
	"testing"
)

// 折行与截断必须与 EstimatedTextWidth/EstimatedTextExtents 同口径：留下的每一行都不超宽，
// 截断后的行数不超过高度预算，且末行以省略号结尾。

func textFits(t *testing.T, lines []string, maxWidth, fontSize float64) {
	t.Helper()
	for _, line := range lines {
		if width := EstimatedTextWidth(line, fontSize); width > maxWidth {
			t.Fatalf("line %q width %.1f exceeds %.1f", line, width, maxWidth)
		}
	}
}

func TestWrapTextKeepsEveryLineInsideWidth(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		maxWidth float64
		fontSize float64
	}{
		{name: "ascii words", value: strings.Repeat("alpha beta gamma ", 20), maxWidth: 220, fontSize: 11},
		{name: "chinese runs", value: strings.Repeat("月球掩星示意图说明", 30), maxWidth: 180, fontSize: 11},
		{name: "mixed", value: strings.Repeat("掩星 occultation 说明；", 40), maxWidth: 260, fontSize: 9},
		{name: "single word", value: strings.Repeat("A", 200), maxWidth: 120, fontSize: 10},
	} {
		t.Run(test.name, func(t *testing.T) {
			lines := WrapText(test.value, test.maxWidth, test.fontSize)
			if len(lines) < 2 {
				t.Fatalf("lines = %d, want the text to be wrapped", len(lines))
			}
			textFits(t, lines, test.maxWidth, test.fontSize)
		})
	}
}

func TestWrapTextKeepsShortTextOnOneLine(t *testing.T) {
	lines := WrapText("  短说明 short note  ", 400, 11)
	if len(lines) != 1 || lines[0] != "短说明 short note" {
		t.Fatalf("lines = %#v, want one trimmed line", lines)
	}
	if lines := WrapText("   ", 400, 11); lines != nil {
		t.Fatalf("blank text lines = %#v, want nil", lines)
	}
}

func TestTruncateTextLinesMarksTheCut(t *testing.T) {
	lines := []string{"第一行", "第二行", "第三行", "第四行"}
	kept := TruncateTextLines(lines, 400, 11, 2)
	if len(kept) != 2 || kept[0] != "第一行" {
		t.Fatalf("kept = %#v, want the first two lines", kept)
	}
	if !strings.HasSuffix(kept[1], ellipsisMark) {
		t.Fatalf("last kept line = %q, want the ellipsis mark", kept[1])
	}
	if kept := TruncateTextLines(lines, 400, 11, 9); len(kept) != 4 || kept[3] != "第四行" {
		t.Fatalf("kept = %#v, want the untouched input", kept)
	}
}

func TestTruncateTextLinesEllipsizesTheLastLineToWidth(t *testing.T) {
	lines := []string{strings.Repeat("宽", 20), strings.Repeat("宽", 20)}
	kept := TruncateTextLines(lines, 60, 10, 1)
	if len(kept) != 1 || !strings.HasSuffix(kept[0], ellipsisMark) {
		t.Fatalf("kept = %#v, want one ellipsized line", kept)
	}
	textFits(t, kept, 60, 10)
}

func TestTextLineLimitCountsExtents(t *testing.T) {
	// 3 行 11 号字、行距 15 的总高约 42.65，预算再少一点就只放得下 2 行。
	if limit := TextLineLimit(11, 15, 42.7); limit != 3 {
		t.Fatalf("limit = %d, want 3", limit)
	}
	if limit := TextLineLimit(11, 15, 42.6); limit != 2 {
		t.Fatalf("limit = %d, want 2", limit)
	}
	if limit := TextLineLimit(11, 15, 1); limit != 1 {
		t.Fatalf("limit = %d, want the one-line floor", limit)
	}
	// 基线口径：首行基线 100，第 2 行要排得下需要底线不低于 100+15+0.3*11。
	if limit := BaselineLineLimit(11, 15, 100, 118.4); limit != 2 {
		t.Fatalf("limit = %d, want 2", limit)
	}
	if limit := BaselineLineLimit(11, 15, 100, 118.2); limit != 1 {
		t.Fatalf("limit = %d, want 1", limit)
	}
}

func TestEllipsizeTextStaysInsideWidth(t *testing.T) {
	value := strings.Repeat("长文本 ellipsis ", 30)
	clamped := EllipsizeText(value, 150, 11)
	if !strings.HasSuffix(clamped, ellipsisMark) {
		t.Fatalf("clamped = %q, want the ellipsis mark", clamped)
	}
	textFits(t, []string{clamped}, 150, 11)
	if short := EllipsizeText("短", 150, 11); short != "短" {
		t.Fatalf("short text = %q, want it untouched", short)
	}
	if empty := EllipsizeText(value, 1, 11); empty != "" {
		t.Fatalf("clamped = %q, want an empty string when even the mark does not fit", empty)
	}
}
