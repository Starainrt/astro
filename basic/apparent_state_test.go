// 被测契约：包级与实例视位置共用同一判据，零方向、非正距离、负光行时与非有限值一律报错。
package basic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

type fixedApparentEngine struct {
	dir       [3]float64
	dist, tau float64
}

func (e fixedApparentEngine) State(target, center ephem.Body, jdTDB float64) (pos, vel [3]float64, err error) {
	return pos, vel, ephem.ErrUnsupportedBody
}

func (e fixedApparentEngine) Coverage(ephem.Body) (start, end float64, ok bool) { return 0, 0, false }

func (e fixedApparentEngine) Apparent(ephem.Body, float64) ([3]float64, float64, float64, error) {
	return e.dir, e.dist, e.tau, nil
}

func TestApparentStateRejectsInvalid(t *testing.T) {
	nan := math.NaN()
	for _, c := range []struct {
		name string
		e    fixedApparentEngine
		ok   bool
	}{
		{"valid", fixedApparentEngine{[3]float64{1, 0, 0}, 1.5e8, 0.0058}, true},
		{"nan direction", fixedApparentEngine{[3]float64{nan, 0, 0}, 1.5e8, 0.0058}, false},
		{"zero direction", fixedApparentEngine{[3]float64{}, 1.5e8, 0.0058}, false},
		{"zero distance", fixedApparentEngine{[3]float64{1, 0, 0}, 0, 0.0058}, false},
		{"inf distance", fixedApparentEngine{[3]float64{1, 0, 0}, math.Inf(1), 0.0058}, false},
		{"negative tau", fixedApparentEngine{[3]float64{1, 0, 0}, 1.5e8, -1}, false},
		{"nan tau", fixedApparentEngine{[3]float64{1, 0, 0}, 1.5e8, nan}, false},
	} {
		_, _, _, _, ok, err := apparentState(ephem.Sun, 2451545, ephem.NewInstance(c.e))
		if c.ok && (err != nil || !ok) || !c.ok && err == nil {
			t.Errorf("%s: ok=%v err=%v", c.name, ok, err)
		}
	}
}
