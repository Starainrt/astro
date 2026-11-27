package svg

import (
	"fmt"
	"math"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/jupiter"
)

const (
	finderLeft   = 594.0
	finderRight  = 870.0
	finderTop    = 392.0
	finderBottom = 534.0
)

type finderGeometry struct {
	snapshot     basic.JupiterGalileanDiagram
	limbArcsec   float64
	pairArcsec   float64
	phaseDegrees float64
}

// mutualFinder 极大时刻地球视点下被遮星到木星临边、两星之间的角距，以及木星相位角。
func mutualFinder(event jupiter.GalileanMutualEvent) (finderGeometry, bool) {
	jde := basic.UTC2TT(basic.Date2JD(event.Greatest.UTC()))
	snapshot, ok := basic.JupiterGalileanDiagramWithInstance(jde, nil)
	if !ok {
		return finderGeometry{}, false
	}
	passive := snapshot.Observations[event.Passive-1]
	active := snapshot.Observations[event.Active-1]
	semidiameter := passive.OffsetXArcsec / passive.OffsetXJupiterRadii
	if math.IsNaN(semidiameter) || math.IsInf(semidiameter, 0) || semidiameter <= 0 {
		semidiameter = passive.OffsetYArcsec / passive.OffsetYJupiterRadii
	}
	if math.IsNaN(semidiameter) || math.IsInf(semidiameter, 0) || semidiameter <= 0 {
		for _, obs := range snapshot.Observations {
			if candidate := obs.OffsetXArcsec / obs.OffsetXJupiterRadii; !math.IsNaN(candidate) && !math.IsInf(candidate, 0) && candidate > 0 {
				semidiameter = candidate
				break
			}
			if candidate := obs.OffsetYArcsec / obs.OffsetYJupiterRadii; !math.IsNaN(candidate) && !math.IsInf(candidate, 0) && candidate > 0 {
				semidiameter = candidate
				break
			}
		}
	}
	if math.IsNaN(semidiameter) || math.IsInf(semidiameter, 0) || semidiameter <= 0 {
		semidiameter = jupiter.Semidiameter(event.Greatest)
	}
	limb := ellipseLimbGap(passive.OffsetXJupiterRadii, passive.OffsetYJupiterRadii, snapshot.PolePositionAngle, snapshot.MinorRadiusRatio)
	return finderGeometry{
		snapshot:     snapshot,
		limbArcsec:   (limb - snapshot.AngularRadiiJupiterR[event.Passive-1]) * semidiameter,
		pairArcsec:   math.Hypot(passive.OffsetXArcsec-active.OffsetXArcsec, passive.OffsetYArcsec-active.OffsetYArcsec),
		phaseDegrees: jupiter.PhaseAngle(event.Greatest),
	}, true
}

// ellipseLimbGap 点到投影木星椭圆临边的近似径向距离（木星赤道半径），正值在盘外。
func ellipseLimbGap(x, y, poleAngle, axisRatio float64) float64 {
	angle := poleAngle * math.Pi / 180
	along := x*math.Cos(angle) - y*math.Sin(angle)
	across := x*math.Sin(angle) + y*math.Cos(angle)
	distance := math.Hypot(along, across)
	if distance == 0 {
		return -axisRatio
	}
	theta := math.Atan2(across, along)
	limb := axisRatio / math.Hypot(axisRatio*math.Cos(theta), math.Sin(theta))
	return distance - limb
}

// drawMutualFinder 右下角地球视点寻星图：木星与四星按极大时刻位置，参与互现象的两星高亮。
func drawMutualFinder(drawing *chart, event jupiter.GalileanMutualEvent, finder finderGeometry) {
	fmt.Fprintf(drawing, `<rect id="finder-frame" x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#fafaf8" stroke="#d6d6d2" stroke-width="0.8"/>`, finderLeft, finderTop, finderRight-finderLeft, finderBottom-finderTop)
	drawing.text(finderLeft+8, finderTop+15, 11, "start", "#555", "700", drawing.words("地球视点（极大时）", "Earth view at greatest"))
	cx, cy := finderLeft+92, finderTop+62
	extent := 2.0
	for _, observation := range finder.snapshot.Observations {
		extent = math.Max(extent, math.Abs(observation.OffsetXJupiterRadii)+0.6)
	}
	scale := math.Min(80/extent, 40/(maxAbsY(finder.snapshot)+1))
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#c9c9c6" stroke-width="0.6" stroke-dasharray="2 3"/>`, finderLeft+6, cy, finderLeft+178, cy)
	drawing.text(finderLeft+8, cy-4, 10, "start", "#777", "", drawing.words("东", "E"))
	drawing.text(finderLeft+178, cy-4, 10, "end", "#777", "", drawing.words("西", "W"))
	paintOrder := galileanPaintOrder(finder.snapshot)
	for _, foreground := range []bool{false, true} {
		if foreground {
			jupiterBody(drawing, finder.snapshot, cx, cy, scale, "finder-jupiter", "finder-limb")
		}
		for _, index := range paintOrder {
			observation := finder.snapshot.Observations[index]
			if observation.InFrontOfJupiter != foreground {
				continue
			}
			x, y := cx-observation.OffsetXJupiterRadii*scale, cy-observation.OffsetYJupiterRadii*scale
			fill, stroke, radius := "#d9d6cf", "#a9a49a", 2.6
			if index+1 == event.Passive {
				fill, stroke, radius = dayFill, passiveStroke, 3.8
			} else if index+1 == event.Active {
				fill, stroke, radius = activeFill, activeStroke, 3.8
			}
			fmt.Fprintf(drawing, `<circle id="finder-satellite-%d" cx="%.3f" cy="%.3f" r="%.2f" fill="%s" stroke="%s" stroke-width="1.1"/>`, index+1, x, y, radius, fill, stroke)
		}
	}
	drawFinderPair(drawing, event, finder)
	passiveName := satelliteName(event.Passive-1, drawing.options.english)
	activeName := satelliteName(event.Active-1, drawing.options.english)
	drawing.text(finderLeft+8, finderBottom-26, 11, "start", passiveStroke, "", fmt.Sprintf(drawing.words("● %s 距木星边缘 %.1f″", "● %s to Jupiter limb %.1f″"), passiveName, finder.limbArcsec))
	drawing.text(finderLeft+8, finderBottom-10, 11, "start", activeStroke, "", fmt.Sprintf(drawing.words("● %s 与%s相距 %.2f″", "● %s to %s %.2f″"), activeName, passiveName, finder.pairArcsec))
}

func maxAbsY(snapshot basic.JupiterGalileanDiagram) float64 {
	value := 0.0
	for _, observation := range snapshot.Observations {
		value = math.Max(value, math.Abs(observation.OffsetYJupiterRadii))
	}
	return value
}

// drawFinderPair 右上角放大两星的相对位置（东左北上），比例标出 1″。
func drawFinderPair(drawing *chart, event jupiter.GalileanMutualEvent, finder finderGeometry) {
	passive := finder.snapshot.Observations[event.Passive-1]
	active := finder.snapshot.Observations[event.Active-1]
	const cx, cy, half = finderRight - 46.0, finderTop + 56.0, 30.0
	fmt.Fprintf(drawing, `<rect id="finder-pair" x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#ffffff" stroke="#d6d6d2" stroke-width="0.6"/>`, cx-half-6, cy-half-6, 2*half+12, 2*half+12)
	span := math.Max(2, 1.3*finder.pairArcsec)
	scale := half / span
	dx, dy := active.OffsetXArcsec-passive.OffsetXArcsec, active.OffsetYArcsec-passive.OffsetYArcsec
	fmt.Fprintf(drawing, `<circle id="finder-pair-passive" cx="%.3f" cy="%.3f" r="3.6" fill="%s" stroke="%s" stroke-width="1.4"/>`, cx, cy, dayFill, passiveStroke)
	fmt.Fprintf(drawing, `<circle id="finder-pair-active" cx="%.3f" cy="%.3f" r="3.6" fill="%s" stroke="%s" stroke-width="1.4"/>`, cx-dx*scale, cy-dy*scale, activeFill, activeStroke)
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#555" stroke-width="1"/>`, cx-half, cy+half+2, cx-half+scale, cy+half+2)
	drawing.text(cx-half+scale+3, cy+half+5, 9, "start", "#555", "", "1″")
}
