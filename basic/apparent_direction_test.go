// 被测契约：日月视赤经纬单值与双值接口一致，且黄经、黄纬取自同一视方向。
package basic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/internal/ephemtest"
)

func TestSunMoonApparentRaDecUseFullDirection(t *testing.T) {
	ephemtest.Pin(t, ephem.Semianalytic)
	const jde = 2460409.25
	for _, body := range []ephem.Body{ephem.Sun, ephem.Moon} {
		lo, bo, _, _, ok := vectorApparent(body, jde, -1)
		if !ok {
			t.Fatalf("body %d: no vector apparent position", body)
		}
		wantRA, wantDec := LoBoToRaDec(jde, lo, bo)
		var ra, dec float64
		if body == ephem.Sun {
			ra, dec = HSunApparentRaDec(jde)
			if r, d := HSunApparentRa(jde), HSunApparentDec(jde); r != ra || d != dec {
				t.Errorf("Sun single (%v, %v) != pair (%v, %v)", r, d, ra, dec)
			}
		} else {
			ra, dec = HMoonGeocentricApparentRaDecN(jde, -1)
		}
		if math.Abs(ra-wantRA)*3600 > 1e-6 || math.Abs(dec-wantDec)*3600 > 1e-6 {
			t.Errorf("body %d: (%v, %v), want (%v, %v)", body, ra, dec, wantRA, wantDec)
		}
	}
}
