// 被测契约：新模型的矢量视位置管线与 DE441 视赤经赤纬（历元真赤道）一致到半解析表本身的精度。
package basic

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/internal/ephemtest"
)

func TestPlanetApparentVectorMatchesDE441(t *testing.T) {
	ephemtest.Pin(t, ephem.Semianalytic)
	f, err := os.Open("testdata/semianalytic/planet_apparent_de441.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fns := [8]func(float64) (float64, float64){
		nil, MercuryApparentRaDec, VenusApparentRaDec, MarsApparentRaDec, JupiterApparentRaDec,
		SaturnApparentRaDec, UranusApparentRaDec, NeptuneApparentRaDec,
	}
	rmsLimit := [8]float64{0, 15, 35, 30, 15, 10, 15, 20}
	maxLimit := [8]float64{0, 35, 110, 150, 40, 20, 45, 50}
	var ss, mx [8]float64
	var n [8]int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		p := strings.Split(sc.Text(), ",")
		b, _ := strconv.Atoi(p[0])
		jd, _ := strconv.ParseFloat(p[1], 64)
		wantRA, _ := strconv.ParseFloat(p[2], 64)
		wantDec, _ := strconv.ParseFloat(p[3], 64)
		ra, dec := fns[b](jd)
		dra := math.Mod(ra-wantRA+540, 360) - 180
		d := math.Hypot(dra*math.Cos(wantDec*math.Pi/180), dec-wantDec) * 3.6e6
		ss[b] += d * d
		n[b]++
		mx[b] = math.Max(mx[b], d)
	}
	for b := 1; b <= 7; b++ {
		rms := math.Sqrt(ss[b] / float64(n[b]))
		if rms > rmsLimit[b] || mx[b] > maxLimit[b] {
			t.Errorf("planet %d: rms %.1f mas (limit %.0f), max %.1f mas (limit %.0f)", b, rms, rmsLimit[b], mx[b], maxLimit[b])
		}
	}
}
