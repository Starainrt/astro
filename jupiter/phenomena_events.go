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
// 这里的 Start/End 基于零半径几何模型是否 active，不是 IMCCE 年表中的 D/F 接触时刻；
// 如果需要有限圆盘/有限影斑的接触窗口，请改用 `GalileanPhenomenonContactEvent`。
// Start, Greatest, and End preserve the caller's timezone.
// GreatestState contains the instantaneous phenomenon flags and shadow offsets at greatest event depth.
// Start/End here are based on whether the zero-radius geometric model is active, not the IMCCE D/F contact times.
// Use `GalileanPhenomenonContactEvent` when you need finite-disk or finite-shadow contact windows.
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
//
// date 表示查询绝对时刻；satellite 取 `1=Io, 2=Europa, 3=Ganymede, 4=Callisto`；
// phenomenonType 取 `transit/occultation/eclipse/shadow_transit` 中之一。
// date is the query instant; satellite is `1=Io, 2=Europa, 3=Ganymede, 4=Callisto`;
// phenomenonType is one of `transit/occultation/eclipse/shadow_transit`.
func LastGalileanPhenomenonEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonEvent {
	return galileanPhenomenonEventFromBasic(
		basic.LastJupiterGalileanPhenomenonEvent(
			basic.Date2JDE(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType),
		),
		date.Location(),
	)
}

// NextGalileanPhenomenonEvent 下一次伽利略卫星现象 / next Galilean-satellite event.
func NextGalileanPhenomenonEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonEvent {
	return galileanPhenomenonEventFromBasic(
		basic.NextJupiterGalileanPhenomenonEvent(
			basic.Date2JDE(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType),
		),
		date.Location(),
	)
}

// ClosestGalileanPhenomenonEvent 最近一次伽利略卫星现象 / closest Galilean-satellite event.
func ClosestGalileanPhenomenonEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonEvent {
	return galileanPhenomenonEventFromBasic(
		basic.ClosestJupiterGalileanPhenomenonEvent(
			basic.Date2JDE(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType),
		),
		date.Location(),
	)
}

func galileanPhenomenonEventFromBasic(event basic.JupiterGalileanPhenomenonEvent, loc *time.Location) GalileanPhenomenonEvent {
	if !event.Valid {
		return GalileanPhenomenonEvent{}
	}
	start := basic.JDE2DateByZone(event.Start, loc, false)
	greatest := basic.JDE2DateByZone(event.Greatest, loc, false)
	end := basic.JDE2DateByZone(event.End, loc, false)
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
