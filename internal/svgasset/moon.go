// Package svgasset 为对外 SVG 包提供共享的内嵌 SVG 片段 / Package svgasset provides shared embedded SVG fragments to the public SVG packages.
package svgasset

import _ "embed"

// moonFaceSymbol 是精简的公版月面图形，针对小型接触盘面做了简化。
// moonFaceSymbol is a compact public-domain Moon face simplified for small contact disks.
//
//go:embed lunar_eclipse_moon.svg
var moonFaceSymbol string

// MoonFaceSymbol 返回共享的月面 <symbol> 定义 / MoonFaceSymbol returns the shared Moon-face <symbol> definition.
func MoonFaceSymbol() string {
	return moonFaceSymbol
}
