package sgp4

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
)

const (
	positionToleranceKM  = 1e-6
	velocityToleranceKMS = 1e-9
)

type verificationFixture struct {
	Groups []struct {
		Label, Source string
		Elements      Elements
		Rows          []json.RawMessage
	}
	Sidereal []struct {
		Label, Source string
		JDUT1         float64 `json:"jd_ut1"`
		Radians       float64
	}
}

func readVerificationFixture(t *testing.T) verificationFixture {
	t.Helper()
	b, err := os.ReadFile("testdata/vallado_verification.json")
	if err != nil {
		t.Fatal(err)
	}
	var f verificationFixture
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Groups) != 33 {
		t.Fatalf("groups=%d", len(f.Groups))
	}
	return f
}

func TestPropagationExternal(t *testing.T) {
	f := readVerificationFixture(t)
	var maxR, maxV float64
	rows, valid, failed := 0, 0, 0
	for _, g := range f.Groups {
		if g.Label == "" || g.Source == "" {
			t.Fatal("missing provenance")
		}
		var s Record
		s.Init(g.Elements)
		for _, raw := range g.Rows {
			var row []json.RawMessage
			if err := json.Unmarshal(raw, &row); err != nil || len(row) != 4 {
				t.Fatalf("invalid row: %s", raw)
			}
			var minutes float64
			var wantCode int
			var wantR, wantV [3]float64
			for i, v := range []interface{}{&minutes, &wantCode, &wantR, &wantV} {
				if err := json.Unmarshal(row[i], v); err != nil {
					t.Fatal(err)
				}
			}
			r, v, code := s.Propagate(minutes)
			rows++
			if code != wantCode {
				t.Fatalf("%s t=%.17g code=%d want=%d", g.Label, minutes, code, wantCode)
			}
			if code != 0 {
				failed++
				continue
			}
			valid++
			var dr, dv float64
			for i := range r {
				dr = math.Hypot(dr, r[i]-wantR[i])
				dv = math.Hypot(dv, v[i]-wantV[i])
			}
			maxR = math.Max(maxR, dr)
			maxV = math.Max(maxV, dv)
			if math.IsNaN(dr) || math.IsNaN(dv) || dr > positionToleranceKM || dv > velocityToleranceKMS {
				t.Fatalf("%s t=%.17g position=%.17g velocity=%.17g", g.Label, minutes, dr, dv)
			}
		}
	}
	if rows != 673 || valid != 666 || failed != 7 {
		t.Fatalf("rows=%d valid=%d errors=%d", rows, valid, failed)
	}
	t.Logf("groups=%d rows=%d valid=%d errors=%d max_position=%.12g km max_velocity=%.12g km/s", len(f.Groups), rows, valid, failed, maxR, maxV)
}

func TestGSTimeExternal(t *testing.T) {
	f := readVerificationFixture(t)
	var max float64
	for _, g := range f.Sidereal {
		got := GSTime(g.JDUT1)
		d := math.Abs(got - g.Radians)
		max = math.Max(max, d)
		if math.IsNaN(got) || math.IsInf(got, 0) || got < 0 || got >= 2*math.Pi || d > 1e-12 {
			t.Fatalf("%s got=%.17g want=%.17g", g.Label, got, g.Radians)
		}
	}
	if len(f.Sidereal) != 7 {
		t.Fatal("missing sidereal samples")
	}
	t.Logf("max_sidereal=%.12g rad", max)
}

func TestPropagationOrderConsistency(t *testing.T) {
	f := readVerificationFixture(t)
	for _, g := range f.Groups {
		var s Record
		s.Init(g.Elements)
		for i := len(g.Rows) - 1; i >= 0; i-- {
			var row []json.RawMessage
			if err := json.Unmarshal(g.Rows[i], &row); err != nil {
				t.Fatal(err)
			}
			var minutes float64
			if err := json.Unmarshal(row[0], &minutes); err != nil {
				t.Fatal(err)
			}
			r, v, code := s.Propagate(minutes)
			var fresh Record
			fresh.Init(g.Elements)
			wantR, wantV, wantCode := fresh.Propagate(minutes)
			if r != wantR || v != wantV || code != wantCode {
				t.Fatalf("%s time=%.17g depends on call order", g.Label, minutes)
			}
		}
	}
}

func TestPropagationRecordIsolationConsistency(t *testing.T) {
	fixture := readVerificationFixture(t)
	resonant := 0
	for _, group := range fixture.Groups {
		var baseline Record
		baseline.Init(group.Elements)
		record := baseline
		if record.irez != 0 {
			resonant++
		}
		for _, minutes := range []float64{1440, -1440} {
			position, velocity, code := record.Propagate(minutes)
			var fresh Record
			fresh.Init(group.Elements)
			wantPosition, wantVelocity, wantCode := fresh.Propagate(minutes)
			if code != wantCode {
				t.Fatalf("%s t=%g code=%d want=%d", group.Label, minutes, code, wantCode)
			}
			for index := range position {
				if math.Float64bits(position[index]) != math.Float64bits(wantPosition[index]) || math.Float64bits(velocity[index]) != math.Float64bits(wantVelocity[index]) {
					t.Fatalf("%s t=%g component=%d differs from fresh Record", group.Label, minutes, index)
				}
			}
			gotFields, wantFields := reflect.ValueOf(record), reflect.ValueOf(baseline)
			if field := changedRecordField(gotFields, wantFields); field != "" {
				t.Fatalf("%s t=%g changed Record.%s", group.Label, minutes, field)
			}
		}
	}
	if resonant == 0 {
		t.Fatal("missing deep-space resonance cases")
	}
	t.Logf("groups=%d resonant=%d", len(fixture.Groups), resonant)
}

func TestRecordReinitializationConsistency(t *testing.T) {
	fixture := readVerificationFixture(t)
	var reused Record
	for _, group := range fixture.Groups {
		reused.Init(group.Elements)
		var fresh Record
		fresh.Init(group.Elements)
		if field := changedRecordField(reflect.ValueOf(reused), reflect.ValueOf(fresh)); field != "" {
			t.Fatalf("%s: reinitialization changed Record.%s", group.Label, field)
		}
	}
}

func changedRecordField(gotFields, wantFields reflect.Value) string {
	for index := 0; index < gotFields.NumField(); index++ {
		got, want := gotFields.Field(index), wantFields.Field(index)
		var equal bool
		switch got.Kind() {
		case reflect.Float64:
			equal = math.Float64bits(got.Float()) == math.Float64bits(want.Float())
		case reflect.Int:
			equal = got.Int() == want.Int()
		case reflect.Uint8:
			equal = got.Uint() == want.Uint()
		case reflect.Struct:
			if field := changedRecordField(got, want); field != "" {
				return field
			}
			equal = true
		}
		if !equal {
			return gotFields.Type().Field(index).Name
		}
	}
	return ""
}
