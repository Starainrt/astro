package jupiter

import (
	"fmt"
	"sort"
	"time"

	"github.com/starainrt/astro/basic"
)

// GalileanEventFamily 伽利略卫星事件族 / Galilean satellite event family.
type GalileanEventFamily string

const (
	// GalileanFamilyPhenomenon 单卫星整场现象 / whole single-satellite phenomenon.
	GalileanFamilyPhenomenon GalileanEventFamily = "phenomenon"
	// GalileanFamilyContact 单卫星有限圆盘接触事件 / finite-disk single-satellite contact event.
	GalileanFamilyContact GalileanEventFamily = "contact"
	// GalileanFamilyMutual 卫星间互现象 / mutual satellite event.
	GalileanFamilyMutual GalileanEventFamily = "mutual"
)

// GalileanEventsOptions 批量事件选项，卫星全 false 表示四颗全选，互现象要求双方均选中 / batch options; all-false selects all satellites, and mutual pairs require both selected.
type GalileanEventsOptions struct {
	Satellites     [4]bool
	IncludeContact bool
	IncludeMutual  bool
}

// GalileanEventEntry 批量事件记录，仅 Family 对应的载荷有效 / batch event entry with only the Family payload populated.
type GalileanEventEntry struct {
	Family   GalileanEventFamily
	Start    time.Time
	Greatest time.Time
	End      time.Time

	Phenomenon GalileanPhenomenonEvent
	Contact    GalileanPhenomenonContactEvent
	Mutual     GalileanMutualEvent
}

// GalileanEventsBetween 从 start 逐次 Next 汇总极大时刻早于 end 的事件，跨度最多 366 天 / collect successive Next events from start with Greatest before end, spanning at most 366 days.
func (module *Jupiter) GalileanEventsBetween(start, end time.Time, options GalileanEventsOptions) ([]GalileanEventEntry, error) {
	if end.Before(start) || end.Sub(start) > 366*24*time.Hour {
		return nil, fmt.Errorf("Galilean event interval must be between 0 and 366 days")
	}
	if start.Equal(end) {
		return nil, nil
	}
	satellites := options.Satellites
	if satellites == [4]bool{} {
		satellites = [4]bool{true, true, true, true}
	}
	var entries []GalileanEventEntry
	for index, selected := range satellites {
		if !selected {
			continue
		}
		for _, phenomenon := range []GalileanPhenomenonType{GalileanPhenomenonTransit, GalileanPhenomenonOccultation, GalileanPhenomenonEclipse, GalileanPhenomenonShadowTransit} {
			for seed := start; seed.Before(end); {
				event := module.NextGalileanPhenomenonEvent(seed, index+1, phenomenon)
				if !event.Valid || !event.Greatest.Before(end) {
					break
				}
				entries = append(entries, GalileanEventEntry{
					Family: GalileanFamilyPhenomenon, Start: event.Start, Greatest: event.Greatest, End: event.End, Phenomenon: event,
				})
				seed = event.End.Add(100 * time.Millisecond)
			}
			if !options.IncludeContact {
				continue
			}
			for seed := start; seed.Before(end); {
				event := module.NextGalileanPhenomenonContactEvent(seed, index+1, phenomenon)
				if !event.Valid {
					whole := module.NextGalileanPhenomenonEvent(seed, index+1, phenomenon)
					if !whole.Valid || !whole.Greatest.Before(end) {
						break
					}
					nextSeed := whole.End.Add(100 * time.Millisecond)
					if !nextSeed.After(seed) {
						break
					}
					seed = nextSeed
					continue
				}
				if !event.Greatest.Before(end) {
					break
				}
				entries = append(entries, GalileanEventEntry{
					Family: GalileanFamilyContact, Start: event.Disappearance.Start, Greatest: event.Greatest, End: event.Reappearance.End, Contact: event,
				})
				seed = event.Reappearance.End.Add(100 * time.Millisecond)
			}
		}
	}
	if options.IncludeMutual {
		endTT := basic.UTC2TT(basic.Date2JD(end.UTC()))
		for first := 1; first <= 4; first++ {
			for second := first + 1; second <= 4; second++ {
				if !satellites[first-1] || !satellites[second-1] {
					continue
				}
				for _, phenomenon := range []basic.JupiterGalileanMutualPhenomenonType{basic.JupiterGalileanMutualOccultation, basic.JupiterGalileanMutualEclipse} {
					for seed := start; seed.Before(end); {
						state := basic.NextJupiterGalileanMutualEventBeforeWithInstance(basic.UTC2TT(basic.Date2JD(seed.UTC())), endTT, first, second, phenomenon, module.inst)
						event := galileanMutualEventFromBasic(state, start.Location())
						if !event.Valid || !event.Greatest.Before(end) {
							break
						}
						entries = append(entries, GalileanEventEntry{
							Family: GalileanFamilyMutual, Start: event.Start, Greatest: event.Greatest, End: event.End, Mutual: event,
						})
						seed = event.End.Add(100 * time.Millisecond)
					}
				}
			}
		}
	}
	sort.SliceStable(entries, func(first, second int) bool {
		if entries[first].Start.Equal(entries[second].Start) {
			return entries[first].Greatest.Before(entries[second].Greatest)
		}
		return entries[first].Start.Before(entries[second].Start)
	})
	return entries, nil
}
