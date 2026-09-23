package svgchart

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ellipsisMark 是截断标记；用单个全角省略号，避免与文本里的三个半角点混淆。
const ellipsisMark = "…"

// estimatedTextRuneWidth 与 EstimatedTextWidth 的逐字符口径必须一致，否则折行结果会与占位判定矛盾。
func estimatedTextRuneWidth(symbol rune, fontSize float64) float64 {
	if symbol < 128 {
		return 0.62 * fontSize
	}
	return 1.05 * fontSize
}

// EllipsizeText 把单行文本裁到 maxWidth 内：放不下时保留前缀并追加省略号，全空的返回值表示连省略号都放不下。
// EllipsizeText clamps one line to maxWidth, appending an ellipsis when the text does not fit;
// an empty result means even the ellipsis does not fit.
func EllipsizeText(value string, maxWidth, fontSize float64) string {
	if EstimatedTextWidth(value, fontSize) <= maxWidth {
		return value
	}
	return truncateWithMark(value, maxWidth, fontSize)
}

// truncateWithMark 无条件追加省略号：被截断的行即使本身放得下，也必须标出后面还有内容。
func truncateWithMark(value string, maxWidth, fontSize float64) string {
	markWidth := EstimatedTextWidth(ellipsisMark, fontSize)
	if markWidth > maxWidth {
		return ""
	}
	runes := []rune(value)
	for len(runes) > 0 && EstimatedTextWidth(string(runes), fontSize)+markWidth > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return strings.TrimRight(string(runes), " ") + ellipsisMark
}

// WrapText 按 EstimatedTextWidth 保守折行：优先在空白处断开，单个词或全角文本按字符断开，空行不返回。
// WrapText wraps conservatively by EstimatedTextWidth, preferring whitespace breaks; empty lines are dropped.
func WrapText(value string, maxWidth, fontSize float64) []string {
	return wrapText(value, maxWidth, fontSize, false)
}

// wrapText 是 WrapText 与 WrapTextBalanced 共用的折行核心；fullWidthBreak 为真时允许在全角字符前断开，
// 只有落在 ASCII 词内部才回退到空白，避免中文夹英文的行被空白回退压成短行。
func wrapText(value string, maxWidth, fontSize float64, fullWidthBreak bool) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if EstimatedTextWidth(value, fontSize) <= maxWidth {
		return []string{value}
	}
	runes := []rune(value)
	lines := make([]string, 0, 4)
	for len(runes) > 0 {
		width := 0.0
		end := 0
		lastSpace := -1
		for end < len(runes) {
			nextWidth := width + estimatedTextRuneWidth(runes[end], fontSize)
			if nextWidth > maxWidth && end > 0 {
				break
			}
			width = nextWidth
			if unicode.IsSpace(runes[end]) {
				lastSpace = end
			}
			end++
		}
		if end < len(runes) && lastSpace > 0 && !(fullWidthBreak && runes[end] >= utf8.RuneSelf) {
			end = lastSpace
		}
		if end == 0 {
			end = 1
		}
		if line := strings.TrimSpace(string(runes[:end])); line != "" {
			lines = append(lines, line)
		}
		runes = runes[end:]
		for len(runes) > 0 && unicode.IsSpace(runes[0]) {
			runes = runes[1:]
		}
	}
	return lines
}

// WrapTextBalanced 均衡折行：行数取 maxWidth 下 WrapText 的行数 n，再二分出仍能压到 n 行的最小宽度折行，使各行宽度接近；
// 折行宽度不比 maxWidth 宽，逐行都不超过 maxWidth。全角字符前允许硬断，因此结果可能少于 n 行（中文夹空格时段落尾部会只剩一个短桩）。
// 只影响断行位置，不改 WrapText 本身的口径。
// WrapTextBalanced keeps the greedy line count for maxWidth and narrows the break width until the lines are close in width.
// It may return fewer lines than WrapText because it may break before a full-width rune.
func WrapTextBalanced(value string, maxWidth, fontSize float64) []string {
	if maxWidth <= 0 {
		return WrapText(value, maxWidth, fontSize)
	}
	return balancedTextLines(value, maxWidth, fontSize, len(wrapText(value, maxWidth, fontSize, true)))
}

// WrapTextBalancedLines 均衡折行到 maxLines 行（maxLines 非正时等同 WrapTextBalanced），供只有固定行数槽位的文本块使用；
// maxWidth 下压不到 maxLines 行时返回放得下的最少行数，由调用方截断。
// WrapTextBalancedLines balances the text into maxLines slots when maxWidth allows it.
func WrapTextBalancedLines(value string, maxWidth, fontSize float64, maxLines int) []string {
	if maxWidth <= 0 {
		return WrapText(value, maxWidth, fontSize)
	}
	if maxLines <= 0 {
		return WrapTextBalanced(value, maxWidth, fontSize)
	}
	return balancedTextLines(value, maxWidth, fontSize, maxLines)
}

// balancedTextLines 二分出仍能压到 targetLines 行的最小折行宽度并折行；行数随折行宽度单调不增，故二分成立。
func balancedTextLines(value string, maxWidth, fontSize float64, targetLines int) []string {
	if targetLines < 2 {
		return wrapText(value, maxWidth, fontSize, true)
	}
	low, high := 0.0, maxWidth
	for iteration := 0; iteration < 40; iteration++ {
		middle := (low + high) / 2
		if len(wrapText(value, middle, fontSize, true)) > targetLines {
			low = middle
			continue
		}
		high = middle
	}
	return wrapText(value, high, fontSize, true)
}

// TextLineLimit 返回高度预算 maxHeight 能容纳的行数，至少一行；口径与 EstimatedTextExtents 一致。
// TextLineLimit reports how many lines fit in maxHeight, never less than one.
func TextLineLimit(fontSize, lineHeight, maxHeight float64) int {
	if lineHeight <= 0 {
		lineHeight = 1.2 * fontSize
	}
	above, below := EstimatedTextExtents(fontSize)
	limit := int(math.Floor((maxHeight-above-below)/lineHeight)) + 1
	if limit < 1 {
		return 1
	}
	return limit
}

// BaselineLineLimit 返回自 firstBaseline 起、底线 bottomLimit 之前能排下的行数，至少一行。
// BaselineLineLimit reports how many lines fit below firstBaseline and above bottomLimit, never less than one.
func BaselineLineLimit(fontSize, lineHeight, firstBaseline, bottomLimit float64) int {
	above, _ := EstimatedTextExtents(fontSize)
	return TextLineLimit(fontSize, lineHeight, bottomLimit-firstBaseline+above)
}

// FooterLineHeight 返回页脚行距：按 1.5 倍字号取整，12px 字得 18px、11px 字得 17px。
// FooterLineHeight returns the footer line step, 1.5x the font size, rounded.
func FooterLineHeight(fontSize float64) float64 {
	return math.Round(fontSize * 1.5)
}

// FooterBottomPadding 返回页脚末行基线距画布底边的留白：按 1.35 倍字号取整，12px 字得 16px、11px 字得 15px。
// FooterBottomPadding returns the gap between the last footer baseline and the canvas bottom.
func FooterBottomPadding(fontSize float64) float64 {
	return math.Round(fontSize * 1.35)
}

// FooterBlock 把页脚各行按统一行距自画布底边往上排：末行基线距底边 padding，其上每行递增 lineHeight；
// 返回各行基线与整块自底边起占用的高度（含首行字高），供上方内容留位。
// FooterBlock lays footer lines bottom-up with one uniform step and returns their baselines
// plus the height the block occupies above the canvas bottom.
func FooterBlock(lineCount int, height, fontSize, lineHeight, padding float64) (baselines []float64, occupied float64) {
	if lineCount < 1 {
		lineCount = 1
	}
	above, _ := EstimatedTextExtents(fontSize)
	baselines = make([]float64, lineCount)
	for index := range baselines {
		baselines[index] = height - padding - float64(lineCount-1-index)*lineHeight
	}
	return baselines, padding + float64(lineCount-1)*lineHeight + above
}

// FooterLinesWithScale 在说明行之后接上时标声明：说明最多占 maxLines-1 行，声明独占最后一行、不参与截断；
// 说明行数超过 maxLines 时先截断说明，再长的自定义说明也挤不掉时标声明。
// FooterLinesWithScale keeps the time-scale declaration on its own final line, outside truncation.
func FooterLinesWithScale(notes []string, scaleNote string, maxWidth, fontSize float64, maxLines int) []string {
	if maxLines <= 0 {
		return nil
	}
	if scaleNote == "" {
		return TruncateTextLines(notes, maxWidth, fontSize, maxLines)
	}
	if maxLines == 1 {
		return []string{scaleNote}
	}
	return append(TruncateTextLines(notes, maxWidth, fontSize, maxLines-1), scaleNote)
}

// TruncateTextLines 按行数上限截断折行结果：截断时末行以省略号结尾并重新裁到 maxWidth 内，未截断时原样返回。
// TruncateTextLines cuts lines to maxLines; a cut last line is ellipsized to maxWidth.
func TruncateTextLines(lines []string, maxWidth, fontSize float64, maxLines int) []string {
	if len(lines) == 0 || maxLines <= 0 || len(lines) <= maxLines {
		return lines
	}
	kept := make([]string, maxLines)
	copy(kept, lines[:maxLines])
	kept[maxLines-1] = truncateWithMark(kept[maxLines-1], maxWidth, fontSize)
	return kept
}
