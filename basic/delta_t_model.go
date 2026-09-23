package basic

import "math"

// DeltaTModel 标识一个可选的已发布 ΔT 模型 / a selectable published ΔT model.
type DeltaTModel string

const (
	// DeltaTModelDefault 内置默认模型：Stephenson–Morrison–Hohenkerk 2016 + Morrison 2021 的样条与积分 lod。
	// DeltaTModelDefault is the built-in model: the SMH2016 + Morrison 2021 spline and integrated lod.
	DeltaTModelDefault DeltaTModel = ""
	// DeltaTModelSMH2016 显式选择内置的 SMH2016 + Morrison 2021 模型。
	// DeltaTModelSMH2016 selects the built-in SMH2016 + Morrison 2021 model explicitly.
	DeltaTModelSMH2016 DeltaTModel = "smh2016"
	// DeltaTModelMS2004 Morrison & Stephenson 2004 的长期抛物线 ΔT = −20 + 32u²（u = (y−1820)/100），单位秒。
	// DeltaTModelMS2004 is the Morrison & Stephenson 2004 long-term parabola in seconds.
	DeltaTModelMS2004 DeltaTModel = "ms2004"
	// DeltaTModelEspenakMeeus2006 Espenak & Meeus 2006 的分段多项式，基于 M&S2004，未加配对修正 c。
	// DeltaTModelEspenakMeeus2006 is the Espenak & Meeus 2006 piecewise polynomial without the pairing term c.
	DeltaTModelEspenakMeeus2006 DeltaTModel = "espenak_meeus_2006"
	// DeltaTModelNASACanon2006 是 Espenak–Meeus 再加配对修正 c = −0.000012932(y−1955)²（y < 1955），
	// 即 NASA 五千年目录实际印刷的 ΔT 口径，配对的是 ELP2000/82 的 −25.858″/cy²。
	// DeltaTModelNASACanon2006 adds the pairing term c for y < 1955, the convention printed by NASA's canon.
	DeltaTModelNASACanon2006 DeltaTModel = "nasa_canon_2006"
	// DeltaTModelManual 表示当前生效的是 SetDeltaTFn 注入的任意函数，不是命名模型。
	// DeltaTModelManual reports that an arbitrary function injected with SetDeltaTFn is active.
	DeltaTModelManual DeltaTModel = "manual"
)

// deltaTModelFunction 把命名模型包成 ΔT 钩子的签名；isJulianDay 为假时入参是十进制年。
func deltaTModelFunction(model DeltaTModel, keepObserved bool) func(float64, bool) float64 {
	return func(date float64, isJulianDay bool) float64 {
		if math.IsNaN(date) || math.IsInf(date, 0) {
			return math.NaN()
		}
		if isJulianDay {
			return DeltaTModelSeconds(model, date, keepObserved)
		}
		if keepObserved {
			year := math.Floor(date)
			start := JDCalc(int(year), 1, 1)
			end := JDCalc(int(year)+1, 1, 1)
			jd := start + (date-year)*(end-start)
			if seconds, ok := deltaTMonthlyAt(jd); ok {
				return seconds
			}
		}
		return deltaTModelSecondsAtYear(model, date)
	}
}

// DeltaTModelSeconds 按命名模型求 ΔT（秒），入参为 UT 儒略日。
//
// keepObserved 为真时，落在逐月实测表覆盖段内的时刻一律返回实测值，模型只接管表外；
// 这是"实测优先"的推荐用法，也是内置默认模型的结构。为假时全程使用模型，仅用于口径对照。
// DeltaTModelSeconds evaluates a named model. With keepObserved the monthly observed table wins
// wherever it covers the instant and the model only takes over outside it.
func DeltaTModelSeconds(model DeltaTModel, jd float64, keepObserved bool) float64 {
	if math.IsNaN(jd) || math.IsInf(jd, 0) {
		return math.NaN()
	}
	switch model {
	case DeltaTModelDefault, DeltaTModelSMH2016:
		if keepObserved {
			// 与内置默认逐位一致（含实测段与样条表尾的常值锚定）。
			return deltaTModelSecondsAtUT(jd)
		}
		return deltaTSplineAtJDE(jd)
	}
	if keepObserved {
		if seconds, ok := deltaTMonthlyAt(jd); ok {
			return seconds
		}
	}
	return deltaTModelSecondsAtYear(model, deltaTYearAtJDE(jd))
}

// SetDeltaTModel 安装命名模型；keepObserved 为真时保留实测表覆盖段（推荐），未知模型返回 false 且不改动现状。
// SetDeltaTModel installs a named model, keeping the observed span when keepObserved is true;
// an unknown model returns false and leaves the current model untouched.
func SetDeltaTModel(model DeltaTModel, keepObserved bool) bool {
	if !knownDeltaTModel(model) {
		return false
	}
	if model == DeltaTModelDefault || model == DeltaTModelSMH2016 {
		if keepObserved {
			// 内置模型本身就是"实测优先"，直接用它可以避免表尾锚定出现差异。
			installDeltaTFn(DefaultDeltaTv2, model, true)
			return true
		}
		installDeltaTFn(deltaTModelFunction(model, false), model, false)
		return true
	}
	installDeltaTFn(deltaTModelFunction(model, keepObserved), model, keepObserved)
	return true
}

// GetDeltaTModel 返回当前生效的模型与是否保留实测段；SetDeltaTFn 注入的任意函数报 DeltaTModelManual。
// GetDeltaTModel returns the active model and whether the observed span is kept; an arbitrary
// function injected with SetDeltaTFn reports DeltaTModelManual.
func GetDeltaTModel() (DeltaTModel, bool) {
	deltaTFnMu.RLock()
	defer deltaTFnMu.RUnlock()
	return activeDeltaTModel, activeDeltaTKeepObserved
}

func knownDeltaTModel(model DeltaTModel) bool {
	switch model {
	case DeltaTModelDefault, DeltaTModelSMH2016, DeltaTModelMS2004,
		DeltaTModelEspenakMeeus2006, DeltaTModelNASACanon2006:
		return true
	}
	return false
}

// deltaTModelSecondsAtYear 按十进制年求模型值；实测段是否介入由调用方决定。
func deltaTModelSecondsAtYear(model DeltaTModel, year float64) float64 {
	switch model {
	case DeltaTModelMS2004:
		u := (year - 1820) / 100
		return -20 + 32*u*u
	case DeltaTModelEspenakMeeus2006:
		return espenakMeeus2006Seconds(year)
	case DeltaTModelNASACanon2006:
		value := espenakMeeus2006Seconds(year)
		if year < 1955 {
			// 配对修正：把 −26.0″/cy² 口径的 ΔT 换到 −25.858″/cy²（1955 起原子时使其与月球历表无关）。
			d := year - 1955
			value += -0.000012932 * d * d
		}
		return value
	}
	return DeltaTSplineY(year)
}

// espenakMeeus2006Seconds 是 Espenak & Meeus 2006 对 −1999..+3000 的分段多项式（秒），
// 自变量 y = year + (month−0.5)/12；分段边界处两段取值一致（−500 处由 17203.7 强制连续）。
func espenakMeeus2006Seconds(y float64) float64 {
	switch {
	case y < -500:
		u := (y - 1820) / 100
		return -20 + 32*u*u
	case y < 500:
		u := y / 100
		return 10583.6 + u*(-1014.41+u*(33.78311+u*(-5.952053+u*(-0.1798452+u*(0.022174192+u*0.0090316521)))))
	case y < 1600:
		u := (y - 1000) / 100
		return 1574.2 + u*(-556.01+u*(71.23472+u*(0.319781+u*(-0.8503463+u*(-0.005050998+u*0.0083572073)))))
	case y < 1700:
		t := y - 1600
		return 120 - 0.9808*t - 0.01532*t*t + t*t*t/7129
	case y < 1800:
		t := y - 1700
		return 8.83 + t*(0.1603+t*(-0.0059285+t*(0.00013336-t/1174000)))
	case y < 1860:
		t := y - 1800
		return 13.72 + t*(-0.332447+t*(0.0068612+t*(0.0041116+t*(-0.00037436+
			t*(0.0000121272+t*(-0.0000001699+t*0.000000000875))))))
	case y < 1900:
		t := y - 1860
		return 7.62 + t*(0.5737+t*(-0.251754+t*(0.01680668+t*(-0.0004473624+t/233174))))
	case y < 1920:
		t := y - 1900
		return -2.79 + t*(1.494119+t*(-0.0598939+t*(0.0061966-t*0.000197)))
	case y < 1941:
		t := y - 1920
		return 21.20 + t*(0.84493+t*(-0.076100+t*0.0020936))
	case y < 1961:
		t := y - 1950
		return 29.07 + t*(0.407-t/233+t*t/2547)
	case y < 1986:
		t := y - 1975
		return 45.45 + t*(1.067-t/260-t*t/718)
	case y < 2005:
		t := y - 2000
		return 63.86 + t*(0.3345+t*(-0.060374+t*(0.0017275+t*(0.000651814+t*0.00002373599))))
	case y < 2050:
		t := y - 2000
		return 62.92 + t*(0.32217+t*0.005589)
	case y < 2150:
		u := (y - 1820) / 100
		return -20 + 32*u*u - 0.5628*(2150-y)
	}
	u := (y - 1820) / 100
	return -20 + 32*u*u
}
