package basic

import "math"

// JupiterGalileanEarthShadowRun 盘面上一段同状态水平扫描段，Y、X 以被食星半径为单位，X 东 Y 北 / one same-state scan segment on the passive disk, in passive radii, X east and Y north.
type JupiterGalileanEarthShadowRun struct {
	Y, X0, X1 float64
	State     int
}

// JupiterGalileanEarthShadow 地球所见被食星盘面：夜面、半影、本影的扫描段与覆盖比例 / Earth-view passive disk with night side, penumbra and umbra scan segments and fractions.
type JupiterGalileanEarthShadow struct {
	PhaseAngle       float64
	UmbraCoverage    float64
	AntumbraCoverage float64
	ShadowCoverage   float64
	RowHeight        float64
	Runs             []JupiterGalileanEarthShadowRun
	ShadowDirection  [2]float64 // 影锥轴相对被食星盘心的东、北单位方向，盘心在轴上时为零 / east-north unit direction from disk centre toward the shadow axis.
}

const (
	// JupiterGalileanEarthNight 夜面 / night side.
	JupiterGalileanEarthNight = 1
	// JupiterGalileanEarthPenumbra 半影 / penumbra.
	JupiterGalileanEarthPenumbra = 2
	// JupiterGalileanEarthUmbra 本影 / umbra.
	JupiterGalileanEarthUmbra = 3
	// JupiterGalileanEarthAntumbra 伪本影：本影锥顶之外，太阳盘被遮挡星整个挡在中间 / antumbra beyond the umbral apex.
	JupiterGalileanEarthAntumbra = 4
)

const earthShadowGrid = 96

// JupiterGalileanEarthShadowWithInstance 接收 TT 时刻地球所见互食被食星盘面 / Earth-view passive disk of a mutual eclipse at reception TT.
func JupiterGalileanEarthShadowWithInstance(jde float64, first, second int, inst *Instance) (JupiterGalileanEarthShadow, bool) {
	first, second, ok := normalizeJupiterGalileanMutualPair(first, second)
	if !ok || !isFinite(jde) {
		return JupiterGalileanEarthShadow{}, false
	}
	for _, passive := range []int{first, second} {
		emission := jupiterGalileanEclipseEmission(jde, passive, inst)
		geometry, valid := jupiterGalileanMutualEclipseGeometryAt(emission, first, second, inst)
		if !valid || geometry.Passive != passive {
			continue
		}
		return jupiterGalileanEarthShadowAt(jde, emission, geometry, inst), true
	}
	return JupiterGalileanEarthShadow{}, false
}

func jupiterGalileanEarthShadowAt(jde, emission float64, geometry jupiterGalileanMutualEclipseGeometry, inst *Instance) JupiterGalileanEarthShadow {
	frame := newJupiterGalileanEclipseFrame(emission, inst)
	active := frame.satellite(emission, geometry.Active)
	target := frame.satellite(emission, geometry.Passive)
	earth := rotateEclipticToEquatorial(earthHeliocentricVectorJ2000(jde, inst), orbitJ2000Obliquity)
	view := normalizeVector(Vector3{earth[0] - target[0], earth[1] - target[1], earth[2] - target[2]})
	toSun := normalizeVector(Vector3{-target[0], -target[1], -target[2]})
	ra, dec := vectorToRaDec(Vector3{-view[0], -view[1], -view[2]})
	east := Vector3{-math.Sin(ra * rad), math.Cos(ra * rad), 0}
	north := Vector3{-math.Cos(ra*rad) * math.Sin(dec*rad), -math.Sin(ra*rad) * math.Sin(dec*rad), math.Cos(dec * rad)}
	field := jupiterGalileanShadowField{
		active: active, axis: normalizeVector(active), distance: vectorMagnitude(active),
		activeRadius:  jupiterGalileanSatelliteRadiusJupiterRadii(geometry.Active) * jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM,
		passiveRadius: geometry.PassiveRadiusKM / astronomicalUnitKM,
	}
	surface := func(x, y float64) Vector3 {
		z := math.Sqrt(math.Max(0, 1-x*x-y*y))
		return Vector3{
			target[0] + field.passiveRadius*(x*east[0]+y*north[0]+z*view[0]),
			target[1] + field.passiveRadius*(x*east[1]+y*north[1]+z*view[1]),
			target[2] + field.passiveRadius*(x*east[2]+y*north[2]+z*view[2]),
		}
	}
	state := func(x, y float64) int {
		point := surface(x, y)
		normal := Vector3{point[0] - target[0], point[1] - target[1], point[2] - target[2]}
		if vectorDot(normal, toSun) <= 0 {
			return JupiterGalileanEarthNight
		}
		return field.state(point)
	}
	result := JupiterGalileanEarthShadow{
		PhaseAngle: math.Acos(math.Max(-1, math.Min(1, vectorDot(view, toSun)))) * deg,
		RowHeight:  2.0 / earthShadowGrid,
	}
	total, umbra, antumbra, shadowed := 0, 0, 0, 0
	for row := 0; row < earthShadowGrid; row++ {
		y := 1 - (float64(row)+0.5)*result.RowHeight
		half := math.Sqrt(math.Max(0, 1-y*y))
		current, start := -1, -half
		cells := int(math.Ceil(half * earthShadowGrid))
		for cell := 0; cell <= 2*cells; cell++ {
			value := -1
			x := -half + (float64(cell)+0.5)*half/float64(cells)
			if cell < 2*cells {
				value = state(x, y)
				total++
				switch value {
				case JupiterGalileanEarthUmbra:
					umbra++
				case JupiterGalileanEarthAntumbra:
					antumbra++
				}
				if value >= JupiterGalileanEarthPenumbra {
					shadowed++
				}
			}
			if value != current {
				if current > 0 {
					result.Runs = append(result.Runs, JupiterGalileanEarthShadowRun{Y: y, X0: start, X1: math.Min(half, -half+float64(cell)*half/float64(cells)), State: current})
				}
				current, start = value, -half+float64(cell)*half/float64(cells)
			}
		}
	}
	if total > 0 {
		result.UmbraCoverage = float64(umbra) / float64(total)
		result.AntumbraCoverage = float64(antumbra) / float64(total)
		result.ShadowCoverage = float64(shadowed) / float64(total)
	}
	offset := Vector3{target[0] - active[0], target[1] - active[1], target[2] - active[2]}
	axial := vectorDot(offset, field.axis)
	toAxis := Vector3{active[0] + axial*field.axis[0] - target[0], active[1] + axial*field.axis[1] - target[1], active[2] + axial*field.axis[2] - target[2]}
	if length := vectorMagnitude(toAxis); length > 0 {
		result.ShadowDirection = [2]float64{vectorDot(toAxis, east) / length, vectorDot(toAxis, north) / length}
	}
	return result
}

type jupiterGalileanShadowField struct {
	active, axis                          Vector3
	distance, activeRadius, passiveRadius float64
}

// state 返回 0 不在影中、半影、本影或伪本影；锥半径与互食几何同式，r_u < 0 时 |r_u| 为伪本影半径。
func (field jupiterGalileanShadowField) state(point Vector3) int {
	delta := Vector3{point[0] - field.active[0], point[1] - field.active[1], point[2] - field.active[2]}
	axial := vectorDot(delta, field.axis)
	if axial <= 0 {
		return 0
	}
	perpendicular := math.Sqrt(math.Max(0, vectorDot(delta, delta)-axial*axial))
	umbra := field.activeRadius - axial*(solarRadiusAU-field.activeRadius)/field.distance
	penumbra := field.activeRadius + axial*(solarRadiusAU+field.activeRadius)/field.distance
	if umbra > 0 && perpendicular < umbra {
		return JupiterGalileanEarthUmbra
	}
	if umbra < 0 && perpendicular < -umbra {
		return JupiterGalileanEarthAntumbra
	}
	if perpendicular < penumbra {
		return JupiterGalileanEarthPenumbra
	}
	return 0
}
