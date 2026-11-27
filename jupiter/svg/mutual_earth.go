package svg

import (
	"fmt"
	"math"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/jupiter"
)

const (
	phaseStripTop    = 446.0
	phaseStripBottom = 532.0
	phaseStripLeft   = 40.0
	phaseStripRight  = 578.0
)

type earthShadowSample struct {
	contact mutualContact
	shadow  basic.JupiterGalileanEarthShadow
}

// earthShadowSamples 各接触点与食甚处地球所见被食星盘面；P1、P4 恰在半影外切，盘面无影。
func earthShadowSamples(event jupiter.GalileanMutualEvent, contacts []mutualContact) ([]earthShadowSample, bool) {
	var samples []earthShadowSample
	for _, contact := range contacts {
		shadow, ok := basic.JupiterGalileanEarthShadowWithInstance(basic.UTC2TT(basic.Date2JD(contact.date.UTC())), event.Active, event.Passive, nil)
		if !ok {
			return nil, false
		}
		samples = append(samples, earthShadowSample{contact, shadow})
	}
	return samples, true
}

func greatestEarthShadow(samples []earthShadowSample) basic.JupiterGalileanEarthShadow {
	for _, sample := range samples {
		if sample.contact.greatest {
			return sample.shadow
		}
	}
	return basic.JupiterGalileanEarthShadow{}
}

// drawEarthPhaseStrip 底部一排地球视点小盘：各接触点被食星的亮面、夜面、半影、本影，左东右西、上北下南。
func drawEarthPhaseStrip(drawing *chart, event jupiter.GalileanMutualEvent, samples []earthShadowSample, visibility []contactVisibility) {
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#d6d6d2" stroke-width="0.8"/>`, phaseStripLeft, phaseStripTop-8, phaseStripRight, phaseStripTop-8)
	drawing.text(phaseStripLeft, phaseStripTop+6, 11, "start", "#555", "700", drawing.words("地球所见被食星（东左北上）", "Eclipsed satellite from Earth (E left, N up)"))
	count := float64(len(samples))
	slot := (phaseStripRight - phaseStripLeft) / count
	radius := math.Min(19, slot/2-8)
	for index, sample := range samples {
		cx := phaseStripLeft + slot*(float64(index)+0.5)
		cy := phaseStripTop + 34
		drawEarthDisk(drawing, fmt.Sprintf("earth-disk-%s", contactID(sample.contact, index)), cx, cy, radius, sample.shadow)
		drawEarthCoreDot(drawing, cx, cy, radius, sample.shadow)
		drawContactTick(drawing, sample, cx, cy, radius)
		color := "#2554c7"
		if visibility[index].hidden() {
			color = "#9a9a9a"
		}
		drawing.text(cx, cy+radius+13, 10, "middle", color, "", contactShort(drawing, sample.contact))
	}
}

// drawEarthDisk 用扫描段画盘面；东为屏幕左，故扫描段 X 取负。
func drawEarthDisk(drawing *chart, id string, cx, cy, radius float64, shadow basic.JupiterGalileanEarthShadow) {
	fmt.Fprintf(drawing, `<g id="%s"><circle cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="#7a705f" stroke-width="0.8"/>`, id, cx, cy, radius, dayFill)
	colors := map[int]string{basic.JupiterGalileanEarthNight: "#4a4742", basic.JupiterGalileanEarthPenumbra: "#c4b9a6", basic.JupiterGalileanEarthUmbra: "#b21f16", basic.JupiterGalileanEarthAntumbra: "#e08a2c"}
	height := shadow.RowHeight * radius
	for _, run := range shadow.Runs {
		color, ok := colors[run.State]
		if !ok {
			continue
		}
		left := cx - run.X1*radius
		fmt.Fprintf(drawing, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="%s"/>`, left, cy-run.Y*radius-height/2, (run.X1-run.X0)*radius, height+0.05, color)
	}
	fmt.Fprintf(drawing, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#7a705f" stroke-width="0.8"/></g>`, cx, cy, radius)
}

// drawEarthCoreDot 本影或伪本影小于一个扫描行时，在其所在处补一个最小可见点，避免盘面上看不出。
func drawEarthCoreDot(drawing *chart, cx, cy, radius float64, shadow basic.JupiterGalileanEarthShadow) {
	for _, state := range []int{basic.JupiterGalileanEarthUmbra, basic.JupiterGalileanEarthAntumbra} {
		var sx, sy, total float64
		for _, run := range shadow.Runs {
			if run.State != state {
				continue
			}
			width := run.X1 - run.X0
			sx += width * (run.X0 + run.X1) / 2
			sy += width * run.Y
			total += width
		}
		if total == 0 || total*shadow.RowHeight > 0.02 {
			continue
		}
		fill := "#b21f16"
		if state == basic.JupiterGalileanEarthAntumbra {
			fill = "#e08a2c"
		}
		fmt.Fprintf(drawing, `<circle class="earth-core-dot" cx="%.3f" cy="%.3f" r="1.8" fill="%s"/>`, cx-sx/total*radius, cy-sy/total*radius, fill)
	}
}

// earthCoverageLines 食甚时地球视向的本影、本影加半影覆盖比例，与图上太阳视向的几何覆盖率对照。
func earthCoverageLines(drawing *chart, shadow basic.JupiterGalileanEarthShadow) []string {
	return []string{
		fmt.Sprintf(drawing.words("地球视角本影/伪本影 %.2f%%", "Earth-view umbra/antumbra %.2f%%"), 100*(shadow.UmbraCoverage+shadow.AntumbraCoverage)),
		fmt.Sprintf(drawing.words("地球视角含半影 %.1f%%", "Earth-view any shadow %.1f%%"), 100*shadow.ShadowCoverage),
	}
}

// drawContactTick 盘缘三角指向影锥轴所在一侧（影从这里进入或离开）：P 灰、U 红；盘心在轴上时不标。
func drawContactTick(drawing *chart, sample earthShadowSample, cx, cy, radius float64) {
	code := sample.contact.code
	direction := sample.shadow.ShadowDirection
	if code == "" || (direction[0] == 0 && direction[1] == 0) {
		return
	}
	color := "#b21f16"
	if code[0] == 'P' {
		color = "#777"
	}
	// 东为屏幕左、北为屏幕上，所以屏幕 x 取 −东、屏幕 y 取 −北。
	dx, dy := -direction[0], -direction[1]
	x0, y0 := cx+dx*(radius+1.5), cy+dy*(radius+1.5)
	x1, y1 := cx+dx*(radius+7), cy+dy*(radius+7)
	fmt.Fprintf(drawing, `<path id="earth-tick-%s" d="M %.3f %.3f L %.3f %.3f L %.3f %.3f Z" fill="%s"/>`, code, x0, y0, x1-dy*3, y1+dx*3, x1+dy*3, y1-dx*3, color)
}

// drawOccultationStrip 底部一排小盘：各接触点被遮星在地球视向上被遮挡星挡住的部分（红），东左北上，比例为两星真实半径比。
func drawOccultationStrip(drawing *chart, event jupiter.GalileanMutualEvent, contacts []mutualContact, visibility []contactVisibility) {
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#d6d6d2" stroke-width="0.8"/>`, phaseStripLeft, phaseStripTop-8, phaseStripRight, phaseStripTop-8)
	drawing.text(phaseStripLeft, phaseStripTop+6, 11, "start", "#555", "700", drawing.words("地球所见被掩星（东左北上）", "Occulted satellite from Earth (E left, N up)"))
	slot := (phaseStripRight - phaseStripLeft) / float64(len(contacts))
	radius := math.Min(19, slot/2-8)
	for index, contact := range contacts {
		geometry := contact.geometry
		cx := phaseStripLeft + slot*(float64(index)+0.5)
		cy := phaseStripTop + 34
		scale := radius / geometry.PassiveRadiusJupiterR
		ax, ay := cx+geometry.XJupiterR*scale, cy+geometry.YJupiterR*scale
		id := contactID(contact, index)
		fmt.Fprintf(drawing, `<g id="occultation-disk-%s"><clipPath id="occultation-clip-%s"><circle cx="%.3f" cy="%.3f" r="%.3f"/></clipPath>`, id, id, cx, cy, radius)
		fmt.Fprintf(drawing, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="none"/>`, cx, cy, radius, dayFill)
		fmt.Fprintf(drawing, `<circle class="occultation-cover" cx="%.3f" cy="%.3f" r="%.3f" fill="#b21f16" opacity="0.75" clip-path="url(#occultation-clip-%s)"/>`, ax, ay, geometry.ActiveRadiusJupiterR*scale, id)
		fmt.Fprintf(drawing, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="%s" stroke-width="1.2"/></g>`, cx, cy, radius, passiveStroke)
		color := "#2554c7"
		if visibility[index].hidden() {
			color = "#9a9a9a"
		}
		drawing.text(cx, cy+radius+13, 10, "middle", color, "", contactShort(drawing, contact))
	}
}
