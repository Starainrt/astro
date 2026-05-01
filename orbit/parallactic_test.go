package orbit

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestParallacticAngleMatchesHourAngleForm(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 20, 0, 0, 0, time.FixedZone("CST", 8*3600))

	got := ParallacticAngle(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	position := ApparentTopocentricEquatorial(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	want := basic.ParallacticAngleByHourAngle(
		HourAngle(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters),
		position.Dec,
		shanghaiLat,
	)
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("parallactic angle mismatch: got %.15f want %.15f", got, want)
	}
}
