package astro_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/planet/spk"
)

type publicEphemerisEngine struct{}

func (publicEphemerisEngine) State(target, center astro.Body, jde float64) ([3]float64, [3]float64, error) {
	return [3]float64{149597870.7, 0, 0}, [3]float64{}, nil
}

func (publicEphemerisEngine) Coverage(target astro.Body) (float64, float64, bool) {
	return 2400000, 2500000, true
}

var _ astro.Engine = publicEphemerisEngine{}
var _ astro.Engine = (*spk.File)(nil)
var _ func(float64, *astro.Instance) basic.LunarEclipseResult = basic.LunarEclipseWithInstance
var _ func(float64, *astro.Instance) float64 = basic.MoonPhaseWithInstance
var _ func(float64, *astro.Instance) float64 = basic.MercuryApparentLoWithInstance
var _ func(float64, *astro.Instance) float64 = basic.HSunTrueLoWithInstance
var _ func(int, int, *astro.Instance) float64 = basic.GetJQTimeWithInstance

func TestP3PublicInstanceConstruction(t *testing.T) {
	for _, model := range []astro.EphemerisModel{astro.EphemerisLegacy, astro.EphemerisSemianalytic} {
		inst := mustInstance(t, model)
		if inst.Model() != model {
			t.Fatalf("model constructor lost identity: %v", inst.Model())
		}
		if result := basic.LunarEclipseWithInstance(basic.JDCalc(2022, 11, 8), inst); !result.HasTotal {
			t.Errorf("public model instance lost eclipse: %+v", result)
		}
	}
	inst := astro.NewInstance(publicEphemerisEngine{})
	position, _, err := inst.State(astro.BodySun, astro.BodyEarth, 2451545)
	if err != nil || position[0] != 149597870.7 || inst.Model() != astro.EphemerisCustom {
		t.Fatalf("custom instance: %v, %v", position, err)
	}
	if distance := basic.EarthAwayNWithInstance(2451545, -1, inst); math.Abs(distance-1) > 1e-9 {
		t.Errorf("public instance could not reach basic: %v", distance)
	}
	_, _, _, err = inst.Apparent(astro.BodySun, 2451545)
	if !errors.Is(err, astro.ErrEphemerisUnsupportedApparent) {
		t.Errorf("public error identity was lost: %v", err)
	}
	if longitude := basic.HSunTrueLoWithInstance(2451545, astro.NewInstance(nil)); !math.IsNaN(longitude) {
		t.Errorf("empty custom engine fell back: %v", longitude)
	}
}

func TestNewModelInstanceRejectsNonBuiltin(t *testing.T) {
	for _, model := range []astro.EphemerisModel{astro.EphemerisCustom, astro.EphemerisModel(42)} {
		if inst, err := astro.NewModelInstance(model); inst != nil || !errors.Is(err, astro.ErrUnknownModel) {
			t.Errorf("model %v: inst=%v err=%v", model, inst, err)
		}
	}
}

func TestP3PublicLunarDateQuery(t *testing.T) {
	inst := mustInstance(t, astro.EphemerisSemianalytic)
	date := time.Date(2022, 11, 8, 12, 0, 0, 0, time.UTC)
	info, ok := eclipse.LunarEclipseOnDateWithInstance(date, inst)
	if !ok || !info.HasTotal {
		t.Errorf("public instance did not reach eclipse query: %+v", info)
	}
}
