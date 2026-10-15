package basic

import (
	"math"
	"reflect"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/internal/geodata"
)

type p3LunarEngine struct {
	epoch        float64
	moonDistance float64
}

func (engine p3LunarEngine) State(target, center ephem.Body, jde float64) ([3]float64, [3]float64, error) {
	if center != ephem.Earth || target != ephem.Sun && target != ephem.Moon {
		return [3]float64{}, [3]float64{}, ephem.ErrUnsupportedBody
	}
	longitude, latitude, distance := 0.0, 0.0, astronomicalUnitKM
	if target == ephem.Moon {
		longitude = math.Pi + 0.23*(jde-engine.epoch)
		latitude = 0.002
		distance = engine.moonDistance
	}
	ecliptic := [3]float64{distance * math.Cos(latitude) * math.Cos(longitude), distance * math.Cos(latitude) * math.Sin(longitude), distance * math.Sin(latitude)}
	sine, cosine := math.Sincos(Obliquity1980(jde) * math.Pi / 180)
	equatorial := [3]float64{ecliptic[0], ecliptic[1]*cosine - ecliptic[2]*sine, ecliptic[1]*sine + ecliptic[2]*cosine}
	matrix := ltpPBMAT(2000 + (jde-2451545)/365.25)
	var position [3]float64
	for axis := range position {
		position[axis] = matrix[0][axis]*equatorial[0] + matrix[1][axis]*equatorial[1] + matrix[2][axis]*equatorial[2]
	}
	return position, [3]float64{}, nil
}

func (engine p3LunarEngine) Coverage(target ephem.Body) (float64, float64, bool) {
	return engine.epoch - 20, engine.epoch + 20, target == ephem.Sun || target == ephem.Moon
}

func p3AssertLunarResultEqual(t *testing.T, got, want LunarEclipseResult) {
	t.Helper()
	gotValue, wantValue := reflect.ValueOf(got), reflect.ValueOf(want)
	for field := 0; field < gotValue.NumField(); field++ {
		gotField, wantField := gotValue.Field(field), wantValue.Field(field)
		if gotField.Kind() == reflect.Float64 && math.IsNaN(gotField.Float()) && math.IsNaN(wantField.Float()) {
			continue
		}
		if !reflect.DeepEqual(gotField.Interface(), wantField.Interface()) {
			t.Errorf("%s: got %v, want %v", gotValue.Type().Field(field).Name, gotField.Interface(), wantField.Interface())
		}
	}
}

func TestP3LunarEclipseBuiltInCompatibility(t *testing.T) {
	previous := ephem.Active()
	t.Cleanup(func() { ephem.SetActive(previous) })
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		ephem.SetActive(model)
		inst := ephem.NewModelInstance(model)
		for _, date := range [][3]int{{2022, 11, 8}, {2023, 5, 5}, {2023, 10, 28}, {2024, 4, 23}} {
			seed := JDCalc(date[0], date[1], float64(date[2]))
			for _, calculate := range []struct {
				old func(float64) LunarEclipseResult
				new func(float64, *ephem.Instance) LunarEclipseResult
			}{
				{LunarEclipse, LunarEclipseWithInstance},
				{LunarEclipseDanjon, LunarEclipseDanjonWithInstance},
				{LunarEclipseChauvenet, LunarEclipseChauvenetWithInstance},
			} {
				want := calculate.old(seed)
				p3AssertLunarResultEqual(t, calculate.new(seed, nil), want)
				p3AssertLunarResultEqual(t, calculate.new(seed, inst), want)
			}
		}
	}
}

func TestP3LunarEclipseInstancesStayIsolated(t *testing.T) {
	const seed = 2459891.9
	first := ephem.NewInstance(p3LunarEngine{epoch: seed, moonDistance: 384400})
	second := ephem.NewInstance(p3LunarEngine{epoch: seed + 0.25, moonDistance: 370000})
	global, model, generation := ephem.Global(), ephem.Active(), ephem.Generation()
	firstResult := LunarEclipseWithInstance(seed, first)
	secondResult := LunarEclipseWithInstance(seed, second)
	if !firstResult.HasTotal || !secondResult.HasTotal {
		t.Fatalf("synthetic eclipses must be total: %+v / %+v", firstResult, secondResult)
	}
	if math.Abs(secondResult.Maximum-firstResult.Maximum-0.25) > 1e-6 {
		t.Errorf("maximum did not follow instance: %.12f / %.12f", firstResult.Maximum, secondResult.Maximum)
	}
	if firstResult.TotalEnd-firstResult.TotalStart == secondResult.TotalEnd-secondResult.TotalStart {
		t.Error("contact durations did not follow instance distances")
	}
	for _, result := range []LunarEclipseResult{firstResult, secondResult} {
		contacts := []float64{result.PenumbralStart, result.PartialStart, result.TotalStart, result.Maximum, result.TotalEnd, result.PartialEnd, result.PenumbralEnd}
		for index := 1; index < len(contacts); index++ {
			if !isFiniteFloat(contacts[index]) || contacts[index] <= contacts[index-1] {
				t.Errorf("invalid contact order: %v", contacts)
			}
		}
	}
	p3AssertLunarResultEqual(t, LunarEclipseWithInstance(seed, first), firstResult)
	if ephem.Global() != global || ephem.Active() != model || ephem.Generation() != generation {
		t.Error("explicit lunar eclipse changed the global ephemeris")
	}
}

func TestP3LunarEclipseGeometryAndDiagramUseInstance(t *testing.T) {
	const seed = 2459891.9
	options := LunarEclipseDiagramOptions{StepDays: 0.02}
	for _, shadowModel := range []LunarEclipseShadowModel{LunarEclipseShadowModelDanjon, LunarEclipseShadowModelChauvenet} {
		inst := ephem.NewInstance(p3LunarEngine{epoch: seed + 0.15, moonDistance: 380000})
		calculate := LunarEclipseDiagramDanjonWithInstance
		if shadowModel == LunarEclipseShadowModelChauvenet {
			calculate = LunarEclipseDiagramChauvenetWithInstance
		}
		diagram := calculate(seed, options, inst)
		geometry := LunarEclipseShadowGeometryAtModelWithInstance(diagram.Eclipse.Maximum, shadowModel, inst)
		if len(diagram.Points) == 0 || math.Abs(geometry.MoonDistanceEarthRadii-380000/geodata.EarthEquatorialRadiusKM) > 1e-10 {
			t.Fatalf("diagram/geometry lost instance: %+v", geometry)
		}
		if math.Abs(diagram.Eclipse.Maximum-seed-0.15) > 0.001 {
			t.Errorf("diagram eclipse lost instance: %v", diagram.Eclipse.Maximum)
		}
		maximumState := computeLunarShadowStateWithInstance(diagram.Eclipse.Maximum, lunarEclipseShadowModel(shadowModel), inst)
		for _, point := range diagram.Points {
			plane := lunarEclipsePlaneStateAtWithInstance(point.JDE, inst)
			if point.X != plane.x/maximumState.moonRadiusRad || point.Y != plane.y/maximumState.moonRadiusRad {
				t.Errorf("diagram sample lost instance: %+v", point)
			}
		}
	}
	for _, calculate := range []struct {
		old func(float64, LunarEclipseDiagramOptions) LunarEclipseDiagramResult
		new func(float64, LunarEclipseDiagramOptions, *ephem.Instance) LunarEclipseDiagramResult
	}{
		{LunarEclipseDiagram, LunarEclipseDiagramWithInstance},
		{LunarEclipseDiagramDanjon, LunarEclipseDiagramDanjonWithInstance},
		{LunarEclipseDiagramChauvenet, LunarEclipseDiagramChauvenetWithInstance},
	} {
		got, want := calculate.new(seed, options, nil), calculate.old(seed, options)
		p3AssertLunarResultEqual(t, got.Eclipse, want.Eclipse)
		got.Eclipse, want.Eclipse = LunarEclipseResult{}, LunarEclipseResult{}
		if !reflect.DeepEqual(got, want) {
			t.Error("nil changed diagram geometry or samples")
		}
	}
}

func TestP3LunarEclipseFailureDoesNotFallback(t *testing.T) {
	inst := ephem.NewInstance(p3StateEngine{err: ephem.ErrOutOfRange})
	result := LunarEclipseWithInstance(2459891.9, inst)
	if !math.IsNaN(result.Maximum) || result.HasPenumbral || result.HasPartial || result.HasTotal {
		t.Errorf("failed engine produced an eclipse: %+v", result)
	}
	diagram := LunarEclipseDiagramWithInstance(2459891.9, LunarEclipseDiagramOptions{}, inst)
	if !math.IsNaN(diagram.Eclipse.Maximum) || len(diagram.Points) != 0 {
		t.Error("failed engine produced diagram samples")
	}
}
