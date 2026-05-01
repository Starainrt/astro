package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

func TestJupiterCentralMeridiansMeeusExample43A(t *testing.T) {
	got := JupiterCentralMeridians(2448972.50068)
	ds, de := JupiterDSDE(2448972.50068)

	assertPlanetPhaseClose(t, "Jupiter.CMI", got.SystemI, 268.06, 0.02)
	assertPlanetPhaseClose(t, "Jupiter.CMII", got.SystemII, 72.74, 0.02)
	assertPlanetPhaseClose(t, "Jupiter.DS.Exact", ds, -1.733360091891, 1e-9)
	assertPlanetPhaseClose(t, "Jupiter.DE.Exact", de, -2.484625891057, 1e-9)
}

func TestJupiterCentralMeridianSystemIIIMatchesHorizonsBaseline(t *testing.T) {
	data, err := os.ReadFile("testdata/planet_physical_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []planetPhysicalSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}

	for _, sample := range samples {
		if sample.Body != "jupiter" {
			continue
		}
		date, err := time.Parse(time.RFC3339, sample.InputUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.InputUTC, err)
		}
		jd := TD2UT(Date2JDE(date.UTC()), true)
		got := JupiterCentralMeridians(jd)
		assertPlanetPhaseClose(t, "Jupiter."+sample.InputUTC+".CMIII", got.SystemIII, sample.SubEarthLongitude, 0.02)
	}
}

func TestJupiterCentralMeridiansNFullMatchesDefault(t *testing.T) {
	jd := TD2UT(Date2JDE(time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)), true)
	got := JupiterCentralMeridians(jd)
	gotN := JupiterCentralMeridiansN(jd, -1)
	ds, de := JupiterDSDE(jd)
	dsN, deN := JupiterDSDEN(jd, -1)
	if math.Float64bits(got.SystemI) != math.Float64bits(gotN.SystemI) {
		t.Fatalf("SystemI mismatch: got %.18f want %.18f", got.SystemI, gotN.SystemI)
	}
	if math.Float64bits(got.SystemII) != math.Float64bits(gotN.SystemII) {
		t.Fatalf("SystemII mismatch: got %.18f want %.18f", got.SystemII, gotN.SystemII)
	}
	if math.Float64bits(got.SystemIII) != math.Float64bits(gotN.SystemIII) {
		t.Fatalf("SystemIII mismatch: got %.18f want %.18f", got.SystemIII, gotN.SystemIII)
	}
	if math.Float64bits(ds) != math.Float64bits(dsN) {
		t.Fatalf("DS mismatch: got %.18f want %.18f", ds, dsN)
	}
	if math.Float64bits(de) != math.Float64bits(deN) {
		t.Fatalf("DE mismatch: got %.18f want %.18f", de, deN)
	}
}

func TestJupiterCentralMeridianAndDSDESampleSweepFiniteAndInRange(t *testing.T) {
	dates := []time.Time{
		time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1969, 7, 20, 20, 17, 40, 0, time.UTC),
		time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC),
		time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	for _, date := range dates {
		jd := TD2UT(Date2JDE(date.UTC()), true)
		meridians := JupiterCentralMeridians(jd)
		ds, de := JupiterDSDE(jd)
		physical := JupiterPhysical(jd)
		prefix := date.Format(time.RFC3339)

		assertFiniteRange(t, prefix+".SystemI", meridians.SystemI, 0, 360, true)
		assertFiniteRange(t, prefix+".SystemII", meridians.SystemII, 0, 360, true)
		assertFiniteRange(t, prefix+".SystemIII", meridians.SystemIII, 0, 360, true)
		assertFiniteRange(t, prefix+".DS", ds, -90, 90, false)
		assertFiniteRange(t, prefix+".DE", de, -90, 90, false)
		assertSameFloat(t, prefix+".CMIII", meridians.SystemIII, physical.SubEarthLongitude)
	}
}
