package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// GalileanPhenomenonType 伽利略卫星现象类型 / Galilean-satellite phenomenon type.
type GalileanPhenomenonType string

const (
	// GalileanPhenomenonTransit 凌日 / satellite transit across Jupiter.
	GalileanPhenomenonTransit GalileanPhenomenonType = "transit"
	// GalileanPhenomenonOccultation 掩蔽 / occultation behind Jupiter.
	GalileanPhenomenonOccultation GalileanPhenomenonType = "occultation"
	// GalileanPhenomenonEclipse 食 / eclipse in Jupiter's shadow.
	GalileanPhenomenonEclipse GalileanPhenomenonType = "eclipse"
	// GalileanPhenomenonShadowTransit 影凌 / shadow transit across Jupiter.
	GalileanPhenomenonShadowTransit GalileanPhenomenonType = "shadow_transit"
)

const (
	// GalileanSatelliteIo 木卫一 / Io.
	GalileanSatelliteIo = 1
	// GalileanSatelliteEuropa 木卫二 / Europa.
	GalileanSatelliteEuropa = 2
	// GalileanSatelliteGanymede 木卫三 / Ganymede.
	GalileanSatelliteGanymede = 3
	// GalileanSatelliteCallisto 木卫四 / Callisto.
	GalileanSatelliteCallisto = 4
)

// GalileanPhenomenonEvent 伽利略卫星整场现象 / full Galilean-satellite event.
//
// Start、Greatest、End 都保持调用者输入的时区。
// GreatestState 是食甚/现象最深时刻对应的瞬时现象标志与影心偏移。
// Start/End 是零半径模型的活动区间。有限圆盘接触用 GalileanPhenomenonContactEvent。
// Start, Greatest, and End preserve the caller's timezone.
// GreatestState contains the instantaneous phenomenon flags and shadow offsets at greatest event depth.
// Start/End are the active interval of the zero-radius model. Finite-disk contacts use GalileanPhenomenonContactEvent.
type GalileanPhenomenonEvent struct {
	Valid     bool
	Satellite int
	Type      GalileanPhenomenonType

	Start    time.Time
	Greatest time.Time
	End      time.Time

	Duration      time.Duration
	GreatestState GalileanSatellitePhenomenon
}

// LastGalileanPhenomenonEvent 上一次伽利略卫星现象 / previous Galilean-satellite event.
func (module *Jupiter) LastGalileanPhenomenonEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonEvent {
	return galileanPhenomenonEventFromBasic(
		basic.LastJupiterGalileanPhenomenonEventWithInstance(
			basic.Date2JD(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

// NextGalileanPhenomenonEvent 下一次伽利略卫星现象 / next Galilean-satellite event.
func (module *Jupiter) NextGalileanPhenomenonEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonEvent {
	return galileanPhenomenonEventFromBasic(
		basic.NextJupiterGalileanPhenomenonEventWithInstance(
			basic.Date2JD(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

// ClosestGalileanPhenomenonEvent 最近一次伽利略卫星现象 / closest Galilean-satellite event.
func (module *Jupiter) ClosestGalileanPhenomenonEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonEvent {
	return galileanPhenomenonEventFromBasic(
		basic.ClosestJupiterGalileanPhenomenonEventWithInstance(
			basic.Date2JD(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

func galileanPhenomenonEventFromBasic(event basic.JupiterGalileanPhenomenonEvent, loc *time.Location) GalileanPhenomenonEvent {
	if !event.Valid {
		return GalileanPhenomenonEvent{}
	}
	start := basic.JD2DateByZone(event.Start, loc, false)
	greatest := basic.JD2DateByZone(event.Greatest, loc, false)
	end := basic.JD2DateByZone(event.End, loc, false)
	return GalileanPhenomenonEvent{
		Valid:         true,
		Satellite:     event.Satellite,
		Type:          GalileanPhenomenonType(event.Type),
		Start:         start,
		Greatest:      greatest,
		End:           end,
		Duration:      end.Sub(start),
		GreatestState: galileanPhenomenonFromBasic(event.GreatestPhenomenon),
	}
}
