package planet

import "fmt"

type coordSeriesView struct {
	orders [6][]float64
}

type planetView struct {
	scale  float64
	coords [3]coordSeriesView
}

var planetViews = buildPlanetViews(planetRawData)

func buildPlanetViews(rawData [][]float64) []planetView {
	views := make([]planetView, len(rawData))
	for bodyIndex, raw := range rawData {
		if len(raw) < 20 {
			panic(fmt.Sprintf("planet raw data %d too short: %d", bodyIndex, len(raw)))
		}
		view := planetView{scale: raw[0]}
		for zn := 0; zn < 3; zn++ {
			pn := zn*6 + 1
			for order := 0; order < 6; order++ {
				start := int(raw[pn+order])
				end := int(raw[pn+order+1])
				if start < 0 || end < start || end > len(raw) {
					panic(fmt.Sprintf("planet raw data %d coord %d order %d invalid cut: %d..%d (len=%d)", bodyIndex, zn, order, start, end, len(raw)))
				}
				view.coords[zn].orders[order] = raw[start:end]
			}
		}
		views[bodyIndex] = view
	}
	return views
}
