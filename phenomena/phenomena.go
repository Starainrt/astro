// Package phenomena 提供任意两天体的黄经穿越、留、相位、最近角距与偕日可见搜索 / generic sky event and heliacal visibility searches.
package phenomena

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

// Body 可搜索的天体 / searchable body.
type Body = basic.SkyBody

// Earth 地球，只用于 Apsides / Earth, valid only for Apsides.
const Earth = basic.SkyEarth

const (
	// Sun 太阳 / the Sun.
	Sun = basic.SkySun
	// Moon 月球 / the Moon.
	Moon = basic.SkyMoon
	// Mercury 水星 / Mercury.
	Mercury = basic.SkyMercury
	// Venus 金星 / Venus.
	Venus = basic.SkyVenus
	// Mars 火星 / Mars.
	Mars = basic.SkyMars
	// Jupiter 木星 / Jupiter.
	Jupiter = basic.SkyJupiter
	// Saturn 土星 / Saturn.
	Saturn = basic.SkySaturn
	// Uranus 天王星 / Uranus.
	Uranus = basic.SkyUranus
	// Neptune 海王星 / Neptune.
	Neptune = basic.SkyNeptune
)

// StarCoordinate 恒星星表坐标 / catalog star coordinate.
type StarCoordinate = basic.StarCoordinate

// HeliacalOptions 偕日可见参数 / heliacal visibility options.
type HeliacalOptions = basic.HeliacalOptions

// HeliacalEventKind 偕日事件类型 / heliacal event kind.
type HeliacalEventKind = basic.HeliacalEventKind

const (
	// HeliacalMorningFirst 晨始见（偕日升） / morning first visibility.
	HeliacalMorningFirst = basic.HeliacalMorningFirst
	// HeliacalMorningLast 晨末见 / morning last visibility.
	HeliacalMorningLast = basic.HeliacalMorningLast
	// HeliacalEveningFirst 昏始见 / evening first visibility.
	HeliacalEveningFirst = basic.HeliacalEveningFirst
	// HeliacalEveningLast 昏末见（偕日落） / evening last visibility.
	HeliacalEveningLast = basic.HeliacalEveningLast
)

// Event 天象事件；Value 为事件对应的角量（度）/ sky event; Value is the event angle in degrees.
type Event struct {
	Time  time.Time
	Value float64
}

// HeliacalEvent 偕日事件，Time 为当次晨昏最佳可见时刻 / heliacal event at the best moment of that twilight.
type HeliacalEvent struct {
	Kind           HeliacalEventKind
	Time           time.Time
	Margin         float64
	TargetAltitude float64
	SunAltitude    float64
}

// Phenomena 指定星历实例的计算模块 / computation module bound to an ephemeris instance.
type Phenomena struct{ inst *ephem.Instance }

// NewWithInstance 创建指定星历实例的模块，nil 使用全局模型 / create a module; nil uses the global model.
func NewWithInstance(inst *ephem.Instance) *Phenomena { return &Phenomena{inst: inst} }

var defaultModule Phenomena

func toTT(date time.Time) float64 { return basic.UTC2TT(basic.Date2JD(date.UTC())) }

func events(list []basic.SkyEvent, loc *time.Location) []Event {
	out := make([]Event, 0, len(list))
	for _, e := range list {
		out = append(out, Event{Time: basic.JD2DateByZone(basic.TT2UTC(e.JDE), loc, false), Value: e.Value})
	}
	return out
}

// LongitudeCrossings 视黄经经过 longitude 的全部时刻 / times when the apparent longitude equals longitude.
func (module *Phenomena) LongitudeCrossings(body Body, longitude float64, start, end time.Time) []Event {
	return events(basic.SkyLongitudeCrossingsWithInstance(body, longitude, toTT(start), toTT(end), module.inst), start.Location())
}

// Stations 视黄经留，Value 为留时视黄经 / longitude stations; Value is the longitude.
func (module *Phenomena) Stations(body Body, start, end time.Time) []Event {
	return events(basic.SkyStationsWithInstance(body, toTT(start), toTT(end), module.inst), start.Location())
}

// AspectCrossings a 与 b 视黄经差（a − b）等于 angle 的时刻 / times when lon(a) − lon(b) equals angle.
func (module *Phenomena) AspectCrossings(a, b Body, angle float64, start, end time.Time) []Event {
	return events(basic.SkyAspectCrossingsWithInstance(a, b, angle, toTT(start), toTT(end), module.inst), start.Location())
}

// ClosestApproaches 两天体地心视角距的局部极小，Value 为角距（度）/ closest approaches within maxSeparation degrees.
func (module *Phenomena) ClosestApproaches(a, b Body, maxSeparation float64, start, end time.Time) []Event {
	return events(basic.SkySeparationMinimaWithInstance(a, b, maxSeparation, toTT(start), toTT(end), module.inst), start.Location())
}

// StarApproaches 天体与恒星地心视角距的局部极小 / closest approaches of a body to a star.
func (module *Phenomena) StarApproaches(body Body, star StarCoordinate, maxSeparation float64, start, end time.Time) []Event {
	return events(basic.SkyStarSeparationMinimaWithInstance(body, star, maxSeparation, toTT(start), toTT(end), module.inst), start.Location())
}

func heliacal(list []basic.HeliacalEvent, loc *time.Location) []HeliacalEvent {
	out := make([]HeliacalEvent, 0, len(list))
	for _, e := range list {
		out = append(out, HeliacalEvent{e.Kind, basic.JD2DateByZone(e.JD, loc, false), e.Margin, e.TargetAltitude, e.SunAltitude})
	}
	return out
}

// PlanetHeliacalEvents 行星晨昏始见、末见（经度东正）/ planet heliacal events; longitude east positive.
func (module *Phenomena) PlanetHeliacalEvents(body Body, start, end time.Time, lon, lat float64, options HeliacalOptions) []HeliacalEvent {
	return heliacal(basic.FindPlanetHeliacalEventsWithInstance(body, basic.Date2JD(start.UTC()), basic.Date2JD(end.UTC()), lon, lat, options, module.inst), start.Location())
}

// StarHeliacalEvents 恒星晨昏始见、末见；magnitude 为 V 星等 / star heliacal events with V magnitude.
func (module *Phenomena) StarHeliacalEvents(star StarCoordinate, magnitude float64, start, end time.Time, lon, lat float64, options HeliacalOptions) []HeliacalEvent {
	return heliacal(basic.FindStarHeliacalEventsWithInstance(star, magnitude, basic.Date2JD(start.UTC()), basic.Date2JD(end.UTC()), lon, lat, options, module.inst), start.Location())
}

// LimitingMagnitude 无月夜肉眼极限星等（Schaefer 模型），消光系数非正时为 NaN / naked-eye limiting magnitude without moonlight.
func LimitingMagnitude(targetAltitude, sunAltitude, separation, extinction float64) float64 {
	return basic.LimitingMagnitude(targetAltitude, sunAltitude, separation, extinction)
}

// OrbitalEventKind 轨道事件类型 / orbital event kind.
type OrbitalEventKind = basic.OrbitalEventKind

const (
	// Perihelion 近日点 / perihelion.
	Perihelion = basic.OrbitalPerihelion
	// Aphelion 远日点 / aphelion.
	Aphelion = basic.OrbitalAphelion
	// AscendingNode 升交点 / ascending node.
	AscendingNode = basic.OrbitalAscendingNode
	// DescendingNode 降交点 / descending node.
	DescendingNode = basic.OrbitalDescendingNode
)

// OrbitalEvent 轨道事件，DistanceAU 为日心距 / orbital event with heliocentric distance in AU.
type OrbitalEvent struct {
	Time       time.Time
	Kind       OrbitalEventKind
	DistanceAU float64
}

func orbital(list []basic.OrbitalEvent, loc *time.Location) []OrbitalEvent {
	out := make([]OrbitalEvent, 0, len(list))
	for _, e := range list {
		out = append(out, OrbitalEvent{Time: basic.JD2DateByZone(basic.TT2UTC(e.JDE), loc, false), Kind: e.Kind, DistanceAU: e.DistanceAU})
	}
	return out
}

// Apsides 行星或地球在 [start, end] 内的近日点与远日点 / perihelia and aphelia in the window.
func (module *Phenomena) Apsides(body Body, start, end time.Time) []OrbitalEvent {
	return orbital(basic.PlanetApsidesWithInstance(body, toTT(start), toTT(end), module.inst), start.Location())
}

// Nodes 行星在 [start, end] 内穿过历元平黄道的升交点与降交点 / ascending and descending node passages.
func (module *Phenomena) Nodes(body Body, start, end time.Time) []OrbitalEvent {
	return orbital(basic.PlanetNodesWithInstance(body, toTT(start), toTT(end), module.inst), start.Location())
}
