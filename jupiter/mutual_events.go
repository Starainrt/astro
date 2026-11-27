package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// GalileanMutualPhenomenonType 伽利略卫星互现象类型 / Galilean mutual phenomenon type.
type GalileanMutualPhenomenonType string

const (
	// GalileanMutualOccultation 互掩 / mutual occultation.
	GalileanMutualOccultation GalileanMutualPhenomenonType = "mutual_occultation"
	// GalileanMutualEclipse 互食 / mutual eclipse.
	GalileanMutualEclipse GalileanMutualPhenomenonType = "mutual_eclipse"
)

// GalileanMutualEvent 全部几何互现象事件，包括卫星被木星掩蔽或处于木影中的事件，可观测性用 GreatestState 判断 / all geometric mutual events, including satellites occulted by Jupiter or in its shadow; assess visibility with GreatestState.
type GalileanMutualEvent struct {
	Valid bool
	Type  GalileanMutualPhenomenonType

	Active  int
	Passive int

	Start    time.Time
	Greatest time.Time
	End      time.Time
	Duration time.Duration

	MinimumSeparationArcsec float64
	ImpactParameter         float64 // 最小圆心距除以接触半径和，互食使用半影半径 / minimum separation divided by the contact radius sum, using the penumbra for eclipses.
	Coverage                float64 // 被遮盘的几何面积比，互食只计正本影，0 到 1 / geometric area fraction of the passive disk, positive umbra only for eclipses, 0 to 1.
	Annular                 bool    // 互掩为前景盘内含，互食为正本影截面内含于被食盘 / foreground disk, or positive umbra for eclipses, contained in the passive disk.

	GreatestState [4]GalileanSatellitePhenomenon
}

// LastGalileanMutualEvent 上一次伽利略卫星互现象 / previous Galilean mutual event.
func (module *Jupiter) LastGalileanMutualEvent(date time.Time, first, second int, phenomenonType GalileanMutualPhenomenonType) GalileanMutualEvent {
	return galileanMutualEventFromBasic(
		basic.LastJupiterGalileanMutualEventWithInstance(
			basic.UTC2TT(basic.Date2JD(date.UTC())), first, second,
			basic.JupiterGalileanMutualPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

// NextGalileanMutualEvent 下一次伽利略卫星互现象 / next Galilean mutual event.
func (module *Jupiter) NextGalileanMutualEvent(date time.Time, first, second int, phenomenonType GalileanMutualPhenomenonType) GalileanMutualEvent {
	return galileanMutualEventFromBasic(
		basic.NextJupiterGalileanMutualEventWithInstance(
			basic.UTC2TT(basic.Date2JD(date.UTC())), first, second,
			basic.JupiterGalileanMutualPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

// ClosestGalileanMutualEvent 最近一次伽利略卫星互现象 / closest Galilean mutual event.
func (module *Jupiter) ClosestGalileanMutualEvent(date time.Time, first, second int, phenomenonType GalileanMutualPhenomenonType) GalileanMutualEvent {
	return galileanMutualEventFromBasic(
		basic.ClosestJupiterGalileanMutualEventWithInstance(
			basic.UTC2TT(basic.Date2JD(date.UTC())), first, second,
			basic.JupiterGalileanMutualPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

func galileanMutualEventFromBasic(event basic.JupiterGalileanMutualEvent, zone *time.Location) GalileanMutualEvent {
	if !event.Valid {
		return GalileanMutualEvent{}
	}
	start := basic.JD2DateByZone(basic.TT2UTC(event.Start), zone, false)
	end := basic.JD2DateByZone(basic.TT2UTC(event.End), zone, false)
	result := GalileanMutualEvent{
		Valid:                   true,
		Type:                    GalileanMutualPhenomenonType(event.Type),
		Active:                  event.Active,
		Passive:                 event.Passive,
		Start:                   start,
		Greatest:                basic.JD2DateByZone(basic.TT2UTC(event.Greatest), zone, false),
		End:                     end,
		Duration:                end.Sub(start),
		MinimumSeparationArcsec: event.MinimumSeparationArcsec,
		ImpactParameter:         event.ImpactParameter,
		Coverage:                event.Coverage,
		Annular:                 event.Annular,
	}
	for index, state := range event.GreatestState {
		result.GreatestState[index] = galileanPhenomenonFromBasic(state)
		if !state.ShadowTransit {
			result.GreatestState[index].ShadowOffsetXArcsec = 0
			result.GreatestState[index].ShadowOffsetYArcsec = 0
			result.GreatestState[index].ShadowOffsetXJupiterR = 0
			result.GreatestState[index].ShadowOffsetYJupiterR = 0
		}
	}
	return result
}
