package jupiter

import (
	"math"
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
// 常规窗口的 Start/End 来自有限圆盘或有限影斑模型；ModelCrossing 是零半径参考点穿越边界的时刻。
// For regular windows, Start/End come from the finite-disk or finite-shadow model.
// ModelCrossing is the zero-radius boundary crossing in this continuous contact model.
type GalileanPhenomenonContact struct {
	Valid bool
	Phase GalileanPhenomenonContactPhase

	// PenumbralFallback 备用半影时段，以事件 Greatest 为界 / penumbral segments split at the event Greatest.
	PenumbralFallback bool

	Start         time.Time
	ModelCrossing time.Time // 参考点不穿越边界时为零值 / zero if the reference point does not cross the boundary.
	End           time.Time

	Duration time.Duration
}

// GalileanPhenomenonContactEvent IMCCE 风格的 D/F 接触事件 / IMCCE-style D/F contact event.
//
// 常规窗口对应 IMCCE 的 TR/OC/EC/SH 接触行。影凌的部分相时长来自半影和本影边界，时刻中心在影轴过盘点上。
// Regular windows match the IMCCE TR/OC/EC/SH contact rows. Shadow-transit partial duration uses the penumbra and umbra edges, centered on the shadow-axis limb crossing.
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
func (module *Jupiter) LastGalileanPhenomenonContactEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonContactEvent {
	return galileanPhenomenonContactEventFromBasic(
		basic.LastJupiterGalileanPhenomenonContactEventWithInstance(
			basic.Date2JD(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

// NextGalileanPhenomenonContactEvent 下一次 IMCCE 风格接触事件 / next IMCCE-style contact event.
func (module *Jupiter) NextGalileanPhenomenonContactEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonContactEvent {
	return galileanPhenomenonContactEventFromBasic(
		basic.NextJupiterGalileanPhenomenonContactEventWithInstance(
			basic.Date2JD(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

// ClosestGalileanPhenomenonContactEvent 最近一次 IMCCE 风格接触事件 / closest IMCCE-style contact event.
func (module *Jupiter) ClosestGalileanPhenomenonContactEvent(date time.Time, satellite int, phenomenonType GalileanPhenomenonType) GalileanPhenomenonContactEvent {
	return galileanPhenomenonContactEventFromBasic(
		basic.ClosestJupiterGalileanPhenomenonContactEventWithInstance(
			basic.Date2JD(date.UTC()),
			satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType), module.inst,
		),
		date.Location(),
	)
}

func galileanPhenomenonContactEventFromBasic(event basic.JupiterGalileanPhenomenonContactEvent, loc *time.Location) GalileanPhenomenonContactEvent {
	if !event.Valid {
		return GalileanPhenomenonContactEvent{}
	}
	greatest := basic.JD2DateByZone(event.Greatest, loc, false)
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
	start := basic.JD2DateByZone(contact.Start, loc, false)
	var modelCrossing time.Time
	if !math.IsNaN(contact.ModelCrossing) {
		modelCrossing = basic.JD2DateByZone(contact.ModelCrossing, loc, false)
	}
	end := basic.JD2DateByZone(contact.End, loc, false)
	return GalileanPhenomenonContact{
		Valid:             true,
		Phase:             GalileanPhenomenonContactPhase(contact.Phase),
		PenumbralFallback: contact.PenumbralFallback,
		Start:             start,
		ModelCrossing:     modelCrossing,
		End:               end,
		Duration:          end.Sub(start),
	}
}
