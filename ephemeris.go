package astro

import (
	"errors"

	"github.com/starainrt/astro/internal/ephem"
)

// Instance 显式星历实例，不改变进程默认模型 / explicit ephemeris instance independent of the process default.
type Instance = ephem.Instance

// Engine 只读 ICRF 几何星历，位置 km、速度 km/日，历元 TDB / read-only ICRF ephemeris in km and km/day at TDB epochs.
type Engine = ephem.Engine

// Stack 按顺序叠加多个只读星历引擎 / ordered stack of read-only ephemeris engines.
type Stack = ephem.Stack

// ApparentEngine 可选的地心 ICRF 单位视方向能力，历元 TT / optional apparent geocentric ICRF unit directions at TT epochs.
type ApparentEngine = ephem.ApparentEngine

// Body NAIF 天体编号 / NAIF body identifier.
type Body = ephem.Body

const (
	// BodySSB 太阳系质心 / solar system barycenter.
	BodySSB = ephem.SSB
	// BodyMercury 水星 / Mercury.
	BodyMercury = ephem.Mercury
	// BodyVenus 金星 / Venus.
	BodyVenus = ephem.Venus
	// BodyEMB 地月质心 / Earth-Moon barycenter.
	BodyEMB = ephem.EMB
	// BodyMars 火星 / Mars.
	BodyMars = ephem.Mars
	// BodyJupiter 木星 / Jupiter.
	BodyJupiter = ephem.Jupiter
	// BodySaturn 土星 / Saturn.
	BodySaturn = ephem.Saturn
	// BodyUranus 天王星 / Uranus.
	BodyUranus = ephem.Uranus
	// BodyNeptune 海王星 / Neptune.
	BodyNeptune = ephem.Neptune
	// BodySun 太阳 / Sun.
	BodySun = ephem.Sun
	// BodyMoon 月球 / Moon.
	BodyMoon = ephem.Moon
	// BodyEarth 地球 / Earth.
	BodyEarth = ephem.Earth
)

// EphemerisModel 内置或自定义星历模型标识 / built-in or custom ephemeris model identifier.
type EphemerisModel = ephem.Model

const (
	// EphemerisLegacy 旧级数模型 / legacy series model.
	EphemerisLegacy = ephem.Legacy
	// EphemerisSemianalytic 半解析模型 / semi-analytic model.
	EphemerisSemianalytic = ephem.Semianalytic
	// EphemerisCustom 自定义几何引擎 / custom geometric engine.
	EphemerisCustom = ephem.Custom
)

var (
	// ErrEphemerisOutOfRange 超出星历覆盖 / epoch outside ephemeris coverage.
	ErrEphemerisOutOfRange = ephem.ErrOutOfRange
	// ErrEphemerisUnsupportedBody 引擎不含该天体 / body not provided by the engine.
	ErrEphemerisUnsupportedBody = ephem.ErrUnsupportedBody
	// ErrEphemerisNoEngine 实例没有几何引擎 / instance has no geometric engine.
	ErrEphemerisNoEngine = ephem.ErrNoEngine
	// ErrEphemerisUnsupportedApparent 引擎不提供视位置 / apparent positions are unsupported.
	ErrEphemerisUnsupportedApparent = ephem.ErrUnsupportedApparent
	// ErrUnknownModel 模型不是本函数接受的模型 / model is not accepted by this function.
	ErrUnknownModel = errors.New("astro: unknown or unsupported ephemeris model")
	// ErrInvalidEngine 引擎个数或取值不符合模型要求 / engine count or value does not fit the model.
	ErrInvalidEngine = errors.New("astro: invalid ephemeris engine")
)

// NewInstance 创建显式自定义星历实例 / create an explicit custom ephemeris instance.
func NewInstance(engine Engine) *Instance {
	return ephem.NewInstance(engine)
}

// NewStack 创建多源星历，先传入的引擎优先 / creates a multi-source ephemeris; earlier engines have priority.
func NewStack(engines ...Engine) *Stack {
	return ephem.NewStack(engines...)
}

// NewStackInstance 创建使用多源星历的显式实例 / creates an explicit instance backed by multiple ephemeris sources.
func NewStackInstance(engines ...Engine) *Instance {
	return ephem.NewInstance(ephem.NewStack(engines...))
}

// NewModelInstance 创建内置模型实例，非内置模型（含 EphemerisCustom）返回 ErrUnknownModel；旧模型目前仅支持标量入口 /
// create a built-in model instance; non-built-in models including EphemerisCustom return ErrUnknownModel.
func NewModelInstance(model EphemerisModel) (*Instance, error) {
	if model != EphemerisLegacy && model != EphemerisSemianalytic {
		return nil, ErrUnknownModel
	}
	return ephem.NewModelInstance(model), nil
}
