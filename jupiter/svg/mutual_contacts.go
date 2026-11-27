package svg

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/jupiter"
)

type mutualContact struct {
	code     string
	name     string
	date     time.Time
	geometry basic.JupiterGalileanMutualDiagram
	greatest bool
}

type contactRadius func(basic.JupiterGalileanMutualDiagram) float64

func centreSeparation(geometry basic.JupiterGalileanMutualDiagram) float64 {
	return math.Hypot(geometry.XJupiterR, geometry.YJupiterR)
}

// mutualContacts 互掩给出外切 C1/C4 与内切 C2/C3，互食给出半影 P1/P4 与本影 U1–U4；内切只在事件确实到达时出现。
func mutualContacts(drawing *chart, event jupiter.GalileanMutualEvent, track mutualTrack) []mutualContact {
	eclipse := event.Type == jupiter.GalileanMutualEclipse
	greatest := track.contacts[1]
	var outer, inner contactRadius
	if eclipse {
		outer = func(g basic.JupiterGalileanMutualDiagram) float64 {
			if g.UmbraRadiusJupiterR <= 0 {
				return -1
			}
			return g.UmbraRadiusJupiterR + g.PassiveRadiusJupiterR
		}
		inner = func(g basic.JupiterGalileanMutualDiagram) float64 {
			if g.UmbraRadiusJupiterR <= 0 {
				return -1
			}
			return math.Abs(g.UmbraRadiusJupiterR - g.PassiveRadiusJupiterR)
		}
	} else {
		inner = func(g basic.JupiterGalileanMutualDiagram) float64 {
			return math.Abs(g.ActiveRadiusJupiterR - g.PassiveRadiusJupiterR)
		}
	}
	names := contactNames(drawing, event, greatest.geometry)
	result := []mutualContact{{names[0][0], names[0][1], track.contacts[0].date, track.contacts[0].geometry, false}}
	var ingress, egress []mutualContact
	for level, radius := range []contactRadius{outer, inner} {
		if radius == nil {
			continue
		}
		in, out, ok := contactPair(event, radius)
		if !ok {
			continue
		}
		index := 1 + level
		if !eclipse {
			index = 2
		}
		ingress = append(ingress, mutualContact{names[index][0], names[index][1], in.date, in.geometry, false})
		egress = append([]mutualContact{{names[len(names)-1-index][0], names[len(names)-1-index][1], out.date, out.geometry, false}}, egress...)
	}
	result = append(result, ingress...)
	result = append(result, mutualContact{names[len(names)/2][0], names[len(names)/2][1], greatest.date, greatest.geometry, true})
	result = append(result, egress...)
	last := names[len(names)-1]
	return append(result, mutualContact{last[0], last[1], track.contacts[2].date, track.contacts[2].geometry, false})
}

// contactNames 依次为外切始、本影外切始、内切始、甚、内切终、本影外切终、外切终（互掩无本影两档）。
func contactNames(drawing *chart, event jupiter.GalileanMutualEvent, greatest basic.JupiterGalileanMutualDiagram) [][2]string {
	if event.Type == jupiter.GalileanMutualEclipse {
		begin, end := drawing.words("食既", "total in"), drawing.words("生光", "total out")
		if greatest.UmbraRadiusJupiterR < greatest.PassiveRadiusJupiterR {
			begin, end = drawing.words("环食始", "annular in"), drawing.words("环食终", "annular out")
		}
		return [][2]string{
			{"P1", drawing.words("半影始", "penumbra in")},
			{"U1", drawing.words("初亏", "umbra in")},
			{"U2", begin},
			{"", drawing.words("食甚", "greatest")},
			{"U3", end},
			{"U4", drawing.words("复圆", "umbra out")},
			{"P4", drawing.words("半影终", "penumbra out")},
		}
	}
	begin, end := drawing.words("全掩始", "total in"), drawing.words("全掩终", "total out")
	if greatest.ActiveRadiusJupiterR < greatest.PassiveRadiusJupiterR {
		begin, end = drawing.words("环掩始", "annular in"), drawing.words("环掩终", "annular out")
	}
	return [][2]string{
		{"C1", drawing.words("初亏", "first contact")},
		{"", ""},
		{"C2", begin},
		{"", drawing.words("掩甚", "greatest")},
		{"C3", end},
		{"", ""},
		{"C4", drawing.words("复圆", "last contact")},
	}
}

type contactSample struct {
	date     time.Time
	geometry basic.JupiterGalileanMutualDiagram
}

// contactPair 在 [Start, Greatest] 找首个、在 [Greatest, End] 找末个 圆心距 = radius 的穿越，1 min 取样后二分到亚毫秒。
func contactPair(event jupiter.GalileanMutualEvent, radius contactRadius) (contactSample, contactSample, bool) {
	gap := func(date time.Time) (float64, basic.JupiterGalileanMutualDiagram, bool) {
		geometry, ok := mutualDiagramAt(date, event)
		if !ok {
			return 0, geometry, false
		}
		limit := radius(geometry)
		if limit < 0 {
			return 1, geometry, true
		}
		return centreSeparation(geometry) - limit, geometry, true
	}
	if value, _, ok := gap(event.Greatest); !ok || value >= 0 {
		return contactSample{}, contactSample{}, false
	}
	find := func(from, to time.Time) (contactSample, bool) {
		step := time.Minute
		if to.Before(from) {
			step = -step
		}
		previous := from
		previousValue, _, ok := gap(previous)
		if !ok {
			return contactSample{}, false
		}
		for current := from.Add(step); ; current = current.Add(step) {
			if (step > 0 && current.After(to)) || (step < 0 && current.Before(to)) {
				current = to
			}
			value, _, valid := gap(current)
			if !valid {
				return contactSample{}, false
			}
			if (previousValue >= 0) != (value >= 0) {
				outside, inside := previous, current
				if previousValue < 0 {
					outside, inside = current, previous
				}
				for iteration := 0; iteration < 48; iteration++ {
					middle := outside.Add(inside.Sub(outside) / 2)
					middleValue, _, ok := gap(middle)
					if !ok {
						return contactSample{}, false
					}
					if middleValue >= 0 {
						outside = middle
					} else {
						inside = middle
					}
				}
				_, geometry, ok := gap(inside)
				return contactSample{inside, geometry}, ok
			}
			if current.Equal(to) {
				return contactSample{}, false
			}
			previous, previousValue = current, value
		}
	}
	in, inOK := find(event.Start, event.Greatest)
	out, outOK := find(event.End, event.Greatest)
	return in, out, inOK && outOK
}
