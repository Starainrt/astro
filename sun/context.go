package sun

import (
	"time"

	"github.com/starainrt/astro/internal/ephem"
)

// Sun 指定星历实例的计算模块 / computation module bound to an ephemeris instance.
type Sun struct{ inst *ephem.Instance }

// NewWithInstance 创建指定星历实例的模块，nil 使用全局模型 / create a module; nil uses the global model.
func NewWithInstance(inst *ephem.Instance) *Sun { return &Sun{inst: inst} }

var defaultModule Sun

// AxialtiltNutation 交角章动（IAU 2000B） / nutation in obliquity, IAU 2000B.
func (module *Sun) AxialtiltNutation(date time.Time) float64 { return AxialtiltNutation(date) }

// AxialtiltNutation1980 交角章动（IAU 1980） / nutation in obliquity, IAU 1980.
func (module *Sun) AxialtiltNutation1980(date time.Time) float64 { return AxialtiltNutation1980(date) }

// EclipticNutation 黄经章动（IAU 2000B） / nutation in longitude, IAU 2000B.
func (module *Sun) EclipticNutation(date time.Time) float64 { return EclipticNutation(date) }

// EclipticNutation1980 黄经章动（IAU 1980） / nutation in longitude, IAU 1980.
func (module *Sun) EclipticNutation1980(date time.Time) float64 { return EclipticNutation1980(date) }

// EclipticObliquity 黄赤交角 / ecliptic obliquity.
func (module *Sun) EclipticObliquity(date time.Time, nutation bool) float64 {
	return EclipticObliquity(date, nutation)
}

// GeometricLo 太阳几何黄经 / geometric ecliptic longitude.
func (module *Sun) GeometricLo(date time.Time) float64 { return GeometricLo(date) }

// MidFunc 太阳中心差 / solar equation of center.
func (module *Sun) MidFunc(date time.Time) float64 { return MidFunc(date) }
