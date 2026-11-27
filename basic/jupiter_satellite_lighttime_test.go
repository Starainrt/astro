package basic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

func jupiterGalileanUnsharedLightTimeForTest(context jupiterGalileanObservationContext, index int) (JupiterGalileanState, Vector3) {
	lightTime := context.jupiterLightTime
	state := JupiterGalileanState{}
	result := Vector3{}
	includeSolarLongPeriod := jupiterGalileanUseSolarLongPeriod(context.jd)
	for iteration := 0; iteration < 8; iteration++ {
		targetJDE := context.jd - lightTime
		jupiterHelio := rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(targetJDE, context.inst), orbitJ2000Obliquity)
		state = jupiterGalileanSatelliteCorrectedStateAtET(targetJDE-jupiterGalileanReferenceJD, index, includeSolarLongPeriod)
		result = Vector3{
			jupiterHelio[0] + state.X - context.earthHelioJ2000[0],
			jupiterHelio[1] + state.Y - context.earthHelioJ2000[1],
			jupiterHelio[2] + state.Z - context.earthHelioJ2000[2],
		}
		nextLightTime := lightTimeDaysPerAU * vectorMagnitude(result)
		if math.Abs(nextLightTime-lightTime) < 1e-12 {
			break
		}
		lightTime = nextLightTime
	}
	return state, result
}

func TestJupiterGalileanLightTimeReuseConsistency(t *testing.T) {
	instances := []*Instance{
		nil,
		ephem.NewModelInstance(ephem.Semianalytic),
		ephem.NewModelInstance(ephem.Legacy),
		ephem.NewInstance(p3ProbeEngine{angle: 0.4, distance: 1.2}),
	}
	for instanceIndex, instance := range instances {
		for _, start := range []float64{2126309.5, 2305447.5, 2415020.5, 2451545, 2461041.5, 2461199.43, 2488069.5, 2670690.5, 2751545.5} {
			for sample := 0; sample < 3; sample++ {
				context := newJupiterGalileanObservationContext(start+float64(sample)*5/1440, instance)
				for index := 0; index < 4; index++ {
					gotState, gotVector := jupiterGalileanSatelliteAstrometricGeocentric(index, context.jd, context.jupiterLightTime, context.earthHelioJ2000, context.jupiterHelioJ2000, context.inst)
					wantState, wantVector := jupiterGalileanUnsharedLightTimeForTest(context, index)
					got := []float64{gotState.X, gotState.Y, gotState.Z, gotState.VX, gotState.VY, gotState.VZ, gotVector[0], gotVector[1], gotVector[2]}
					want := []float64{wantState.X, wantState.Y, wantState.Z, wantState.VX, wantState.VY, wantState.VZ, wantVector[0], wantVector[1], wantVector[2]}
					for component := range got {
						if math.Float64bits(got[component]) != math.Float64bits(want[component]) {
							t.Fatalf("instance=%d jd=%.12f satellite=%d component=%d: got=%.17g want=%.17g", instanceIndex, context.jd, index+1, component, got[component], want[component])
						}
					}
				}
			}
		}
	}
}

var jupiterGalileanLightTimeBenchmarkSink float64

func BenchmarkJupiterGalileanLightTimeVaryingTime(b *testing.B) {
	if _, ok := jupiterGalileanMutualMetricAt(2461041.5, 1, 4, nil); !ok {
		b.Fatal("invalid metric")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for sample := 0; sample < b.N; sample++ {
		metric, ok := jupiterGalileanMutualMetricAt(2461041.5+float64(sample%210240)*5/1440, 1, 4, nil)
		if !ok {
			b.Fatal("invalid metric")
		}
		jupiterGalileanLightTimeBenchmarkSink = metric
	}
}
