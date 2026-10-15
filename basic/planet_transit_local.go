package basic

import (
	"math"
	"sort"
	"sync"
	"time"
)

// 站心凌日：以地心凌日为种子，对太阳与行星的视赤经赤纬分别做观测者视差修正后重解接触与凌甚。
// 太阳视半径按站心距离缩放（影响最多约 0.04″）；行星视半径沿用地心值，站心修正低于 0.01″。
// 结果里的太阳高度一律为几何中心高度；可见性、日出日落与等时线裁剪按太阳中心视高度 0° 判定，
// 即计入标准大气（1010 hPa、10 °C）下的地平折射，几何高度阈值约 -34′。

const (
	// planetTransitLocalWindowDays 是站心接触相对地心接触的搜索半宽。金星视差差约 22″，
	// 接触偏移最多约十分钟；四十分钟同时容纳近掠凌时变慢的接触。
	planetTransitLocalWindowDays = 40.0 / 1440
	// planetTransitLocalScanStepDays 是站点求根前的分档步长。
	planetTransitLocalScanStepDays = 1.0 / 1440
	// planetTransitLocalToleranceDays 是站点接触与凌甚的收敛容差。
	planetTransitLocalToleranceDays = 0.05 / 86400
	// planetTransitLocalSunStepDays 是判定凌日期间日出日落的采样步长。
	planetTransitLocalSunStepDays = 2.0 / 1440
)

// planetTransitSunHorizonAltitudeDeg 是太阳中心视高度为 0° 时的几何高度（标准大气地平折射取负），单位度。
var planetTransitSunHorizonAltitudeDeg = -RefractionFromApparentAltitude(0, refractionStandardPressureHPa, refractionStandardTemperatureC)

// PlanetTransitContactKind 标识凌日接触 / identifies a transit contact.
type PlanetTransitContactKind int

const (
	// PlanetTransitC1 是外切入凌（一触） / external ingress (first contact).
	PlanetTransitC1 PlanetTransitContactKind = 1 + iota
	// PlanetTransitC2 是内切入凌（二触） / internal ingress (second contact).
	PlanetTransitC2
	// PlanetTransitC3 是内切出凌（三触） / internal egress (third contact).
	PlanetTransitC3
	// PlanetTransitC4 是外切出凌（四触） / external egress (fourth contact).
	PlanetTransitC4
	// PlanetTransitGreatest 是凌甚 / greatest transit.
	PlanetTransitGreatest
)

// PlanetTransitLocalContact 是站心凌日的一个接触或凌甚时刻。
// PlanetTransitLocalContact is one topocentric transit contact or the greatest instant.
type PlanetTransitLocalContact struct {
	// Valid 表示该时刻存在 / Valid reports whether the instant exists.
	Valid bool
	// JD 是 UTC 儒略日 / JD is the UTC Julian day.
	JD float64
	// SunAltitudeDeg 与 SunAzimuthDeg 是站心太阳中心几何高度与方位（北起向东），单位度。
	// SunAltitudeDeg and SunAzimuthDeg are the topocentric geometric altitude and azimuth (from north through east) of the Sun's centre.
	SunAltitudeDeg float64
	SunAzimuthDeg  float64
	// PositionAngleDeg 是行星中心相对日心的位置角，北起向东，单位度。
	// PositionAngleDeg is the position angle of the planet centre from the solar centre, measured from north through east.
	PositionAngleDeg float64
}

// PlanetTransitLocalResult 是一个观测地点的站心凌日结果。
// PlanetTransitLocalResult is the topocentric transit seen from one observer.
type PlanetTransitLocalResult struct {
	// Valid 为 false 表示该地点看不到行星圆面与太阳圆面相交（几何上，与太阳是否在地平上无关）。
	// Valid is false when the planet and solar disks do not overlap from this site, regardless of the Sun's altitude.
	Valid bool
	// PlanetIndex 为 1 表示水星，2 表示金星 / PlanetIndex is 1 for Mercury and 2 for Venus.
	PlanetIndex int
	Observer    Observer
	// 四触与凌甚；掠凌没有内切接触时 InternalIngress、InternalEgress 的 Valid 为 false。
	// Contacts and greatest; a grazing transit has invalid internal contacts.
	ExternalIngress PlanetTransitLocalContact
	InternalIngress PlanetTransitLocalContact
	Greatest        PlanetTransitLocalContact
	InternalEgress  PlanetTransitLocalContact
	ExternalEgress  PlanetTransitLocalContact
	HasInternal     bool
	// 凌甚时的站心角距与视半径，单位角秒；太阳视半径为站心值，行星视半径沿用地心值（站心差约 0.006″，忽略）。
	// Separation and semidiameters at greatest, in arcseconds: separation and solar semidiameter are topocentric,
	// the planet semidiameter keeps the geocentric value (topocentric difference ≈0.006″, neglected).
	MinimumSeparationArcsec  float64
	SunSemidiameterArcsec    float64
	PlanetSemidiameterArcsec float64
	// Visible 表示一触到四触之间太阳中心至少有一刻在视地平上（计入标准大气折射）。
	// Visible reports whether the Sun's centre is above the apparent horizon (standard refraction) at some instant between C1 and C4.
	Visible bool
	// Sunrise 与 Sunset 是一触到四触之间太阳中心第一次升出、落入视地平的 UTC 儒略日；没有时为 0。
	// Sunrise and Sunset are the first UTC Julian days between C1 and C4 when the Sun's centre crosses the apparent horizon, or 0 when absent.
	Sunrise float64
	Sunset  float64
}

// PlanetTransitLocal 返回地心凌日 transit 在观测者 observer 处的站心接触与凌甚。
// PlanetTransitLocal returns the topocentric contacts and greatest instant of the geocentric transit seen by observer.
func PlanetTransitLocal(transit PlanetTransitResult, observer Observer) PlanetTransitLocalResult {
	return PlanetTransitLocalWithInstance(transit, observer, nil)
}

// PlanetTransitLocalWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func PlanetTransitLocalWithInstance(transit PlanetTransitResult, observer Observer, inst *Instance) PlanetTransitLocalResult {
	result := PlanetTransitLocalResult{PlanetIndex: transit.PlanetIndex, Observer: observer}
	if observer.Validate() != nil {
		return result
	}
	snapshot, ok := newPlanetTransitSnapshot(transit, inst)
	if !ok {
		return result
	}
	return snapshot.local(transit, observer)
}

// planetTransitSnapshot 缓存凌日窗口内的地心太阳、行星位置与恒星时，站点求解只做插值与视差修正。
// 等时线要在全球网格上逐点求接触，直接调用星历会慢上千倍。
type planetTransitSnapshot struct {
	planetIndex int
	startTT     float64
	step        float64
	// 按列存储：太阳赤经、赤纬、距离，行星赤经、赤纬、距离，太阳与行星视半径，格林尼治视恒星时（度）。
	// 赤经与恒星时已展开为连续值，避免插值跨越 0°/360°。
	columns [planetTransitSnapshotColumns][]float64
}

const (
	planetTransitSnapshotSunRA = iota
	planetTransitSnapshotSunDec
	planetTransitSnapshotSunDistance
	planetTransitSnapshotPlanetRA
	planetTransitSnapshotPlanetDec
	planetTransitSnapshotPlanetDistance
	planetTransitSnapshotSunRadius
	planetTransitSnapshotPlanetRadius
	planetTransitSnapshotSidereal
	planetTransitSnapshotColumns
)

const (
	// planetTransitSnapshotStepDays 是缓存节点间距；四点拉格朗日插值在十分钟间距下误差远低于 1e-6″。
	planetTransitSnapshotStepDays = 10.0 / 1440
	// planetTransitSnapshotMarginDays 是缓存相对地心一触、四触向外扩展的余量。
	planetTransitSnapshotMarginDays = 2.0 / 24
	// planetTransitMaxSpanDays 是地心一触到四触的跨度上限。实际凌日最长约 8 小时（水星）、
	// 金星不到 9 小时；超出视为异常输入，避免按跨度分配巨大缓存。
	planetTransitMaxSpanDays = 1.0
)

// planetTransitSnapshotCache 保留最近一次构建的快照：全球图四类等时线、站心结果与日面帧
// 对同一凌日连续调用，复用可省掉重复的星历求值。快照构建后只读，可并发共享。
var planetTransitSnapshotCache struct {
	sync.Mutex
	transit  PlanetTransitResult
	inst     *Instance
	snapshot *planetTransitSnapshot
}

func newPlanetTransitSnapshot(transit PlanetTransitResult, inst *Instance) (*planetTransitSnapshot, bool) {
	planetTransitSnapshotCache.Lock()
	if planetTransitSnapshotCache.snapshot != nil && planetTransitSnapshotCache.transit == transit && planetTransitSnapshotCache.inst == inst {
		snapshot := planetTransitSnapshotCache.snapshot
		planetTransitSnapshotCache.Unlock()
		return snapshot, true
	}
	planetTransitSnapshotCache.Unlock()
	snapshot, ok := buildPlanetTransitSnapshot(transit, inst)
	if ok {
		planetTransitSnapshotCache.Lock()
		planetTransitSnapshotCache.transit, planetTransitSnapshotCache.inst, planetTransitSnapshotCache.snapshot = transit, inst, snapshot
		planetTransitSnapshotCache.Unlock()
	}
	return snapshot, ok
}

func buildPlanetTransitSnapshot(transit PlanetTransitResult, inst *Instance) (*planetTransitSnapshot, bool) {
	if !transit.Valid || !transit.HasExternal || !isFiniteFloat(transit.ExternalIngress) || !isFiniteFloat(transit.ExternalEgress) ||
		!isFiniteFloat(transit.Greatest) || transit.ExternalEgress <= transit.ExternalIngress ||
		transit.ExternalEgress-transit.ExternalIngress > planetTransitMaxSpanDays ||
		transit.Greatest < transit.ExternalIngress || transit.Greatest > transit.ExternalEgress {
		return nil, false
	}
	var cfg planetTransitConfig
	switch transit.PlanetIndex {
	case 1:
		cfg = mercuryTransitConfig(inst)
	case 2:
		cfg = venusTransitConfig(inst)
	default:
		return nil, false
	}
	startTT := UTC2TT(transit.ExternalIngress) - planetTransitSnapshotMarginDays
	endTT := UTC2TT(transit.ExternalEgress) + planetTransitSnapshotMarginDays
	count := int(math.Ceil((endTT-startTT)/planetTransitSnapshotStepDays)) + 1
	snapshot := &planetTransitSnapshot{planetIndex: transit.PlanetIndex, startTT: startTT, step: planetTransitSnapshotStepDays}
	for column := range snapshot.columns {
		snapshot.columns[column] = make([]float64, count)
	}
	for index := 0; index < count; index++ {
		tt := startTT + float64(index)*planetTransitSnapshotStepDays
		sunRA, sunDec := HSunApparentRaDecNWithInstance(tt, -1, inst)
		planetRA, planetDec := cfg.apparentRaDecN(tt, -1)
		values := [planetTransitSnapshotColumns]float64{
			sunRA, sunDec, EarthAwayNWithInstance(tt, -1, inst),
			planetRA, planetDec, cfg.earthDistanceN(tt, -1),
			SunSemidiameterNWithInstance(tt, -1, inst), cfg.semidiameterN(tt, -1),
			ApparentSiderealTime(TT2UT1(tt)) * 15,
		}
		for column, value := range values {
			if !isFiniteFloat(value) {
				return nil, false
			}
			snapshot.columns[column][index] = value
		}
	}
	for _, column := range []int{planetTransitSnapshotSunRA, planetTransitSnapshotPlanetRA, planetTransitSnapshotSidereal} {
		unwrapDegrees(snapshot.columns[column])
	}
	return snapshot, true
}

// unwrapDegrees 把相邻差值折回 (-180°,180°]，得到连续的角度序列。
func unwrapDegrees(values []float64) {
	for index := 1; index < len(values); index++ {
		difference := math.Mod(values[index]-values[index-1], 360)
		if difference > 180 {
			difference -= 360
		} else if difference <= -180 {
			difference += 360
		}
		values[index] = values[index-1] + difference
	}
}

// interpolate 对缓存做四点拉格朗日插值；越出缓存范围时返回 false。
func (snapshot *planetTransitSnapshot) interpolate(tt float64) ([planetTransitSnapshotColumns]float64, bool) {
	var values [planetTransitSnapshotColumns]float64
	count := len(snapshot.columns[0])
	position := (tt - snapshot.startTT) / snapshot.step
	if !isFiniteFloat(position) || position < 0 || position > float64(count-1) {
		return values, false
	}
	first := int(math.Floor(position)) - 1
	if first < 0 {
		first = 0
	}
	if first > count-4 {
		first = count - 4
	}
	u := position - float64(first)
	weights := [4]float64{
		-(u - 1) * (u - 2) * (u - 3) / 6,
		u * (u - 2) * (u - 3) / 2,
		-u * (u - 1) * (u - 3) / 2,
		u * (u - 1) * (u - 2) / 6,
	}
	for column := range values {
		data := snapshot.columns[column]
		values[column] = weights[0]*data[first] + weights[1]*data[first+1] + weights[2]*data[first+2] + weights[3]*data[first+3]
	}
	return values, true
}

// planetTransitLocalState 是一个站点在一个时刻的站心日面几何。
type planetTransitLocalState struct {
	separationArcsec float64
	positionAngleDeg float64
	sunRadiusArcsec  float64
	planetRadius     float64
	sunAltitudeDeg   float64
	sunAzimuthDeg    float64
	external         float64
	internal         float64
}

func (snapshot *planetTransitSnapshot) stateAt(tt float64, observer Observer) (planetTransitLocalState, bool) {
	values, ok := snapshot.interpolate(tt)
	if !ok {
		return planetTransitLocalState{}, false
	}
	sidereal := values[planetTransitSnapshotSidereal]
	sunRA, sunDec := topocentricRaDecWithSidereal(values[planetTransitSnapshotSunRA], values[planetTransitSnapshotSunDec],
		observer.Latitude, observer.Longitude, sidereal, values[planetTransitSnapshotSunDistance], observer.Height)
	planetRA, planetDec := topocentricRaDecWithSidereal(values[planetTransitSnapshotPlanetRA], values[planetTransitSnapshotPlanetDec],
		observer.Latitude, observer.Longitude, sidereal, values[planetTransitSnapshotPlanetDistance], observer.Height)
	// 以日心为切点的球面偏移：角距用 atan2 求，几百角秒的小角也不丢精度。
	deltaRA := (planetRA - sunRA) * math.Pi / 180
	sunDecRad := sunDec * math.Pi / 180
	planetDecRad := planetDec * math.Pi / 180
	east := math.Cos(planetDecRad) * math.Sin(deltaRA)
	north := math.Cos(sunDecRad)*math.Sin(planetDecRad) - math.Sin(sunDecRad)*math.Cos(planetDecRad)*math.Cos(deltaRA)
	along := math.Sin(sunDecRad)*math.Sin(planetDecRad) + math.Cos(sunDecRad)*math.Cos(planetDecRad)*math.Cos(deltaRA)
	separation := math.Atan2(math.Hypot(east, north), along) * 180 / math.Pi * 3600
	// 太阳视半径按站心距离缩放：站心距离比地心近 ρ·cos z，最多约 4.3e-5，相当于 0.04″。
	// 行星同一比例约 2e-4·r ≈ 0.006″（金星），可忽略。
	geocentricHourAngle := (sidereal + observer.Longitude - values[planetTransitSnapshotSunRA]) * math.Pi / 180
	geocentricSunDec := values[planetTransitSnapshotSunDec] * math.Pi / 180
	sunDistanceRatio := 1 - math.Sin(0.0024427777777*math.Pi/180)/values[planetTransitSnapshotSunDistance]*
		(psini(observer.Latitude, observer.Height)*math.Sin(geocentricSunDec)+
			pcosi(observer.Latitude, observer.Height)*math.Cos(geocentricSunDec)*math.Cos(geocentricHourAngle))
	sunRadius := values[planetTransitSnapshotSunRadius] / sunDistanceRatio
	planetRadius := values[planetTransitSnapshotPlanetRadius]
	hourAngle := signedAngleDifference(sidereal+observer.Longitude, sunRA) * math.Pi / 180
	latitude := observer.Latitude * math.Pi / 180
	azimuth := normalizeRA(math.Atan2(math.Sin(hourAngle),
		math.Cos(hourAngle)*math.Sin(latitude)-math.Tan(sunDecRad)*math.Cos(latitude))*180/math.Pi + 180)
	return planetTransitLocalState{
		separationArcsec: separation,
		positionAngleDeg: normalizeRA(math.Atan2(east, north) * 180 / math.Pi),
		sunRadiusArcsec:  sunRadius,
		planetRadius:     planetRadius,
		sunAltitudeDeg:   occultationAltitudeWithSidereal(sidereal, observer, sunRA, sunDec),
		sunAzimuthDeg:    azimuth,
		external:         separation - (sunRadius + planetRadius),
		internal:         separation - (sunRadius - planetRadius),
	}, true
}

func (snapshot *planetTransitSnapshot) local(transit PlanetTransitResult, observer Observer) PlanetTransitLocalResult {
	result := PlanetTransitLocalResult{PlanetIndex: snapshot.planetIndex, Observer: observer}
	greatestTT, ok := snapshot.localGreatestTT(UTC2TT(transit.Greatest), observer)
	if !ok {
		return result
	}
	greatest, ok := snapshot.stateAt(greatestTT, observer)
	if !ok || greatest.external >= 0 {
		return result
	}
	contact := func(kind PlanetTransitContactKind) (PlanetTransitLocalContact, bool) {
		tt, ok := snapshot.localContactTT(kind, planetTransitGeocentricContactTT(transit, kind, greatestTT), greatestTT, observer)
		if !ok {
			return PlanetTransitLocalContact{}, false
		}
		return snapshot.contactAt(tt, observer)
	}
	if result.ExternalIngress, ok = contact(PlanetTransitC1); !ok {
		return result
	}
	if result.ExternalEgress, ok = contact(PlanetTransitC4); !ok {
		return result
	}
	if greatest.internal < 0 {
		c2, ok2 := contact(PlanetTransitC2)
		c3, ok3 := contact(PlanetTransitC3)
		if ok2 && ok3 {
			result.InternalIngress, result.InternalEgress, result.HasInternal = c2, c3, true
		}
	}
	result.Greatest, _ = snapshot.contactAt(greatestTT, observer)
	result.Valid = true
	result.MinimumSeparationArcsec = greatest.separationArcsec
	result.SunSemidiameterArcsec = greatest.sunRadiusArcsec
	result.PlanetSemidiameterArcsec = greatest.planetRadius
	snapshot.localSunEvents(&result, UTC2TT(result.ExternalIngress.JD), UTC2TT(result.ExternalEgress.JD), observer)
	return result
}

func (snapshot *planetTransitSnapshot) contactAt(tt float64, observer Observer) (PlanetTransitLocalContact, bool) {
	state, ok := snapshot.stateAt(tt, observer)
	if !ok {
		return PlanetTransitLocalContact{}, false
	}
	return PlanetTransitLocalContact{
		Valid:            true,
		JD:               TT2UTC(tt),
		SunAltitudeDeg:   state.sunAltitudeDeg,
		SunAzimuthDeg:    state.sunAzimuthDeg,
		PositionAngleDeg: state.positionAngleDeg,
	}, true
}

// planetTransitGeocentricContactTT 给出站点求根的种子：地心有该接触时取地心时刻，否则取站心凌甚。
func planetTransitGeocentricContactTT(transit PlanetTransitResult, kind PlanetTransitContactKind, fallbackTT float64) float64 {
	var jd float64
	switch kind {
	case PlanetTransitC1:
		jd = transit.ExternalIngress
	case PlanetTransitC2:
		jd = transit.InternalIngress
	case PlanetTransitC3:
		jd = transit.InternalEgress
	case PlanetTransitC4:
		jd = transit.ExternalEgress
	}
	if jd == 0 {
		return fallbackTT
	}
	return UTC2TT(jd)
}

func (snapshot *planetTransitSnapshot) localGreatestTT(seedTT float64, observer Observer) (float64, bool) {
	const goldenRatio = 0.6180339887498949
	separation := func(tt float64) float64 {
		state, ok := snapshot.stateAt(tt, observer)
		if !ok {
			return math.Inf(1)
		}
		return state.separationArcsec
	}
	left, right := seedTT-planetTransitLocalWindowDays, seedTT+planetTransitLocalWindowDays
	x1, x2 := right-goldenRatio*(right-left), left+goldenRatio*(right-left)
	f1, f2 := separation(x1), separation(x2)
	for iteration := 0; iteration < 100 && right-left > planetTransitLocalToleranceDays; iteration++ {
		if f1 <= f2 {
			right, x2, f2 = x2, x1, f1
			x1 = right - goldenRatio*(right-left)
			f1 = separation(x1)
		} else {
			left, x1, f1 = x1, x2, f2
			x2 = left + goldenRatio*(right-left)
			f2 = separation(x2)
		}
	}
	center := (left + right) / 2
	return center, isFiniteFloat(separation(center))
}

// localContactTT 从种子出发按分钟分档找变号区间再二分。入凌时接触量随时间由正变负，出凌相反；
// 种子在盘内时向外走，在盘外时向凌甚走，两种情形都不越过凌甚。
func (snapshot *planetTransitSnapshot) localContactTT(
	kind PlanetTransitContactKind,
	seedTT, greatestTT float64,
	observer Observer,
) (float64, bool) {
	internal := kind == PlanetTransitC2 || kind == PlanetTransitC3
	outward := 1.0
	if kind == PlanetTransitC1 || kind == PlanetTransitC2 {
		outward = -1
	}
	metric := func(tt float64) float64 {
		state, ok := snapshot.stateAt(tt, observer)
		if !ok {
			return math.NaN()
		}
		if internal {
			return state.internal
		}
		return state.external
	}
	if (seedTT-greatestTT)*outward < 0 {
		seedTT = greatestTT
	}
	seedValue := metric(seedTT)
	if !isFiniteFloat(seedValue) {
		return 0, false
	}
	direction := outward
	if seedValue > 0 {
		direction = -outward
	}
	limit := int(math.Ceil(planetTransitLocalWindowDays / planetTransitLocalScanStepDays))
	near, nearValue := seedTT, seedValue
	for step := 1; step <= limit; step++ {
		far := seedTT + direction*float64(step)*planetTransitLocalScanStepDays
		if direction != outward && (far-greatestTT)*outward < 0 {
			far = greatestTT
		}
		farValue := metric(far)
		if !isFiniteFloat(farValue) {
			return 0, false
		}
		if (nearValue > 0) != (farValue > 0) {
			return bisectPlanetTransitLocal(near, nearValue, far, metric)
		}
		if far == greatestTT {
			return 0, false
		}
		near, nearValue = far, farValue
	}
	return 0, false
}

func bisectPlanetTransitLocal(left, leftValue, right float64, metric func(float64) float64) (float64, bool) {
	for iteration := 0; iteration < 80 && math.Abs(right-left) > planetTransitLocalToleranceDays; iteration++ {
		middle := (left + right) / 2
		value := metric(middle)
		if !isFiniteFloat(value) {
			return 0, false
		}
		if (value > 0) == (leftValue > 0) {
			left, leftValue = middle, value
		} else {
			right = middle
		}
	}
	return (left + right) / 2, true
}

// localSunEvents 判定一触到四触之间太阳中心是否在视地平上，并找出其间第一次日出、日落。
func (snapshot *planetTransitSnapshot) localSunEvents(result *PlanetTransitLocalResult, startTT, endTT float64, observer Observer) {
	altitude := func(tt float64) float64 {
		state, ok := snapshot.stateAt(tt, observer)
		if !ok {
			return math.NaN()
		}
		return state.sunAltitudeDeg - planetTransitSunHorizonAltitudeDeg
	}
	count := int(math.Ceil((endTT-startTT)/planetTransitLocalSunStepDays)) + 1
	previousTT, previous := startTT, altitude(startTT)
	result.Visible = previous > 0
	for index := 1; index < count; index++ {
		tt := math.Min(endTT, startTT+float64(index)*planetTransitLocalSunStepDays)
		value := altitude(tt)
		if !isFiniteFloat(value) || !isFiniteFloat(previous) {
			previousTT, previous = tt, value
			continue
		}
		if value > 0 {
			result.Visible = true
		}
		if (previous > 0) != (value > 0) {
			crossing, ok := bisectPlanetTransitLocal(previousTT, previous, tt, altitude)
			if ok && value > 0 && result.Sunrise == 0 {
				result.Sunrise = TT2UTC(crossing)
			} else if ok && value <= 0 && result.Sunset == 0 {
				result.Sunset = TT2UTC(crossing)
			}
		}
		previousTT, previous = tt, value
	}
}

const (
	// planetTransitIsochroneGridDegrees 是等时线网格间距。站心接触时刻是观测者地心矢量的近线性函数，
	// 2° 网格的线性插值误差在地图上不到一像素。
	planetTransitIsochroneGridDegrees = 2.0
	// planetTransitIsochroneMaxLevels 是每个接触的等时线条数上限；超出时步长自动加倍。
	planetTransitIsochroneMaxLevels = 64
)

// PlanetTransitIsochronePoint 是等时线上的一点 / one point on an isochrone.
type PlanetTransitIsochronePoint struct {
	// Longitude 东经为正，Latitude 北纬为正，单位度 / east-positive longitude and north-positive latitude in degrees.
	Longitude float64
	Latitude  float64
}

// PlanetTransitIsochrone 是同一站心接触时刻的地点连线，只保留该时刻太阳中心在视地平上（计入标准大气折射）的部分。
// PlanetTransitIsochrone joins the sites that see one contact at the same instant; only the part where the Sun's centre is above the apparent horizon (standard refraction) is kept.
type PlanetTransitIsochrone struct {
	Kind PlanetTransitContactKind
	// JD 是该等时线表示的 UTC 儒略日 / JD is the UTC Julian day represented by the line.
	JD float64
	// Segments 是连续支路；一条支路止于地平线或网格边缘 / continuous branches ending at the horizon or the grid edge.
	Segments [][]PlanetTransitIsochronePoint
}

// PlanetTransitContactIsochrones 返回接触 kind 的站心等时线，时刻按 step 对齐到整刻度（UTC）；
// 条数超过 64 时步长按倍数放大，保证覆盖整个接触时段；
// step 非正、kind 不是 C1–C4 或地心凌日没有该接触时返回 nil。
// 时刻是站心接触时刻，受视差影响可落在地心 C1–C4 之外（金星约数分钟）；
// 一触、四触与远离掠凌边界的二触、三触，线上地点的站心接触时刻与 PlanetTransitLocal 相差不到 1 秒；
// 掠凌边界附近（站心内切余量 <0.5″）内切时刻梯度发散，偏差可达数十秒。
// PlanetTransitContactIsochrones returns topocentric isochrones of contact kind at UTC instants aligned to step.
// When more than 64 lines would be needed, the step is doubled until the whole contact span fits.
// It returns nil for a non-positive step, a kind other than C1-C4, or a contact the geocentric transit lacks.
// Points agree with PlanetTransitLocal to within a second, except C2/C3 near the grazing limit (internal margin
// below 0.5″), where the contact-time gradient diverges and errors reach tens of seconds.
func PlanetTransitContactIsochrones(transit PlanetTransitResult, kind PlanetTransitContactKind, step time.Duration) []PlanetTransitIsochrone {
	return PlanetTransitContactIsochronesWithInstance(transit, kind, step, nil)
}

// PlanetTransitContactIsochronesWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func PlanetTransitContactIsochronesWithInstance(
	transit PlanetTransitResult,
	kind PlanetTransitContactKind,
	step time.Duration,
	inst *Instance,
) []PlanetTransitIsochrone {
	if step <= 0 || kind < PlanetTransitC1 || kind > PlanetTransitC4 {
		return nil
	}
	if (kind == PlanetTransitC2 || kind == PlanetTransitC3) && !transit.HasInternal {
		return nil
	}
	snapshot, ok := newPlanetTransitSnapshot(transit, inst)
	if !ok {
		return nil
	}
	grid := snapshot.contactGrid(transit, kind)
	minimum, maximum := math.Inf(1), math.Inf(-1)
	for _, row := range grid.times {
		for _, value := range row {
			if isFiniteFloat(value) {
				minimum, maximum = math.Min(minimum, value), math.Max(maximum, value)
			}
		}
	}
	if !(maximum > minimum) {
		return nil
	}
	// 条数超过上限时步长加倍，保证等时线覆盖整个接触时段，而不是截掉后半段。
	levels := greatestTimeContourAlignedLevels(minimum, maximum, step, planetTransitIsochroneMaxLevels+1)
	for len(levels) > planetTransitIsochroneMaxLevels {
		step *= 2
		levels = greatestTimeContourAlignedLevels(minimum, maximum, step, planetTransitIsochroneMaxLevels+1)
	}
	var result []PlanetTransitIsochrone
	for _, level := range levels {
		segments := grid.contour(level.tt)
		segments = snapshot.clipToDaylight(segments, level.tt)
		if len(segments) == 0 {
			continue
		}
		result = append(result, PlanetTransitIsochrone{Kind: kind, JD: Date2JD(level.at), Segments: segments})
	}
	return result
}

// planetTransitContactGrid 是全球网格上的站心接触时刻（TT），times[row][column]，无解处为 NaN。
// 行从南极到北极，列从 -180° 起向东，经度方向首尾相接。
type planetTransitContactGrid struct {
	times [][]float64
}

func (grid planetTransitContactGrid) point(row, column int) PlanetTransitIsochronePoint {
	return PlanetTransitIsochronePoint{
		Longitude: -180 + float64(column)*planetTransitIsochroneGridDegrees,
		Latitude:  -90 + float64(row)*planetTransitIsochroneGridDegrees,
	}
}

func (snapshot *planetTransitSnapshot) contactGrid(transit PlanetTransitResult, kind PlanetTransitContactKind) planetTransitContactGrid {
	rows := int(math.Round(180/planetTransitIsochroneGridDegrees)) + 1
	columns := int(math.Round(360 / planetTransitIsochroneGridDegrees))
	greatestTT := UTC2TT(transit.Greatest)
	seedTT := planetTransitGeocentricContactTT(transit, kind, greatestTT)
	grid := planetTransitContactGrid{times: make([][]float64, rows)}
	for row := range grid.times {
		grid.times[row] = make([]float64, columns)
		for column := range grid.times[row] {
			point := grid.point(row, column)
			observer := Observer{Longitude: point.Longitude, Latitude: point.Latitude}
			grid.times[row][column] = snapshot.fastContactTT(kind, seedTT, greatestTT, observer)
		}
	}
	return grid
}

// fastContactTT 先用割线法从地心接触出发求站心接触：接触量对时间近似线性，通常四五次就收敛；
// 不收敛（近掠凌）时退回分档求根。
func (snapshot *planetTransitSnapshot) fastContactTT(kind PlanetTransitContactKind, seedTT, greatestTT float64, observer Observer) float64 {
	internal := kind == PlanetTransitC2 || kind == PlanetTransitC3
	metric := func(tt float64) float64 {
		state, ok := snapshot.stateAt(tt, observer)
		if !ok {
			return math.NaN()
		}
		if internal {
			return state.internal
		}
		return state.external
	}
	x0, x1 := seedTT, seedTT+1.0/1440
	f0, f1 := metric(x0), metric(x1)
	for iteration := 0; iteration < 12 && isFiniteFloat(f0) && isFiniteFloat(f1) && f1 != f0; iteration++ {
		x2 := x1 - f1*(x1-x0)/(f1-f0)
		if !isFiniteFloat(x2) || math.Abs(x2-seedTT) > planetTransitLocalWindowDays {
			break
		}
		if math.Abs(x2-x1) <= planetTransitLocalToleranceDays {
			// 割线根可能落在另一侧接触上：入凌接触必须早于凌甚，出凌接触必须晚于凌甚。
			if (kind == PlanetTransitC1 || kind == PlanetTransitC2) == (x2 < greatestTT) {
				return x2
			}
			break
		}
		x0, f0, x1, f1 = x1, f1, x2, metric(x2)
	}
	localGreatestTT, ok := snapshot.localGreatestTT(greatestTT, observer)
	if !ok {
		return math.NaN()
	}
	if tt, ok := snapshot.localContactTT(kind, seedTT, localGreatestTT, observer); ok {
		return tt
	}
	return math.NaN()
}

// contour 用 marching squares 求 level 等值线，跨单元的交点按网格边编号拼接成折线；
// 经度方向首尾相接，输出经度规范到 [-180°,180°)，跨日界线的拆分交给绘图层。
func (grid planetTransitContactGrid) contour(level float64) [][]PlanetTransitIsochronePoint {
	rows := len(grid.times)
	if rows < 2 {
		return nil
	}
	columns := len(grid.times[0])
	// 边编号：水平边 (row,column)-(row,column+1) 为 2k，竖直边 (row,column)-(row+1,column) 为 2k+1，k=row*columns+column。
	edgeID := func(row, column int, vertical bool) int {
		id := (row*columns + column) * 2
		if vertical {
			id++
		}
		return id
	}
	points := map[int]PlanetTransitIsochronePoint{}
	edgePoint := func(row, column int, vertical bool) int {
		id := edgeID(row, column, vertical)
		if _, ok := points[id]; ok {
			return id
		}
		nextRow, nextColumn := row, (column+1)%columns
		if vertical {
			nextRow, nextColumn = row+1, column
		}
		a, b := grid.times[row][column], grid.times[nextRow][nextColumn]
		fraction := (level - a) / (b - a)
		start := grid.point(row, column)
		if vertical {
			start.Latitude += fraction * planetTransitIsochroneGridDegrees
		} else {
			start.Longitude += fraction * planetTransitIsochroneGridDegrees
		}
		start.Longitude = normalizeLongitude180(start.Longitude)
		points[id] = start
		return id
	}
	links := map[int][]int{}
	link := func(a, b int) {
		links[a] = append(links[a], b)
		links[b] = append(links[b], a)
	}
	for row := 0; row < rows-1; row++ {
		for column := 0; column < columns; column++ {
			next := (column + 1) % columns
			corners := [4]float64{grid.times[row][column], grid.times[row][next], grid.times[row+1][next], grid.times[row+1][column]}
			mask, valid := 0, true
			for index, value := range corners {
				if !isFiniteFloat(value) {
					valid = false
					break
				}
				if value > level {
					mask |= 1 << index
				}
			}
			if !valid || mask == 0 || mask == 15 {
				continue
			}
			// 四条边：0 下（水平 row）、1 右（竖直 next）、2 上（水平 row+1）、3 左（竖直 column）。
			edge := func(side int) int {
				switch side {
				case 0:
					return edgePoint(row, column, false)
				case 1:
					return edgePoint(row, next, true)
				case 2:
					return edgePoint(row+1, column, false)
				}
				return edgePoint(row, column, true)
			}
			crossed := make([]int, 0, 4)
			for side := 0; side < 4; side++ {
				if (mask>>side)&1 != (mask>>((side+1)%4))&1 {
					crossed = append(crossed, side)
				}
			}
			if len(crossed) == 2 {
				link(edge(crossed[0]), edge(crossed[1]))
				continue
			}
			// 鞍点：按中心均值决定连法，中心与角 0 同侧时把角 0 所在的两条边分开连。
			center := (corners[0] + corners[1] + corners[2] + corners[3]) / 4
			if (center > level) == (mask&1 != 0) {
				link(edge(0), edge(1))
				link(edge(2), edge(3))
			} else {
				link(edge(3), edge(0))
				link(edge(1), edge(2))
			}
		}
	}
	return chainPlanetTransitContour(points, links)
}

// chainPlanetTransitContour 把边交点按邻接关系串成折线：先从端点（度为 1）出发，剩下的是闭环。
func chainPlanetTransitContour(points map[int]PlanetTransitIsochronePoint, links map[int][]int) [][]PlanetTransitIsochronePoint {
	ids := make([]int, 0, len(links))
	for id := range links {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	used := map[[2]int]bool{}
	key := func(a, b int) [2]int {
		if a > b {
			a, b = b, a
		}
		return [2]int{a, b}
	}
	walk := func(start int) []PlanetTransitIsochronePoint {
		line := []PlanetTransitIsochronePoint{points[start]}
		current := start
		for {
			next := -1
			for _, candidate := range links[current] {
				if !used[key(current, candidate)] {
					next = candidate
					break
				}
			}
			if next < 0 {
				return line
			}
			used[key(current, next)] = true
			line = append(line, points[next])
			current = next
		}
	}
	var result [][]PlanetTransitIsochronePoint
	for pass := 0; pass < 2; pass++ {
		for _, id := range ids {
			if pass == 0 && len(links[id]) != 1 {
				continue
			}
			if line := walk(id); len(line) >= 2 {
				result = append(result, line)
			}
		}
	}
	return result
}

func normalizeLongitude180(longitude float64) float64 {
	longitude = math.Mod(longitude+180, 360)
	if longitude < 0 {
		longitude += 360
	}
	return longitude - 180
}

// clipToDaylight 只保留 tt 时刻太阳中心在视地平上的部分，切点沿折线段二分求出。
func (snapshot *planetTransitSnapshot) clipToDaylight(segments [][]PlanetTransitIsochronePoint, tt float64) [][]PlanetTransitIsochronePoint {
	values, ok := snapshot.interpolate(tt)
	if !ok {
		return nil
	}
	sidereal := values[planetTransitSnapshotSidereal]
	altitude := func(point PlanetTransitIsochronePoint) float64 {
		observer := Observer{Longitude: point.Longitude, Latitude: point.Latitude}
		ra, dec := topocentricRaDecWithSidereal(values[planetTransitSnapshotSunRA], values[planetTransitSnapshotSunDec],
			observer.Latitude, observer.Longitude, sidereal, values[planetTransitSnapshotSunDistance], 0)
		return occultationAltitudeWithSidereal(sidereal, observer, ra, dec) - planetTransitSunHorizonAltitudeDeg
	}
	between := func(a, b PlanetTransitIsochronePoint, fraction float64) PlanetTransitIsochronePoint {
		delta := normalizeLongitude180(b.Longitude - a.Longitude)
		return PlanetTransitIsochronePoint{
			Longitude: normalizeLongitude180(a.Longitude + fraction*delta),
			Latitude:  a.Latitude + fraction*(b.Latitude-a.Latitude),
		}
	}
	crossing := func(a, b PlanetTransitIsochronePoint, aUp bool) PlanetTransitIsochronePoint {
		left, right := 0.0, 1.0
		for iteration := 0; iteration < 30; iteration++ {
			middle := (left + right) / 2
			if (altitude(between(a, b, middle)) > 0) == aUp {
				left = middle
			} else {
				right = middle
			}
		}
		return between(a, b, (left+right)/2)
	}
	var result [][]PlanetTransitIsochronePoint
	for _, segment := range segments {
		var current []PlanetTransitIsochronePoint
		previousUp := false
		for index, point := range segment {
			up := altitude(point) > 0
			if index > 0 && up != previousUp {
				cut := crossing(segment[index-1], point, previousUp)
				current = append(current, cut)
				if previousUp {
					if len(current) >= 2 {
						result = append(result, current)
					}
					current = nil
				}
			}
			if up {
				current = append(current, point)
			}
			previousUp = up
		}
		if len(current) >= 2 {
			result = append(result, current)
		}
	}
	return result
}

// PlanetTransitLocalContactInfo 是以 time.Time 表示的站心凌日接触。
// PlanetTransitLocalContactInfo is a topocentric transit contact expressed with time.Time.
type PlanetTransitLocalContactInfo struct {
	// Time 为零值表示该接触不存在 / a zero Time means the contact does not exist.
	Time time.Time
	// SunAltitude 与 SunAzimuth 是太阳中心几何高度与方位（北起向东），不含折射，单位度。
	// SunAltitude and SunAzimuth are the geometric altitude and azimuth (north through east) of the Sun's centre, without refraction.
	SunAltitude float64
	SunAzimuth  float64
	// PositionAngle 是行星中心相对日心的位置角，北起向东，单位度。
	// PositionAngle is the position angle of the planet centre from the solar centre, north through east.
	PositionAngle float64
}

// PlanetTransitLocalInfo 是以 time.Time 表示的站心凌日，供水星、金星包共用。
// PlanetTransitLocalInfo is the topocentric transit expressed with time.Time, shared by the Mercury and Venus packages.
type PlanetTransitLocalInfo struct {
	// Valid 为 false 表示该地点几何上看不到凌日，其余字段为零值。
	// Valid is false when the site sees no geometric transit; other fields are then zero.
	Valid bool
	// Start、InternalStart、Greatest、InternalEnd、End 是一触、二触、凌甚、三触、四触。
	// Start, InternalStart, Greatest, InternalEnd and End are C1, C2, greatest, C3 and C4.
	Start         PlanetTransitLocalContactInfo
	InternalStart PlanetTransitLocalContactInfo
	Greatest      PlanetTransitLocalContactInfo
	InternalEnd   PlanetTransitLocalContactInfo
	End           PlanetTransitLocalContactInfo
	HasInternal   bool
	// 凌甚时的站心角距与视半径，单位角秒；太阳视半径为站心值，行星视半径沿用地心值（站心差约 0.006″，忽略）。
	// Separation and semidiameters at greatest, in arcseconds: separation and solar semidiameter are topocentric,
	// the planet semidiameter keeps the geocentric value (topocentric difference ≈0.006″, neglected).
	MinimumSeparationArcsec  float64
	SunSemidiameterArcsec    float64
	PlanetSemidiameterArcsec float64
	// Visible 表示一触到四触之间太阳中心至少有一刻在视地平上（计入标准大气折射）。
	// Visible reports whether the Sun's centre is above the apparent horizon (standard refraction) at some instant between C1 and C4.
	Visible bool
	// Sunrise 与 Sunset 是凌日期间太阳中心第一次升出、落入视地平；没有时为零值。
	// Sunrise and Sunset are the first sunrise and sunset during the transit, or zero when absent.
	Sunrise time.Time
	Sunset  time.Time
}

// PlanetTransitLocalInfoFromResult 把站心凌日结果换成 loc 时区的 time.Time 表示。
// PlanetTransitLocalInfoFromResult converts a topocentric transit result to time.Time values in loc.
func PlanetTransitLocalInfoFromResult(result PlanetTransitLocalResult, loc *time.Location) PlanetTransitLocalInfo {
	if !result.Valid {
		return PlanetTransitLocalInfo{}
	}
	if loc == nil {
		loc = time.UTC
	}
	at := func(jd float64) time.Time {
		if jd == 0 {
			return time.Time{}
		}
		return JD2DateByZone(jd, loc, false)
	}
	contact := func(value PlanetTransitLocalContact) PlanetTransitLocalContactInfo {
		if !value.Valid {
			return PlanetTransitLocalContactInfo{}
		}
		return PlanetTransitLocalContactInfo{
			Time:          at(value.JD),
			SunAltitude:   value.SunAltitudeDeg,
			SunAzimuth:    value.SunAzimuthDeg,
			PositionAngle: value.PositionAngleDeg,
		}
	}
	return PlanetTransitLocalInfo{
		Valid:                    true,
		Start:                    contact(result.ExternalIngress),
		InternalStart:            contact(result.InternalIngress),
		Greatest:                 contact(result.Greatest),
		InternalEnd:              contact(result.InternalEgress),
		End:                      contact(result.ExternalEgress),
		HasInternal:              result.HasInternal,
		MinimumSeparationArcsec:  result.MinimumSeparationArcsec,
		SunSemidiameterArcsec:    result.SunSemidiameterArcsec,
		PlanetSemidiameterArcsec: result.PlanetSemidiameterArcsec,
		Visible:                  result.Visible,
		Sunrise:                  at(result.Sunrise),
		Sunset:                   at(result.Sunset),
	}
}

// PlanetTransitResultFromTimes 由各接触时刻重建地心凌日结果，零值时刻表示该接触不存在。
// 行星包的 TransitInfo 只保存 time.Time，站心计算用它还原 basic 结果。
// PlanetTransitResultFromTimes rebuilds a geocentric transit result from contact times; a zero time means the contact is absent.
func PlanetTransitResultFromTimes(planetIndex int, start, internalStart, greatest, internalEnd, end time.Time) PlanetTransitResult {
	jd := func(value time.Time) float64 {
		if value.IsZero() {
			return 0
		}
		return Date2JD(value.UTC())
	}
	result := PlanetTransitResult{
		Valid:           !start.IsZero() && !greatest.IsZero() && !end.IsZero(),
		PlanetIndex:     planetIndex,
		ExternalIngress: jd(start),
		InternalIngress: jd(internalStart),
		Greatest:        jd(greatest),
		InternalEgress:  jd(internalEnd),
		ExternalEgress:  jd(end),
		HasExternal:     true,
	}
	result.HasInternal = result.InternalIngress != 0 && result.InternalEgress != 0
	return result
}

// PlanetTransitLocalFrame 是站心日面上一个时刻的行星位置。
// PlanetTransitLocalFrame is the topocentric planet position on the solar disk at one instant.
type PlanetTransitLocalFrame struct {
	// JD 是 UTC 儒略日 / JD is the UTC Julian day.
	JD float64
	// PlanetXArcsec 与 PlanetYArcsec 是行星中心相对日心的切平面偏移，X 向东、Y 向北为正，单位角秒。
	// PlanetXArcsec and PlanetYArcsec are tangent-plane offsets from the solar centre; X is east and Y is north.
	PlanetXArcsec float64
	PlanetYArcsec float64
	// SunRadiusArcsec 与 PlanetRadiusArcsec 是视半径 / apparent semidiameters.
	SunRadiusArcsec    float64
	PlanetRadiusArcsec float64
	// SunAltitudeDeg 是太阳中心几何高度 / geometric altitude of the Sun's centre.
	SunAltitudeDeg float64
}

const (
	// planetTransitLocalMaxFrames 是 PlanetTransitLocalFrames 返回帧数上限（含事件帧）。
	planetTransitLocalMaxFrames = 2000
	// planetTransitLocalFrameEpsilonDays 内的相邻时刻视为同一帧，约 0.01 秒。
	planetTransitLocalFrameEpsilonDays = 1e-7
)

// PlanetTransitLocalFrames 在站心一触到四触之间按 stepDays 采样行星日面位置，首尾与各接触、凌甚时刻都保留；
// stepDays 非正时取两分钟；返回帧数（含接触与凌甚帧）上限 2000，步长过细时自动放宽。
// PlanetTransitLocalFrames samples the planet position on the solar disk between topocentric C1 and C4 every stepDays,
// always including the contacts and greatest. Non-positive stepDays uses two minutes; at most 2000 frames are returned, contacts and greatest included, widening the step when needed.
func PlanetTransitLocalFrames(transit PlanetTransitResult, local PlanetTransitLocalResult, stepDays float64) []PlanetTransitLocalFrame {
	return PlanetTransitLocalFramesWithInstance(transit, local, stepDays, nil)
}

// PlanetTransitLocalFramesWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func PlanetTransitLocalFramesWithInstance(
	transit PlanetTransitResult,
	local PlanetTransitLocalResult,
	stepDays float64,
	inst *Instance,
) []PlanetTransitLocalFrame {
	if !local.Valid {
		return nil
	}
	snapshot, ok := newPlanetTransitSnapshot(transit, inst)
	if !ok {
		return nil
	}
	if !(stepDays > 0) || !isFiniteFloat(stepDays) {
		stepDays = 2.0 / 1440
	}
	startTT, endTT := UTC2TT(local.ExternalIngress.JD), UTC2TT(local.ExternalEgress.JD)
	events := []float64{}
	for _, contact := range []PlanetTransitLocalContact{local.InternalIngress, local.Greatest, local.InternalEgress, local.ExternalEgress} {
		if contact.Valid {
			events = append(events, UTC2TT(contact.JD))
		}
	}
	// 事件帧计入 2000 帧预算：常规采样（含一触）最多 2000-len(events) 点。
	maxRegular := float64(planetTransitLocalMaxFrames - len(events))
	if span := endTT - startTT; span/stepDays > maxRegular {
		stepDays = span / maxRegular
	}
	times := make([]float64, 0, planetTransitLocalMaxFrames)
	for index := 0; ; index++ {
		tt := startTT + float64(index)*stepDays
		if tt >= endTT || float64(index) >= maxRegular {
			break
		}
		times = append(times, tt)
	}
	times = append(times, events...)
	sort.Float64s(times)
	// 去掉与事件重合的常规采样，避免重复帧。
	deduped := times[:0]
	for _, tt := range times {
		if len(deduped) > 0 && tt-deduped[len(deduped)-1] < planetTransitLocalFrameEpsilonDays {
			continue
		}
		deduped = append(deduped, tt)
	}
	times = deduped
	frames := make([]PlanetTransitLocalFrame, 0, len(times))
	for _, tt := range times {
		state, ok := snapshot.stateAt(tt, local.Observer)
		if !ok {
			return nil
		}
		angle := state.positionAngleDeg * math.Pi / 180
		frames = append(frames, PlanetTransitLocalFrame{
			JD:                 TT2UTC(tt),
			PlanetXArcsec:      state.separationArcsec * math.Sin(angle),
			PlanetYArcsec:      state.separationArcsec * math.Cos(angle),
			SunRadiusArcsec:    state.sunRadiusArcsec,
			PlanetRadiusArcsec: state.planetRadius,
			SunAltitudeDeg:     state.sunAltitudeDeg,
		})
	}
	return frames
}
