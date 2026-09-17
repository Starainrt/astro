package svgchart

import (
	"math"
	"strings"
	"unicode"
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
		if end < len(runes) && lastSpace > 0 {
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
