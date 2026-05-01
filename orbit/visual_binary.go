package orbit

import (
	"math"
	"time"
)

const visualBinaryDeg = 180 / math.Pi
const visualBinaryRad = math.Pi / 180

// VisualBinaryElements 视双星轨道要素，采用《天文算法》第 55 章的经典口径。
type VisualBinaryElements struct {
	PeriodYears        float64 // 周期 P，单位平太阳年 / orbital period in mean solar years.
	PeriastronYear     float64 // 过近星点时刻 T，采用带小数的年 / epoch of periastron as a decimal year.
	Eccentricity       float64 // 离心率 e / eccentricity.
	SemiMajorAxis      float64 // 半长轴 a，单位角秒 / semi-major axis in arcseconds.
	Inclination        float64 // 倾角 i，单位度 / inclination in degrees.
	AscendingNode      float64 // 升交点位置角 Ω，单位度 / position angle of ascending node in degrees.
	PeriastronArgument float64 // 近星点角距 ω，单位度 / argument of periastron in degrees.
}

// VisualBinaryPosition 视双星在天球上的计算结果。
type VisualBinaryPosition struct {
	Year             float64 // 计算使用的小数年 / decimal year used for the evaluation.
	MeanAnomaly      float64 // 平近点角 M，单位度 / mean anomaly in degrees.
	EccentricAnomaly float64 // 偏近点角 E，单位度 / eccentric anomaly in degrees.
	TrueAnomaly      float64 // 真近点角 v，单位度 / true anomaly in degrees.
	Radius           float64 // 径矢 r，单位角秒 / radius vector in arcseconds.
	PositionAngle    float64 // 位置角 θ，北为 0°、东为 90° / position angle, north through east.
	Separation       float64 // 角距离 ρ，单位角秒 / apparent separation in arcseconds.
}

// VisualBinary 视双星位置 / visual binary position.
//
// 输入时刻会先换算为 UTC 小数年，再按经典视轨道公式求解。
// The input instant is converted to a UTC decimal year before evaluation.
func VisualBinary(date time.Time, elements VisualBinaryElements) VisualBinaryPosition {
	return VisualBinaryByYear(decimalYearUTC(date), elements)
}

// VisualBinaryByYear 视双星位置（按小数年） / visual binary position by decimal year.
//
// 返回给定小数年对应的视双星位置角和角距离。
func VisualBinaryByYear(year float64, elements VisualBinaryElements) VisualBinaryPosition {
	if !validVisualBinaryElements(year, elements) {
		return invalidVisualBinaryPosition(year)
	}

	meanAnomaly := normalize360Local(360 / elements.PeriodYears * (year - elements.PeriastronYear))
	eccentricAnomaly, ok := solveVisualBinaryEccentricAnomaly(meanAnomaly*visualBinaryRad, elements.Eccentricity)
	if !ok {
		return invalidVisualBinaryPosition(year)
	}

	sinE, cosE := math.Sincos(eccentricAnomaly)
	radius := elements.SemiMajorAxis * (1 - elements.Eccentricity*cosE)
	trueAnomaly := math.Atan2(
		math.Sqrt(1-elements.Eccentricity*elements.Eccentricity)*sinE,
		cosE-elements.Eccentricity,
	) * visualBinaryDeg

	u := (trueAnomaly + elements.PeriastronArgument) * visualBinaryRad
	sinU, cosU := math.Sincos(u)
	cosI := math.Cos(elements.Inclination * visualBinaryRad)

	thetaMinusNode := math.Atan2(sinU*cosI, cosU) * visualBinaryDeg
	positionAngle := normalize360Local(thetaMinusNode + elements.AscendingNode)
	separation := radius * math.Hypot(cosU, sinU*cosI)

	return VisualBinaryPosition{
		Year:             year,
		MeanAnomaly:      meanAnomaly,
		EccentricAnomaly: normalize360Local(eccentricAnomaly * visualBinaryDeg),
		TrueAnomaly:      normalize360Local(trueAnomaly),
		Radius:           radius,
		PositionAngle:    positionAngle,
		Separation:       separation,
	}
}

func decimalYearUTC(date time.Time) float64 {
	date = date.UTC()
	year := date.Year()
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	return float64(year) + float64(date.Sub(start))/float64(end.Sub(start))
}

func validVisualBinaryElements(year float64, elements VisualBinaryElements) bool {
	if !isFiniteLocal(year) ||
		!isFiniteLocal(elements.PeriodYears) ||
		!isFiniteLocal(elements.PeriastronYear) ||
		!isFiniteLocal(elements.Eccentricity) ||
		!isFiniteLocal(elements.SemiMajorAxis) ||
		!isFiniteLocal(elements.Inclination) ||
		!isFiniteLocal(elements.AscendingNode) ||
		!isFiniteLocal(elements.PeriastronArgument) {
		return false
	}
	return elements.PeriodYears > 0 &&
		elements.SemiMajorAxis > 0 &&
		elements.Eccentricity >= 0 &&
		elements.Eccentricity < 1
}

func invalidVisualBinaryPosition(year float64) VisualBinaryPosition {
	nan := math.NaN()
	return VisualBinaryPosition{
		Year:             year,
		MeanAnomaly:      nan,
		EccentricAnomaly: nan,
		TrueAnomaly:      nan,
		Radius:           nan,
		PositionAngle:    nan,
		Separation:       nan,
	}
}

func solveVisualBinaryEccentricAnomaly(meanAnomalyRad, eccentricity float64) (float64, bool) {
	if !isFiniteLocal(meanAnomalyRad) || !isFiniteLocal(eccentricity) || eccentricity < 0 || eccentricity >= 1 {
		return math.NaN(), false
	}
	if meanAnomalyRad > math.Pi {
		meanAnomalyRad -= 2 * math.Pi
	} else if meanAnomalyRad < -math.Pi {
		meanAnomalyRad += 2 * math.Pi
	}

	eccentricAnomaly := meanAnomalyRad
	if eccentricity >= 0.8 {
		eccentricAnomaly = math.Pi
		if meanAnomalyRad < 0 {
			eccentricAnomaly = -math.Pi
		}
	}

	for i := 0; i < 32; i++ {
		sinE, cosE := math.Sincos(eccentricAnomaly)
		delta := (eccentricAnomaly - eccentricity*sinE - meanAnomalyRad) / (1 - eccentricity*cosE)
		eccentricAnomaly -= delta
		if math.Abs(delta) < 1e-14 {
			return eccentricAnomaly, true
		}
	}
	return eccentricAnomaly, true
}

func normalize360Local(value float64) float64 {
	value = math.Mod(value, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func isFiniteLocal(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
