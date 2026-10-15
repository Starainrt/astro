package eclipse

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

type p3LunarShiftEngine struct {
	base   *ephem.Instance
	offset float64
}

func (engine p3LunarShiftEngine) State(target, center ephem.Body, jde float64) ([3]float64, [3]float64, error) {
	return engine.base.State(target, center, jde+engine.offset)
}

func (engine p3LunarShiftEngine) Apparent(target ephem.Body, jde float64) ([3]float64, float64, float64, error) {
	return engine.base.Apparent(target, jde+engine.offset)
}

func (engine p3LunarShiftEngine) Coverage(target ephem.Body) (float64, float64, bool) {
	start, end, ok := engine.base.Coverage(target)
	return start - engine.offset, end - engine.offset, ok
}

func TestP3LunarQueriesPreserveNilAndLegacy(t *testing.T) {
	previous := ephem.Active()
	t.Cleanup(func() { ephem.SetActive(previous) })
	date := time.Date(2022, 11, 8, 12, 0, 0, 0, time.UTC)
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		ephem.SetActive(model)
		inst := ephem.NewModelInstance(model)
		for _, query := range []struct {
			old func(time.Time) (LunarEclipseInfo, bool)
			new func(time.Time, *ephem.Instance) (LunarEclipseInfo, bool)
		}{
			{LunarEclipseOnDate, LunarEclipseOnDateWithInstance},
			{LunarEclipseOnDateDanjon, LunarEclipseOnDateDanjonWithInstance},
			{LunarEclipseOnDateChauvenet, LunarEclipseOnDateChauvenetWithInstance},
		} {
			want, wantOK := query.old(date)
			got, gotOK := query.new(date, nil)
			if gotOK != wantOK || !reflect.DeepEqual(got, want) {
				t.Error("nil changed local-date query")
			}
			if model == ephem.Legacy {
				got, gotOK = query.new(date, inst)
				if gotOK != wantOK || !reflect.DeepEqual(got, want) {
					t.Error("explicit Legacy changed local-date query")
				}
			}
		}
		for _, query := range []struct {
			old func(time.Time) LunarEclipseInfo
			new func(time.Time, *ephem.Instance) LunarEclipseInfo
		}{
			{LastLunarEclipse, LastLunarEclipseWithInstance},
			{LastLunarEclipseDanjon, LastLunarEclipseDanjonWithInstance},
			{LastLunarEclipseChauvenet, LastLunarEclipseChauvenetWithInstance},
			{NextLunarEclipse, NextLunarEclipseWithInstance},
			{NextLunarEclipseDanjon, NextLunarEclipseDanjonWithInstance},
			{NextLunarEclipseChauvenet, NextLunarEclipseChauvenetWithInstance},
			{ClosestLunarEclipse, ClosestLunarEclipseWithInstance},
			{ClosestLunarEclipseDanjon, ClosestLunarEclipseDanjonWithInstance},
			{ClosestLunarEclipseChauvenet, ClosestLunarEclipseChauvenetWithInstance},
		} {
			want := query.old(date)
			if got := query.new(date, nil); !reflect.DeepEqual(got, want) {
				t.Error("nil changed directional search")
			}
			if model == ephem.Legacy {
				if got := query.new(date, inst); !reflect.DeepEqual(got, want) {
					t.Error("explicit Legacy changed directional search")
				}
			}
		}
	}
}

func TestP3LunarQueriesUseInstanceThroughout(t *testing.T) {
	previous := ephem.Active()
	t.Cleanup(func() { ephem.SetActive(previous) })
	base := ephem.NewModelInstance(ephem.Semianalytic)
	first := ephem.NewInstance(p3LunarShiftEngine{base: base, offset: 0.1})
	second := ephem.NewInstance(p3LunarShiftEngine{base: base, offset: 0.2})
	date := time.Date(2022, 11, 8, 12, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	global, model, generation := ephem.Global(), ephem.Active(), ephem.Generation()
	firstInfo, firstOK := LunarEclipseOnDateWithInstance(date, first)
	secondInfo, secondOK := LunarEclipseOnDateWithInstance(date, second)
	if !firstOK || !secondOK || !firstInfo.HasTotal || !secondInfo.HasTotal {
		t.Fatalf("explicit query lost total eclipse: %+v / %+v", firstInfo, secondInfo)
	}
	if firstInfo.HasSaros || secondInfo.HasSaros {
		t.Error("custom engines must not inherit built-in Saros metadata")
	}
	if math.Abs(firstInfo.GreatestJDE-secondInfo.GreatestJDE-0.1) > 0.001 {
		t.Errorf("event time did not follow instances: %v / %v", firstInfo.GreatestJDE, secondInfo.GreatestJDE)
	}
	if firstInfo.Maximum.Location() != date.Location() || len(firstInfo.ContactPoints) != 6 {
		t.Error("local timezone or contacts were lost")
	}
	if ephem.Global() != global || ephem.Active() != model || ephem.Generation() != generation {
		t.Error("query changed global ephemeris state")
	}
	for _, active := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		ephem.SetActive(active)
		got, ok := LunarEclipseOnDateWithInstance(date, first)
		if !ok || !reflect.DeepEqual(got, firstInfo) {
			t.Error("result, contact orientation or shadow geometry read the global model")
		}
		for _, query := range []func(time.Time, *ephem.Instance) LunarEclipseInfo{LastLunarEclipseWithInstance, NextLunarEclipseWithInstance, ClosestLunarEclipseWithInstance} {
			got := query(firstInfo.Maximum.Add(-time.Minute), first)
			if got.Type == LunarEclipseNone || got.Maximum.IsZero() {
				t.Error("directional search lost instance")
			}
		}
	}
}

func TestP3LunarSarosSelectsInstanceModel(t *testing.T) {
	previous := ephem.Active()
	t.Cleanup(func() { ephem.SetActive(previous) })
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		ephem.SetActive(model)
		for _, seed := range []float64{basic.JDCalc(2022, 11, 8), basic.JDCalc(7000, 1, 1), basic.JDCalc(-5000, 1, 1)} {
			want, wantOK := lunarSarosInfoWithInstance(seed, inst)
			ephem.SetActive(ephem.Model(1 - int(model)))
			got, gotOK := lunarSarosInfoWithInstance(seed, inst)
			if gotOK != wantOK || got != want {
				t.Errorf("Saros read active model instead of instance: %+v / %+v", got, want)
			}
		}
	}
}

func TestP3LunarQueriesStopOnEngineFailure(t *testing.T) {
	inst := ephem.NewInstance(nil)
	date := time.Date(2022, 11, 8, 12, 0, 0, 0, time.UTC)
	if info, ok := LunarEclipseOnDateWithInstance(date, inst); ok || info.Type != "" {
		t.Error("failed engine produced a local-date eclipse")
	}
	for _, query := range []func(time.Time, *ephem.Instance) LunarEclipseInfo{LastLunarEclipseWithInstance, NextLunarEclipseWithInstance, ClosestLunarEclipseWithInstance} {
		if info := query(date, inst); !info.Maximum.IsZero() {
			t.Error("failed engine fell back during directional search")
		}
	}
}
