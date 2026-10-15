// 被测契约：非法 Body 与非有限历元必须报错而不是 panic 或静默返回 NaN；缓存槽不得跨天体串用。
package semianalytic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

func TestApparentRejectsUnsupportedBody(t *testing.T) {
	for _, body := range []ephem.Body{ephem.SSB, ephem.EMB, ephem.Earth, ephem.Body(-1), ephem.Body(2000001)} {
		if _, _, _, err := Apparent(body, 2451545); err != ephem.ErrUnsupportedBody {
			t.Errorf("Apparent(%d) err = %v, want %v", body, err, ephem.ErrUnsupportedBody)
		}
	}
}

func TestApparentCacheSlotsAreDistinct(t *testing.T) {
	const jd = 2451545.0
	sunDir, sunDist, _, err := Apparent(ephem.Sun, jd)
	if err != nil {
		t.Fatalf("Apparent(Sun) err = %v", err)
	}
	for _, body := range []ephem.Body{ephem.Mercury, ephem.Mars, ephem.Neptune, ephem.Moon} {
		dir, dist, _, err := Apparent(body, jd)
		if err != nil {
			t.Fatalf("Apparent(%d) err = %v", body, err)
		}
		if dir == sunDir && dist == sunDist {
			t.Errorf("Apparent(%d) returned the Sun entry", body)
		}
	}
}

func TestNonFiniteEpochIsOutOfRange(t *testing.T) {
	for _, jd := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, _, err := (Ephemeris{}).State(ephem.Mars, ephem.Sun, jd); err != ephem.ErrOutOfRange {
			t.Errorf("State(%v) err = %v, want %v", jd, err, ephem.ErrOutOfRange)
		}
		if _, _, _, err := (Ephemeris{}).Apparent(ephem.Mars, jd); err != ephem.ErrOutOfRange {
			t.Errorf("Apparent(%v) err = %v, want %v", jd, err, ephem.ErrOutOfRange)
		}
	}
}
