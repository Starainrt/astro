package svgchart

import (
	"reflect"
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

func TestFooterLinesWithScaleRespectsLineLimit(t *testing.T) {
	notes := []string{"说明一"}
	scale := "图中时刻为 UTC"
	for _, test := range []struct {
		name      string
		maxLines  int
		scaleNote string
		want      []string
	}{
		{name: "no lines", maxLines: 0, scaleNote: scale},
		{name: "negative lines", maxLines: -1, scaleNote: scale},
		{name: "scale only", maxLines: 1, scaleNote: scale, want: []string{scale}},
		{name: "notes and scale", maxLines: 2, scaleNote: scale, want: []string{"说明一", scale}},
		{name: "notes without scale", maxLines: 1, want: []string{"说明一"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := FooterLinesWithScale(notes, test.scaleNote, 400, 11, test.maxLines)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("lines = %#v, want %#v", got, test.want)
			}
			if test.maxLines > 0 && len(got) > test.maxLines {
				t.Fatalf("lines = %d, want no more than %d", len(got), test.maxLines)
			}
		})
	}
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

// 均衡折行只改断行位置：行数不超过同宽度下的 WrapText，行宽互相接近，且逐行仍在 maxWidth 内。
// textRunes 拼掉断行处被吃掉的空白，用于比较两种折行的字符内容。
func textRunes(lines []string) string {
	return strings.Join(strings.Fields(strings.Join(lines, " ")), "")
}

func textSpread(lines []string, fontSize float64) float64 {
	minimum, maximum := 0.0, 0.0
	for index, line := range lines {
		width := EstimatedTextWidth(line, fontSize)
		if index == 0 || width < minimum {
			minimum = width
		}
		if index == 0 || width > maximum {
			maximum = width
		}
	}
	if minimum <= 0 {
		return 0
	}
	return maximum / minimum
}

func TestWrapTextBalancedKeepsLineCountAndWidth(t *testing.T) {
	for _, test := range []struct {
		name       string
		value      string
		maxWidth   float64
		fontSize   float64
		sameCount  bool
		wantSpread bool
	}{
		{name: "english prose", maxWidth: 500, fontSize: 10, sameCount: true, wantSpread: true,
			value: "Moon fixed at center; east is left and north is up. The blue-gray dashed line is the local lunar path; the red dashed line is the star track."},
		{name: "chinese with one space", maxWidth: 420, fontSize: 12, wantSpread: true,
			value: "上方为全局路径，C2/C3 只标点位；下方为各阶段独立视圆图。接触点位置角从天球北点起向东量。 图中时刻为 UTC（显示时区 UTC+08:00）"},
		{name: "mixed page", maxWidth: 300, fontSize: 9,
			value: strings.Repeat("月球掩星 occultation 说明；", 8)},
	} {
		t.Run(test.name, func(t *testing.T) {
			greedy := WrapText(test.value, test.maxWidth, test.fontSize)
			balanced := WrapTextBalanced(test.value, test.maxWidth, test.fontSize)
			if len(balanced) > len(greedy) {
				t.Fatalf("balanced lines = %d, want no more than the greedy %d", len(balanced), len(greedy))
			}
			if test.sameCount && len(balanced) != len(greedy) {
				t.Fatalf("balanced lines = %d, want the greedy %d", len(balanced), len(greedy))
			}
			if len(balanced) < 2 {
				t.Fatalf("balanced lines = %d, want the text to be wrapped", len(balanced))
			}
			textFits(t, balanced, test.maxWidth, test.fontSize)
			if got, want := textRunes(balanced), textRunes(greedy); got != want {
				t.Fatalf("balanced text = %q, want the same runes as %q", got, want)
			}
			spread := textSpread(balanced, test.fontSize)
			if test.sameCount && spread > textSpread(greedy, test.fontSize) {
				t.Fatalf("balanced spread = %.3f, want no worse than the greedy %.3f (%#v)",
					spread, textSpread(greedy, test.fontSize), balanced)
			}
			if test.wantSpread && spread > 1.15 {
				t.Fatalf("balanced spread = %.3f, want <= 1.15 (%#v)", spread, balanced)
			}
		})
	}
}

// 中文里只有一个空格时，WrapText 会为躲开西文词把首行回退到该空格，WrapTextBalanced 允许在全角字符前断开，不再留下短行。
func TestWrapTextBalancedBreaksFullWidthTextWithoutShortTail(t *testing.T) {
	value := "上方为全局路径，C2/C3 只标点位；下方为各阶段独立视圆图。接触点位置角从天球北点起向东量。 图中时刻为 UTC（显示时区 UTC+08:00）"
	greedy := WrapText(value, 420, 12)
	if len(greedy) < 3 {
		t.Fatalf("greedy lines = %#v, want the whitespace backoff to cost extra lines", greedy)
	}
	if width := EstimatedTextWidth(greedy[0], 12); width > 0.5*420 {
		t.Fatalf("greedy first line = %q (%.1f), want the short backoff line", greedy[0], width)
	}
	balanced := WrapTextBalanced(value, 420, 12)
	if len(balanced) >= len(greedy) {
		t.Fatalf("balanced lines = %d, want fewer than the greedy %d", len(balanced), len(greedy))
	}
	textFits(t, balanced, 420, 12)
	if len(balanced) != 2 || textSpread(balanced, 12) > 1.15 {
		t.Fatalf("balanced = %#v, want two lines of close width", balanced)
	}
}

func TestWrapTextBalancedLinesFillsSlots(t *testing.T) {
	value := strings.Repeat("月球掩星 occultation 说明；", 2)
	filled := WrapTextBalancedLines(value, 400, 9, 2)
	if len(filled) != 2 {
		t.Fatalf("filled lines = %d, want 2", len(filled))
	}
	textFits(t, filled, 400, 9)
	if spread := textSpread(filled, 9); spread > 1.15 {
		t.Fatalf("filled spread = %.3f, want <= 1.15 (%#v)", spread, filled)
	}
	if natural := WrapTextBalanced(value, 400, 9); len(natural) != 1 {
		t.Fatalf("natural lines = %d, want the natural wrap to keep one line", len(natural))
	}
	if lines := WrapTextBalancedLines(value, 400, 9, 0); len(lines) != 1 {
		t.Fatalf("lines = %d, want maxLines 0 to fall back to the natural wrap", len(lines))
	}
	// 槽位放不下时不强行行数：返回可用宽度下的折行结果，由调用方截断。
	long := strings.Repeat("月球掩星 occultation 说明；", 12)
	if lines := WrapTextBalancedLines(long, 400, 9, 2); len(lines) < 2 {
		t.Fatalf("lines = %d, want the width-limited wrap", len(lines))
	}
}
