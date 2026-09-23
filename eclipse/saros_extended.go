package eclipse

import (
	"math"
	"sync"

	"github.com/starainrt/astro/basic"
)

const (
	sarosInexLunations = 358
	sarosLunationEpoch = 2451550.09765
	// Keep the acceptance interval below half a synodic month while allowing
	// the mean-phase seed to drift in remote eras.
	sarosEclipseSeedToleranceDays = 7.0
	// 表外外推按 ±90 个沙罗周期求值同一批朔望月，直映缓存让相邻序列成员复用结果。
	sarosEclipseMemoSlots = 4096
)

// 键是（朔望月序号，相位）；判定与模型无关，只有月序号和相位决定结果。
type sarosEclipseMemoEntry struct {
	key     int64
	maximum float64
	exists  bool
	valid   bool
	set     bool
}

var (
	sarosEclipseMemoMu      sync.RWMutex
	sarosEclipseMemoEntries [sarosEclipseMemoSlots]sarosEclipseMemoEntry
)

// 直接取模即可：外推窗口的步长 223 与槽数互质，±90 个回次落在互不相同的槽位上。
func sarosEclipseMemoSlot(key int64) int {
	return int(uint64(key) % sarosEclipseMemoSlots)
}

func sarosEclipseLookup(k, phase int) (float64, bool, bool, bool) {
	key := int64(k)*2 + int64(phase)
	slot := sarosEclipseMemoSlot(key)
	sarosEclipseMemoMu.RLock()
	entry := sarosEclipseMemoEntries[slot]
	sarosEclipseMemoMu.RUnlock()
	if !entry.set || entry.key != key {
		return 0, false, false, false
	}
	return entry.maximum, entry.exists, entry.valid, true
}

func sarosEclipseStore(k, phase int, maximum float64, exists, valid bool) {
	key := int64(k)*2 + int64(phase)
	slot := sarosEclipseMemoSlot(key)
	sarosEclipseMemoMu.Lock()
	sarosEclipseMemoEntries[slot] = sarosEclipseMemoEntry{
		key: key, maximum: maximum, exists: exists, valid: valid, set: true,
	}
	sarosEclipseMemoMu.Unlock()
}

// A span contains consecutive 223-month returns. Separate spans preserve gaps
// in shallow series without counting the missing returns as eclipses.
type sarosSpan struct {
	Series int
	First  int
	Last   int
	Member int
	Count  int
}

func sarosLunation(ttJDE float64, phase int) (int, bool) {
	k := (ttJDE-sarosLunationEpoch)/solarEclipseSynodicMonthDays - float64(phase)/2
	if math.IsNaN(k) || math.IsInf(k, 0) || math.Abs(k) > 1e7 {
		return 0, false
	}
	return int(math.Round(k)), true
}

// 358*38 = 1 (mod 223). Select the Inex column whose Saros row is nearest
// the catalog reference; neighboring solutions are 358 Saros rows apart.
// See NASA SEperiodicity.html, sections 1.7 and 1.9 (also valid for lunar series).
func sarosNumber(k, phase int) int {
	anchors, base := solarSarosAnchors[:], 0
	if phase == 1 {
		anchors, base = lunarSarosAnchors[:], 1
	}
	anchor := decodeSarosMagic(anchors[len(anchors)/2], base+len(anchors)/2)
	refTT := basic.JDCalc(int(anchor.Year), int(anchor.Month), float64(anchor.Day))
	refK, _ := sarosLunation(refTT, phase)
	delta := k - refK
	column := (38 * delta) % sarosCycleLunations
	column += sarosCycleLunations * int(math.Round((float64(delta)/sarosInexLunations-float64(column))/sarosCycleLunations))
	return int(anchor.Series) + column
}

func matchSarosSpans(spans []sarosSpan, k int) (SarosInfo, bool) {
	for _, span := range spans {
		if k < span.First || k > span.Last || (k-span.First)%sarosCycleLunations != 0 {
			continue
		}
		return SarosInfo{
			Series: span.Series,
			Member: span.Member + (k-span.First)/sarosCycleLunations,
			Count:  span.Count,
		}, true
	}
	return SarosInfo{}, false
}

func extendedSarosInfo(ttJDE float64, phase int) (SarosInfo, bool) {
	k, ok := sarosLunation(ttJDE, phase)
	if !ok {
		return SarosInfo{}, false
	}
	if ttJDE >= sarosExtendedStartTT && ttJDE < sarosExtendedEndTT {
		spans := solarSarosExtended[:]
		if phase == 1 {
			spans = lunarSarosExtended[:]
		}
		info, ok := matchSarosSpans(spans, k)
		if !ok {
			return SarosInfo{}, false
		}
		if anchor, known := sarosAnchorRangeFor(phase, info.Series); known {
			info.Count = anchor.count
		}
		return info, true
	}
	return extrapolateSaros(k, phase)
}

func sarosEclipse(k, phase int) (float64, bool, bool) {
	if maximum, exists, valid, ok := sarosEclipseLookup(k, phase); ok {
		return maximum, exists, valid
	}
	maximum, exists, valid := sarosEclipseUncached(k, phase)
	sarosEclipseStore(k, phase, maximum, exists, valid)
	return maximum, exists, valid
}

// Saros metadata uses the Split-K solar model and the union of Danjon and
// Chauvenet lunar detections, independent of the observer or display model.
func sarosEclipseUncached(k, phase int) (float64, bool, bool) {
	seed := sarosLunationEpoch + (float64(k)+float64(phase)/2)*solarEclipseSynodicMonthDays
	var maximum float64
	var exists bool
	if phase == 0 {
		result := basic.SolarEclipseNASABulletinSplitK(seed)
		maximum, exists = result.GreatestEclipse, result.Type != basic.SolarEclipseNone
	} else {
		result := basic.LunarEclipseDanjon(seed)
		if result.Type == basic.LunarEclipseNone {
			result = basic.LunarEclipseChauvenet(seed)
		}
		maximum, exists = result.Maximum, result.Type != basic.LunarEclipseNone
	}
	valid := !math.IsNaN(maximum) && !math.IsInf(maximum, 0) && math.Abs(maximum-seed) < sarosEclipseSeedToleranceDays
	return maximum, exists, valid
}

func extrapolateSaros(k, phase int) (SarosInfo, bool) {
	info := SarosInfo{Series: sarosNumber(k, phase)}
	for offset := -sarosExtrapolationWindow; offset <= sarosExtrapolationWindow; offset++ {
		_, exists, valid := sarosEclipse(k+offset*sarosCycleLunations, phase)
		if !valid || (offset == 0 && !exists) {
			return SarosInfo{}, false
		}
		if !exists {
			continue
		}
		if offset == -sarosExtrapolationWindow || offset == sarosExtrapolationWindow {
			return SarosInfo{}, false
		}
		info.Count++
		if offset <= 0 {
			info.Member++
		}
	}
	return info, true
}
