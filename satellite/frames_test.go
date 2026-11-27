package satellite

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"
)

func TestOrbitPropagationConsistency(t *testing.T) {
	count := 0
	for _, g := range readPropagationFixtures(t) {
		orbit, err := newSatelliteOrbit(g.Elements)
		if err != nil {
			t.Fatal(err)
		}
		for i := len(g.Rows) - 1; i >= 0; i-- {
			var row []json.RawMessage
			if err = json.Unmarshal(g.Rows[i], &row); err != nil {
				t.Fatal(err)
			}
			var when time.Time
			if err = json.Unmarshal(row[0], &when); err != nil {
				t.Fatal(err)
			}
			got, gotErr := orbit.stateAt(when)
			want, wantErr := Propagate(g.Elements, when)
			if got != want || !errors.Is(gotErr, wantErr) {
				t.Fatalf("%s %s got=%+v err=%v want=%+v err=%v", g.Label, when, got, gotErr, want, wantErr)
			}
			count++
		}
	}
	if count != 387 {
		t.Fatalf("count=%d", count)
	}
	t.Logf("reused initialization matches %d propagation rows bit-for-bit", count)
}

func TestEarthFixedExternal(t *testing.T) {
	var maxR, maxV float64
	for _, g := range readLookFixtures(t).Groups {
		for _, row := range g.Ground {
			state, err := Propagate(g.Elements, row.Time)
			if err != nil {
				t.Fatal(err)
			}
			r, v := earthFixed(state)
			for i := range r {
				dr, dv := math.Abs(r[i]-row.Position[i]), math.Abs(v[i]-row.Velocity[i])
				maxR = math.Max(maxR, dr)
				maxV = math.Max(maxV, dv)
				if !finite(dr) || !finite(dv) || dr > 1e-4 || dv > 1e-6 {
					t.Fatalf("%s position=%.12g velocity=%.12g", g.Label, dr, dv)
				}
			}
		}
	}
	t.Logf("max Earth-fixed component position=%.12g km velocity=%.12g km/s", maxR, maxV)
}

func TestGeographicPositionSynthetic(t *testing.T) {
	a := 6378.137
	b := a * (1 - 1/298.257223563)
	for _, tc := range []struct {
		r           [3]float64
		lon, lat, h float64
	}{
		{[3]float64{a + 400, 0, 0}, 0, 0, 400},
		{[3]float64{-a - 400, 0, 0}, 180, 0, 400},
		{[3]float64{0, 0, b + 400}, 0, 90, 400},
		{[3]float64{0, 0, -b - 400}, 0, -90, 400},
	} {
		lon, lat, h := geographicPosition(tc.r)
		if math.Abs(lon-tc.lon) > 1e-12 || math.Abs(lat-tc.lat) > 1e-12 || math.Abs(h-tc.h) > 1e-9 {
			t.Fatalf("r=%v got=%g/%g/%g", tc.r, lon, lat, h)
		}
	}
	r := rotateZ(math.Pi/2, [3]float64{1, 0, 2})
	if math.Abs(r[0]) > 1e-15 || r[1] != -1 || r[2] != 2 {
		t.Fatalf("rotation handedness %v", r)
	}
}
