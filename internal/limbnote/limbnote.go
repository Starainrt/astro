// Package limbnote 生成出图时写入图注的月面口径声明。
package limbnote

import "strings"

const english = "en"

// Contacts 接触时刻所用月面口径的短标签，保持站心图页脚平衡 / short tag for the contact-time limb convention.
func Contacts(applied bool, language string) string {
	if applied {
		if language == english {
			return "Limb profile."
		}
		return "月缘剖面。"
	}
	if language == english {
		return "Circular limb."
	}
	return "圆月面。"
}

// Limits 界线所用月面口径；可见域与等时线始终按圆月面 / limb convention for limit tracks only.
func Limits(applied bool, language string) string {
	if applied {
		if language == english {
			return "Limits use the lunar limb profile (areas and isolines circular)."
		}
		return "界线按月缘剖面修正（可见域与等时线仍按圆月面）。"
	}
	if language == english {
		return "Circular lunar limb."
	}
	return "月面按圆月面计算。"
}

// SolarMapLayers 日食全球图各层是否真正按月缘剖面修正 / which global solar map layers the lunar limb actually corrected.
type SolarMapLayers struct {
	// GlobalContacts 全球偏食始末与食甚点中心食时长 / global P1/P4 and the central duration at greatest eclipse.
	GlobalContacts bool
	// LocalPanel 食甚点站心面板 / the local panel at greatest eclipse.
	LocalPanel bool
	// CentralLimits 中心带南北界与带宽 / central-band limits and width.
	CentralLimits bool
	// PartialBoundary 偏食可见域外边界 / the outer partial visibility boundary.
	PartialBoundary bool
}

// SolarMap 按实际生效的层列出日食全球图的月缘修正，食分线与等时线始终按圆月面 / lists the corrected layers of the global solar map.
func SolarMap(layers SolarMapLayers, language string) string {
	type layer struct {
		applied     bool
		zh, english string
	}
	var names []string
	for _, item := range []layer{
		{layers.GlobalContacts, "全球始末", "global contacts"},
		{layers.LocalPanel, "食甚点站心值", "local values at greatest eclipse"},
		{layers.CentralLimits, "中心带界线", "central limits"},
		{layers.PartialBoundary, "可见域边界", "visibility boundary"},
	} {
		if !item.applied {
			continue
		}
		if language == english {
			names = append(names, item.english)
		} else {
			names = append(names, item.zh)
		}
	}
	if len(names) == 0 {
		return Limits(false, language)
	}
	if language == english {
		return "Lunar limb profile applied to " + strings.Join(names, ", ") + " (isolines circular)."
	}
	return "按月缘剖面修正" + strings.Join(names, "、") + "（食分线与等时线仍按圆月面）。"
}

// Join 把月面口径声明接在时标声明之后，二者同占一行 / appends the limb declaration to the time-scale note.
func Join(scaleNote, limbNote, language string) string {
	switch {
	case scaleNote == "":
		return limbNote
	case limbNote == "":
		return scaleNote
	case strings.HasSuffix(scaleNote, "。") || strings.HasSuffix(scaleNote, "."):
		if language == english {
			return scaleNote + " " + limbNote
		}
		return scaleNote + limbNote
	case language == english:
		return scaleNote + ". " + limbNote
	default:
		return scaleNote + "；" + limbNote
	}
}
