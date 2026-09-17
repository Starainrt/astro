package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/moon"
)

// localOccultationMoonAppearance contains the apparent lunar illumination used
// by the fixed-site occultation diagrams.  The limb angle is topocentric so the
// rendered bright side follows the same observer geometry as the occultation.
type localOccultationMoonAppearance struct {
	illumination float64
	brightLimbPA float64
}

func localOccultationMoonAppearanceAt(date time.Time, observer moon.Observer) localOccultationMoonAppearance {
	illumination := moon.Phase(date)
	if !starOccultationFinite(illumination) {
		illumination = 1
	}
	illumination = math.Max(0, math.Min(1, illumination))
	brightLimbPA := moon.TopocentricBrightLimbPositionAngle(
		date, observer.Longitude, observer.Latitude, observer.Height,
	)
	if !starOccultationFinite(brightLimbPA) {
		brightLimbPA = 0
	}
	return localOccultationMoonAppearance{
		illumination: illumination,
		brightLimbPA: brightLimbPA,
	}
}

// writeLocalOccultationMoon draws a textured lunar disk with a phase-aware
// terminator.  clipID is unique within the SVG document because each stage
// can have a different event time and therefore a different lunar phase.
func writeLocalOccultationMoon(
	b *strings.Builder,
	cx, cy, radius float64,
	class string,
	appearance localOccultationMoonAppearance,
	clipID string,
) {
	if !starOccultationFinite(radius) || radius <= 0 {
		return
	}
	illumination := math.Max(0, math.Min(1, appearance.illumination))
	brightLimbPA := appearance.brightLimbPA
	if !starOccultationFinite(brightLimbPA) {
		brightLimbPA = 0
	}
	fmt.Fprintf(b, `<g class="%s" data-illumination="%.6f" data-bright-limb-position-angle="%.3f">`,
		html.EscapeString(class), illumination, brightLimbPA)
	if illumination < 1-1e-6 {
		fmt.Fprintf(b, `<circle class="moon-phase-dark" cx="%.3f" cy="%.3f" r="%.3f" fill="#24272a"/>`, cx, cy, radius)
	}
	if illumination > 1e-6 {
		if illumination >= 1-1e-6 {
			writeLocalOccultationMoonTexture(b, cx, cy, radius)
		} else {
			path := localOccultationMoonPhasePath(radius, illumination)
			rotation := localOccultationMoonSVGRotation(brightLimbPA)
			fmt.Fprintf(b, `<clipPath id="%s"><path d="%s" transform="translate(%.3f %.3f) rotate(%.3f)"/></clipPath>`,
				html.EscapeString(clipID), path, cx, cy, rotation)
			fmt.Fprintf(b, `<g class="moon-phase-lit" clip-path="url(#%s)">`, html.EscapeString(clipID))
			writeLocalOccultationMoonTexture(b, cx, cy, radius)
			b.WriteString(`</g>`)
		}
	}
	fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#4f5d60" stroke-width="1.1" opacity="0.9"/>`, cx, cy, radius)
	b.WriteString(`</g>`)
}

func writeLocalOccultationMoonTexture(b *strings.Builder, cx, cy, radius float64) {
	fmt.Fprintf(b, `<use href="#le-moon" x="%.3f" y="%.3f" width="%.3f" height="%.3f"/>`,
		cx-radius, cy-radius, radius*2, radius*2)
}

// The chart uses north-up and east-left coordinates.  SVG rotations are
// clockwise from +X, so a celestial position angle of 0 points upward and 90
// points left after this conversion.
func localOccultationMoonSVGRotation(brightLimbPA float64) float64 {
	return -brightLimbPA - 90
}

// localOccultationMoonPhasePath returns the illuminated portion of a unit
// projected lunar disk.  The outer arc is the visible limb; the inner ellipse
// is the projected terminator.  The +X side is the bright side before rotation.
func localOccultationMoonPhasePath(radius, illumination float64) string {
	cosPhase := 2*illumination - 1
	terminatorRadius := radius * math.Abs(cosPhase)
	sweep := 1
	if cosPhase < 0 {
		sweep = 0
	}
	return fmt.Sprintf(
		"M 0 %.3f A %.3f %.3f 0 0 1 0 %.3f A %.3f %.3f 0 0 %d 0 %.3f Z",
		-radius, radius, radius, radius, terminatorRadius, radius, sweep, -radius,
	)
}
