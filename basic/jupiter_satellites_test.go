package basic

import (
	"math"
	"testing"
)

const galileanSampleToleranceAU = 1e-15
const galileanSampleToleranceAUDay = 1e-15

type galileanSample struct {
	name  string
	index int
	state JupiterGalileanState
}

func TestJupiterGalileanSatelliteStateMatchesIMCCESample(t *testing.T) {
	samples := []galileanSample{
		{
			name:  "Io",
			index: 1,
			state: JupiterGalileanState{X: 2.671999370920431e-003, Y: 7.644018403387422e-004, Z: 4.087344808808269e-004, VX: -3.116203340625001e-003, VY: 8.645679572984422e-003, VZ: 4.066210333795641e-003},
		},
		{
			name:  "Europa",
			index: 2,
			state: JupiterGalileanState{X: -3.751373844521062e-003, Y: -2.136179970327756e-003, Z: -1.056765216826830e-003, VX: 4.310591732986133e-003, VY: -6.143199976514738e-003, VZ: -2.800434328620005e-003},
		},
		{
			name:  "Ganymede",
			index: 3,
			state: JupiterGalileanState{X: -5.490036250442612e-003, Y: -4.112229247907583e-003, Z: -2.033821277493470e-003, VX: 4.036147912130572e-003, VY: -4.364866691392988e-003, VZ: -2.037111499364415e-003},
		},
		{
			name:  "Callisto",
			index: 4,
			state: JupiterGalileanState{X: 2.172082907229073e-003, Y: 1.118792302205555e-002, Z: 5.322275059416266e-003, VX: -4.662583658656747e-003, VY: 7.976685330152526e-004, VZ: 3.092058747362411e-004},
		},
	}

	const jd = 2451545.0
	maxPosDiff := 0.0
	maxVelDiff := 0.0
	for _, sample := range samples {
		got := JupiterGalileanSatelliteState(jd, sample.index)
		posDiffs := []float64{
			math.Abs(got.X - sample.state.X),
			math.Abs(got.Y - sample.state.Y),
			math.Abs(got.Z - sample.state.Z),
		}
		velDiffs := []float64{
			math.Abs(got.VX - sample.state.VX),
			math.Abs(got.VY - sample.state.VY),
			math.Abs(got.VZ - sample.state.VZ),
		}
		for i, diff := range posDiffs {
			if diff > maxPosDiff {
				maxPosDiff = diff
			}
			if diff > galileanSampleToleranceAU {
				t.Fatalf("%s position[%d] mismatch: got %.18e want %.18e", sample.name, i, []float64{got.X, got.Y, got.Z}[i], []float64{sample.state.X, sample.state.Y, sample.state.Z}[i])
			}
		}
		for i, diff := range velDiffs {
			if diff > maxVelDiff {
				maxVelDiff = diff
			}
			if diff > galileanSampleToleranceAUDay {
				t.Fatalf("%s velocity[%d] mismatch: got %.18e want %.18e", sample.name, i, []float64{got.VX, got.VY, got.VZ}[i], []float64{sample.state.VX, sample.state.VY, sample.state.VZ}[i])
			}
		}
	}
	// Official IMCCE README example for V1_1 at JD 2451545.0.
	t.Logf("galilean IMCCE sample max diff: position=%.3e AU velocity=%.3e AU/day", maxPosDiff, maxVelDiff)
}
