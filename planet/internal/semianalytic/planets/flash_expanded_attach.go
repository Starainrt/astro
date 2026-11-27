//go:build astro_lowmem
// +build astro_lowmem

package planets

type expandedSpan struct {
	argsOffset uint32
	argsCount  uint32
	coefOffset uint32
	coefCount  uint32
}

type expandedSource uint8

const (
	expandedLong expandedSource = iota
	expandedModern
	expandedFast
)

func attachExpanded(bodies [9]*flashBody, table string, source expandedSource) {
	for i := range bodies {
		body := bodies[i]
		if body == nil {
			continue
		}
		for j := range body.fel {
			var span expandedSpan
			switch source {
			case expandedLong:
				span = flashExpandedLongSpans[i][j]
			case expandedModern:
				span = flashExpandedModernSpans[i][j]
			case expandedFast:
				span = flashExpandedFastSpans[i][j]
			}
			if span.coefCount == 0 {
				continue
			}
			argsStart := int(span.argsOffset)
			argsEnd := argsStart + int(span.argsCount)*2
			coefStart := int(span.coefOffset)
			coefEnd := coefStart + int(span.coefCount)*8
			body.fel[j].expanded = expandedData{args: table[argsStart:argsEnd], coef: table[coefStart:coefEnd]}
		}
	}
}
