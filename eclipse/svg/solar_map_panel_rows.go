package svg

import (
	"fmt"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgchart"
)

// 本文件的函数只负责把日食数据变成面板行；画格由 writeEclipsePanelBox 统一处理，
// 竖版与横版两种排布共用同一批行，避免各写一遍。

func solarEclipseContactTitles(options SolarEclipseMapSVGOptions) (penumbra, circumstances, umbra string) {
	if options.Language == "en" {
		penumbra, circumstances, umbra = "Penumbral contacts", "Local circumstances at greatest", "Umbral contacts"
	} else {
		penumbra, circumstances, umbra = "半影外切 / 内切接触", "食甚点的地方情况", "本影外切 / 内切接触"
	}
	// 全球阶段量（食甚经纬度、中心食始/终）就在这一块里，它的标题可由 EventsTitle 覆盖。
	if options.EventsTitle != "" {
		circumstances = options.EventsTitle
	}
	return penumbra, circumstances, umbra
}

func solarEclipseEphemerisTitles(options SolarEclipseMapSVGOptions) (ephemeris, libration string) {
	if options.Language == "en" {
		return "Ephemeris and constants", "Libration"
	}
	return "历表与常数", "天平动"
}

func solarEclipseGeocentricTitles(options SolarEclipseMapSVGOptions) (sun, moon string) {
	if options.Language == "en" {
		return "Sun at greatest eclipse", "Moon at greatest eclipse"
	}
	return "食甚时的太阳（地心坐标）", "食甚时的月亮（地心坐标）"
}

func solarEclipseContactRows(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	local eclipsecore.LocalSolarEclipseInfo,
	hasLocal bool,
	options SolarEclipseMapSVGOptions,
) (penumbra, umbra, circumstances []svgchart.PanelRow) {
	english := options.Language == "en"
	name := func(zh, en string) string {
		if english {
			return en
		}
		return zh
	}
	contact := func(label string, point eclipsecore.SolarEclipsePathPoint) svgchart.PanelRow {
		if point.Time.IsZero() {
			return svgchart.PanelRow{}
		}
		return svgchart.PanelRow{Label: label, Value: point.Time.In(options.Location).Format("15:04:05")}
	}
	penumbra = []svgchart.PanelRow{
		contact("P1 "+name("半影外切", "partial begins"), partial.P1),
		contact("P2 "+name("半影内切", "internal contact"), partial.P2),
		contact("P3 "+name("半影内切", "internal contact"), partial.P3),
		contact("P4 "+name("半影外切", "partial ends"), partial.P4),
	}
	umbra = []svgchart.PanelRow{
		contact("U1 "+name("本影外切", "umbra begins"), partial.U1),
		contact("U2 "+name("本影内切", "internal contact"), partial.U2),
		contact("U3 "+name("本影内切", "internal contact"), partial.U3),
		contact("U4 "+name("本影外切", "umbra ends"), partial.U4),
	}
	info := partial.Eclipse
	// 非中心食没有中心线，带宽与中心食时长无从谈起，留空而不是写 0。
	pathWidth, centralDuration := "—", "—"
	if info.HasCentral {
		pathWidth = fmt.Sprintf("%.1f km", info.PathWidthKM)
		centralDuration = formatSolarEclipseMapDuration(info.CentralDuration)
	}
	circumstances = []svgchart.PanelRow{
		{Label: name("食甚", "Greatest"), Value: info.GreatestEclipse.In(options.Location).Format("15:04:05")},
		{Label: name("纬度 Lat.", "Latitude"), Value: solarEclipseFormatCoordinates(info.GreatestLongitude, info.GreatestLatitude)},
		{Label: name("中心食带宽", "Central path width"), Value: pathWidth},
		{Label: name("中心食时长", "Central duration"), Value: centralDuration},
	}
	// 地球范围的中心食始/终只在全球阶段块出现，数据块缺了它会丢掉这两个时刻。
	if !info.CentralBeginOnEarth.IsZero() {
		circumstances = append(circumstances, svgchart.PanelRow{
			Label: name("中心食始", "Central begins"),
			Value: info.CentralBeginOnEarth.In(options.Location).Format("15:04:05"),
		})
	}
	if !info.CentralEndOnEarth.IsZero() {
		circumstances = append(circumstances, svgchart.PanelRow{
			Label: name("中心食终", "Central ends"),
			Value: info.CentralEndOnEarth.In(options.Location).Format("15:04:05"),
		})
	}
	if hasLocal {
		circumstances[1] = svgchart.PanelRow{
			Label: name("站心食分", "Local magnitude"),
			Value: fmt.Sprintf("%.4f", local.Magnitude),
		}
	}
	return penumbra, umbra, circumstances
}

func solarEclipseEphemerisRows(
	panel eclipsecore.SolarEclipseGeocentricPanel,
	options SolarEclipseMapSVGOptions,
) (ephemeris, libration []svgchart.PanelRow) {
	ephemeris = []svgchart.PanelRow{
		{Label: "历表", Value: panel.Ephemeris},
		{Label: "ΔT", Value: fmt.Sprintf("%.1f s", panel.DeltaTSeconds)},
		{Label: "k1", Value: fmt.Sprintf("%.7f", panel.PenumbralK)},
		{Label: "k2", Value: fmt.Sprintf("%.7f", panel.UmbralK)},
		{Label: "Δb", Value: fmt.Sprintf("%+.1f\"", panel.BodyShiftLatitudeArcsec)},
		{Label: "Δl", Value: fmt.Sprintf("%+.1f\"", panel.BodyShiftLongitudeArcsec)},
	}
	libration = []svgchart.PanelRow{
		{Label: "经天平动 l", Value: fmt.Sprintf("%+.2f°", panel.LibrationLongitudeDeg)},
		{Label: "纬天平动 b", Value: fmt.Sprintf("%+.2f°", panel.LibrationLatitudeDeg)},
		{Label: "自转轴位置角 c", Value: fmt.Sprintf("%.2f°", panel.LibrationPositionAngleDeg)},
		{Label: "布朗月序数", Value: fmt.Sprintf("%d", panel.BrownLunationNumber)},
	}
	if options.Language == "en" {
		ephemeris[0].Label = "Ephemeris"
		libration[0].Label, libration[1].Label = "Libration l", "Libration b"
		libration[2].Label, libration[3].Label = "Axis position angle c", "Brown lunation"
	}
	return ephemeris, libration
}

func solarEclipseGeocentricRows(
	panel eclipsecore.SolarEclipseGeocentricPanel,
) (sun, moon []svgchart.PanelRow) {
	sun = []svgchart.PanelRow{
		{Label: "赤经 R.A.", Value: formatSolarEclipseRA(panel.SunRightAscensionDeg)},
		{Label: "赤纬 Dec.", Value: formatSolarEclipseDec(panel.SunDeclinationDeg)},
		{Label: "视半径 S.D.", Value: formatSolarEclipseArcsec(panel.SunSemidiameterArcsec)},
		{Label: "地平视差 H.P.", Value: formatSolarEclipseArcsec(panel.SunParallaxArcsec)},
	}
	moon = []svgchart.PanelRow{
		{Label: "赤经 R.A.", Value: formatSolarEclipseRA(panel.MoonRightAscensionDeg)},
		{Label: "赤纬 Dec.", Value: formatSolarEclipseDec(panel.MoonDeclinationDeg)},
		{Label: "视半径 S.D.", Value: formatSolarEclipseArcsec(panel.MoonSemidiameterArcsec)},
		{Label: "地平视差 H.P.", Value: formatSolarEclipseArcsec(panel.MoonParallaxArcsec)},
	}
	return sun, moon
}
