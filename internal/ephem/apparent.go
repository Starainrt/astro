package ephem

// ApparentEngine 支持高精度视位置的星历引擎；Apparent 返回 ICRF 视方向单位矢量（含光行时、太阳引力偏折、相对论光行差）、
// 几何地心距 km 与光行时（日） / ephemeris engine with apparent positions.
type ApparentEngine interface {
	Engine
	Apparent(target Body, jdTT float64) (dir [3]float64, distKM, tauDays float64, err error)
}
