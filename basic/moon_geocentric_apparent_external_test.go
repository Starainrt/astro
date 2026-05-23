package basic

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

type moonGeocentricApparentSample struct {
	InputUTC          string  `json:"input_utc"`
	RightAscension    float64 `json:"right_ascension"`
	Declination       float64 `json:"declination"`
	EclipticLongitude float64 `json:"ecliptic_longitude"`
	EclipticLatitude  float64 `json:"ecliptic_latitude"`
}

func TestMoonGeocentricApparentCoordinatesMatchHorizonsBaseline(t *testing.T) {
	data, err := os.ReadFile("testdata/moon_geocentric_apparent_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []moonGeocentricApparentSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}
	if len(samples) == 0 {
		t.Fatal("empty moon apparent baseline")
	}

	for _, sample := range samples {
		date, err := time.Parse(time.RFC3339, sample.InputUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.InputUTC, err)
		}
		jd := TD2UT(Date2JDE(date.UTC()), true)
		prefix := "moon." + sample.InputUTC

		assertPlanetApparentAngleClose(t, prefix+".RightAscension", HMoonGeocentricApparentRa(jd), sample.RightAscension, 0.001)
		assertPlanetPhaseClose(t, prefix+".Declination", HMoonGeocentricApparentDec(jd), sample.Declination, 0.001)
		assertPlanetApparentAngleClose(t, prefix+".EclipticLongitude", HMoonApparentLo(jd), sample.EclipticLongitude, 0.001)
		assertPlanetPhaseClose(t, prefix+".EclipticLatitude", HMoonTrueBo(jd), sample.EclipticLatitude, 0.001)
	}
}

func TestMoonGeocentricTrueCoordinatesFollowDefinition(t *testing.T) {
	samples := []time.Time{
		time.Date(1900, 1, 14, 12, 0, 0, 0, time.UTC),
		time.Date(1950, 6, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2000, 2, 29, 18, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC),
		time.Date(2100, 8, 17, 9, 0, 0, 0, time.UTC),
	}

	for _, sample := range samples {
		jd := TD2UT(Date2JDE(sample.UTC()), true)
		wantRA, wantDec := LoBoToRaDec(jd, HMoonTrueLo(jd), HMoonTrueBo(jd))
		gotRA, gotDec := HMoonGeocentricTrueRaDec(jd)

		assertPlanetApparentAngleClose(t, sample.Format(time.RFC3339)+".TrueRightAscension", gotRA, wantRA, 1e-12)
		assertPlanetPhaseClose(t, sample.Format(time.RFC3339)+".TrueDeclination", gotDec, wantDec, 1e-12)
	}
}
