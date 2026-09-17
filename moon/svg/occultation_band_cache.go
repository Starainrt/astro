package svg

import (
	"math"
	"sync"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon"
)

// occultationBandPolygons 是一次紧凑掩带合并的只读结果。
type occultationBandPolygons struct {
	polygons      [][]geodata.GeoPoint
	authoritative bool
}

type occultationBandCacheKey struct {
	kind   occultationFootprintSweepKind
	digest [2]uint64
}

// occultationBandCacheCapacity 限制常驻的合并结果数量；掩带多边形可能上万点。
const occultationBandCacheCapacity = 6

var occultationBandCache = struct {
	sync.RWMutex
	entries map[occultationBandCacheKey]occultationBandPolygons
	order   []occultationBandCacheKey
}{entries: make(map[occultationBandCacheKey]occultationBandPolygons, occultationBandCacheCapacity)}

func lookupOccultationBandPolygons(key occultationBandCacheKey) (occultationBandPolygons, bool) {
	occultationBandCache.RLock()
	value, ok := occultationBandCache.entries[key]
	occultationBandCache.RUnlock()
	return value, ok
}

func storeOccultationBandPolygons(key occultationBandCacheKey, value occultationBandPolygons) {
	occultationBandCache.Lock()
	defer occultationBandCache.Unlock()
	if _, exists := occultationBandCache.entries[key]; exists {
		return
	}
	if len(occultationBandCache.order) >= occultationBandCacheCapacity {
		delete(occultationBandCache.entries, occultationBandCache.order[0])
		occultationBandCache.order = occultationBandCache.order[1:]
	}
	occultationBandCache.entries[key] = value
	occultationBandCache.order = append(occultationBandCache.order, key)
}

// occultationBandDigest 汇总合并的全部输入；按 64 位字混合而不是逐字节，摘要成本必须远低于合并本身。
func occultationBandDigest(s occultationFootprintSweep) [2]uint64 {
	const (
		seed1 = 0x9e3779b97f4a7c15
		seed2 = 0xc2b2ae3d27d4eb4f
		prime = 0x100000001b3
	)
	first, second := uint64(seed1), uint64(seed2)
	write := func(value uint64) {
		first = (first ^ value) * prime
		first ^= first >> 31
		second = (second + value) * prime
		second ^= second >> 33
	}
	writeFloat := func(value float64) { write(math.Float64bits(value)) }
	writeTime := func(value int64) { write(uint64(value)) }
	writePoint := func(point moon.OccultationPathPoint) {
		writeTime(point.Time.UnixNano())
		writeFloat(point.Longitude)
		writeFloat(point.Latitude)
		writeFloat(point.MoonAltitude)
		writeFloat(point.WidthKM)
		writeFloat(point.LimitSeparationKM)
	}
	writePolygons := func(polygons [][]moon.OccultationPathPoint) {
		write(uint64(len(polygons)))
		for _, polygon := range polygons {
			write(uint64(len(polygon)))
			for _, point := range polygon {
				writePoint(point)
			}
		}
	}
	writeString := func(value string) {
		write(uint64(len(value)))
		for index := 0; index < len(value); index++ {
			write(uint64(value[index]))
		}
	}
	writeFootprints := func(footprints []moon.PlanetOccultationFootprint) {
		write(uint64(len(footprints)))
		for _, footprint := range footprints {
			writeTime(footprint.Time.UnixNano())
			writePolygons(footprint.Polygons)
			writePolygons(footprint.InteriorPolygons)
			writePolygons(footprint.Boundaries)
			if footprint.Closed {
				write(1)
			} else {
				write(0)
			}
		}
	}
	writeCurves := func(curves []moon.OccultationRiseSetCurve) {
		write(uint64(len(curves)))
		for _, curve := range curves {
			writeString(string(curve.Phase))
			writeString(string(curve.Direction))
			writePolygons(curve.Segments)
		}
	}
	write(uint64(s.kind))
	writeFootprints(s.footprints)
	writePolygons(s.contours)
	writePolygons(s.visibilityContours)
	writePolygons([][]moon.OccultationPathPoint{s.northern, s.southern})
	writeCurves(s.curves)
	return [2]uint64{first, second}
}
