package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// GalileanPhenomenonContactPhase 接触阶段 / contact phase.
type GalileanPhenomenonContactPhase string

const (
	// GalileanPhenomenonContactDisappearance 初亏/初入接触阶段 / disappearance ingress contact.
	GalileanPhenomenonContactDisappearance GalileanPhenomenonContactPhase = "disappearance"
	// GalileanPhenomenonContactReappearance 复圆/复出接触阶段 / reappearance egress contact.
	GalileanPhenomenonContactReappearance GalileanPhenomenonContactPhase = "reappearance"
)

// GalileanPhenomenonContact 伽利略卫星接触窗口 / Galilean-satellite contact window.
//
// Start/End 是有限圆盘或有限影斑开始/结束接触的时刻；ModelCrossing 是这套连续接触模型下，
// 零半径参考点穿越边界的时刻。
// Start/End mark the beginning/end of the finite-disk or finite-shadow contact interval.
// ModelCrossing is the zero-radius boundary crossing in this continuous contact model.
type GalileanPhenomenonContact struct {
	Valid bool
	Phase GalileanPhenomenonContactPhase

	Start         time.Time
	ModelCrossing time.Time
	End           time.Time

	Duration time.Duration
}

// GalileanPhenomenonContactEvent IMCCE 风格的 D/F 接触事件 / IMCCE-style D/F contact event.
//
// 这个 API 返回有限圆盘/有限影斑的接触窗口，适合和 IMCCE 的 `TR.D/TR.F/OC.D/OC.F/EC.D/EC.F/SH.D/SH.F` 对齐；
// 现有 `GalileanPhenomenonEvent` 返回的是零半径几何模型保持 active 的整段区间，两者语义不同。
// 其中 `shadow_transit` 先用半影/本影边界求部分相持续时间，再把这段持续时间中心放在旧影轴过盘时刻上。
// This API returns finite-disk / finite-shadow contact windows and is intended to align with IMCCE
// `TR.D/TR.F/OC.D/OC.F/EC.D/EC.F/SH.D/SH.F` rows. The existing `GalileanPhenomenonEvent` returns the
// whole active interval of the zero-radius geometric model, so the semantics are different.
// For `shadow_transit`, the partial-phase duration comes from penumbra/umbra boundaries,
// while the reported D/F time is centered on the shadow-axis limb crossing from the existing full-event model.
type GalileanPhenomenonContactEvent struct {
	Valid     bool
	Satellite int
	Type      GalileanPhenomenonType

	Disappearance GalileanPhenomenonContact
	Greatest      time.Time
	Reappearance  GalileanPhenomenonContact

	GreatestState GalileanSatellitePhenomenon
}

// LastGalileanPhenomenonContactEvent 上一次 IMCCE 风格接触事件 / previous IMCCE-style contact event.
func LastGalileanPhenomenonContactEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonContactEvent {
	return galileanPhenomenonContactEventFromBasic(
		basic.LastJupiterGalileanPhenomenonContactEvent(
			basic.Date2JDE(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType),
		),
		date.Location(),
	)
}

// NextGalileanPhenomenonContactEvent 下一次 IMCCE 风格接触事件 / next IMCCE-style contact event.
func NextGalileanPhenomenonContactEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonContactEvent {
	return galileanPhenomenonContactEventFromBasic(
		basic.NextJupiterGalileanPhenomenonContactEvent(
			basic.Date2JDE(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType),
		),
		date.Location(),
	)
}

// ClosestGalileanPhenomenonContactEvent 最近一次 IMCCE 风格接触事件 / closest IMCCE-style contact event.
func ClosestGalileanPhenomenonContactEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonContactEvent {
	return galileanPhenomenonContactEventFromBasic(
		basic.ClosestJupiterGalileanPhenomenonContactEvent(
			basic.Date2JDE(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType),
		),
		date.Location(),
	)
}

func galileanPhenomenonContactEventFromBasic(event basic.JupiterGalileanPhenomenonContactEvent, loc *time.Location) GalileanPhenomenonContactEvent {
	if !event.Valid {
		return GalileanPhenomenonContactEvent{}
	}
	greatest := basic.JDE2DateByZone(event.Greatest, loc, false)
	return GalileanPhenomenonContactEvent{
		Valid:         true,
		Satellite:     event.Satellite,
		Type:          GalileanPhenomenonType(event.Type),
		Disappearance: galileanPhenomenonContactFromBasic(event.Disappearance, loc),
		Greatest:      greatest,
		Reappearance:  galileanPhenomenonContactFromBasic(event.Reappearance, loc),
		GreatestState: galileanPhenomenonFromBasic(event.GreatestPhenomenon),
	}
}

func galileanPhenomenonContactFromBasic(contact basic.JupiterGalileanPhenomenonContact, loc *time.Location) GalileanPhenomenonContact {
	if !contact.Valid {
		return GalileanPhenomenonContact{}
	}
	start := basic.JDE2DateByZone(contact.Start, loc, false)
	modelCrossing := basic.JDE2DateByZone(contact.ModelCrossing, loc, false)
	end := basic.JDE2DateByZone(contact.End, loc, false)
	return GalileanPhenomenonContact{
		Valid:         true,
		Phase:         GalileanPhenomenonContactPhase(contact.Phase),
		Start:         start,
		ModelCrossing: modelCrossing,
		End:           end,
		Duration:      end.Sub(start),
	}
}
