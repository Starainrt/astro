package basic

import (
	"errors"
	"math"
	"testing"
)

func TestBodyOrientationConsistencyPoleRotation(t *testing.T) {
	for _, sample := range []struct {
		body SkyBody
		pole func(float64) (float64, float64, float64)
		rate float64
	}{
		{SkyMercury, mercuryPoleRotation, 6.1385108},
		{SkyVenus, venusPoleRotation, -1.4813688},
		{SkyMars, marsPoleRotation, 350.891982443297},
		{SkyJupiter, jupiterPoleRotation, 870.536},
		{SkySaturn, saturnPoleRotation, 810.7939024},
		{SkyUranus, uranusPoleRotation, -501.1600928},
		{SkyNeptune, neptunePoleRotation, 541.1397757},
	} {
		for _, jdTT := range []float64{2451545 - 50*365.25, 2451545, 2451545 + 50*365.25} {
			got, err := BodyOrientationAt(sample.body, jdTT)
			if err != nil {
				t.Fatal(err)
			}
			ra, dec, w := sample.pole(jdTT)
			wantW := math.Mod(w, 360)
			if wantW < 0 {
				wantW += 360
			}
			if math.Float64bits(got.PoleRA) != math.Float64bits(ra) || math.Float64bits(got.PoleDec) != math.Float64bits(dec) || math.Float64bits(got.PrimeMeridianAngle) != math.Float64bits(wantW) {
				t.Errorf("body=%d jdTT=%.8f got=%+v want=(%.17g,%.17g,%.17g)", sample.body, jdTT, got, ra, dec, wantW)
			}
			if got.RotationRate != sample.rate || got.PrimeMeridianAngle < 0 || got.PrimeMeridianAngle >= 360 {
				t.Errorf("body=%d: %+v", sample.body, got)
			}
		}
	}
}

func TestBodyOrientationSyntheticValidity(t *testing.T) {
	for _, sample := range []struct {
		body     SkyBody
		from, to float64
	}{
		{SkyMercury, 1800, 2200}, {SkyVenus, 1800, 2200}, {SkyMars, 1800, 2200},
		{SkyJupiter, 1900, 2100},
		{SkySaturn, 1950, 2050}, {SkyUranus, 1950, 2050}, {SkyNeptune, 1950, 2050},
	} {
		for _, boundary := range []float64{2451545 + (sample.from-2000)*365.25, 2451545 + (sample.to-2000)*365.25} {
			for _, jdTT := range []float64{math.Nextafter(boundary, math.Inf(-1)), boundary, math.Nextafter(boundary, math.Inf(1))} {
				got, err := BodyOrientationAt(sample.body, jdTT)
				if err != nil {
					t.Fatal(err)
				}
				want := jdTT < 2451545+(sample.from-2000)*365.25 || jdTT > 2451545+(sample.to-2000)*365.25
				if got.OutsideValidity != want {
					t.Errorf("body=%d jdTT=%.17g outside=%v want=%v", sample.body, jdTT, got.OutsideValidity, want)
				}
			}
		}
	}
}

func TestBodyOrientationSyntheticFixedFrame(t *testing.T) {
	for _, sample := range []struct {
		orientation BodyOrientation
		want        Matrix3
	}{
		{BodyOrientation{PoleRA: 0, PoleDec: 0}, Matrix3{{0, 0, 1}, {1, 0, 0}, {0, 1, 0}}},
		{BodyOrientation{PoleRA: 0, PoleDec: 0, PrimeMeridianAngle: 90}, Matrix3{{0, 0, 1}, {0, -1, 0}, {1, 0, 0}}},
	} {
		got := bodyFixedToICRF(sample.orientation)
		for row := range got {
			for col := range got[row] {
				if math.Abs(got[row][col]-sample.want[row][col]) > 1e-14 {
					t.Errorf("matrix[%d][%d]=%.17g want=%.17g", row, col, got[row][col], sample.want[row][col])
				}
			}
		}
	}
	for body := SkyMercury; body <= SkyNeptune; body++ {
		orientation, err := BodyOrientationAt(body, 2461319.5)
		if err != nil {
			t.Fatal(err)
		}
		matrix := bodyFixedToICRF(orientation)
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				dot := 0.0
				for k := 0; k < 3; k++ {
					dot += matrix[k][i] * matrix[k][j]
				}
				want := 0.0
				if i == j {
					want = 1
				}
				if math.Abs(dot-want) > 1e-14 {
					t.Errorf("body=%d basis dot(%d,%d)=%.17g", body, i, j, dot)
				}
			}
		}
		x := Vector3{matrix[0][0], matrix[1][0], matrix[2][0]}
		y := Vector3{matrix[0][1], matrix[1][1], matrix[2][1]}
		z := Vector3{matrix[0][2], matrix[1][2], matrix[2][2]}
		if math.Abs(vectorDot(pxp(x, y), z)-1) > 1e-14 {
			t.Errorf("body=%d basis is not right-handed", body)
		}
	}
}

func TestBodyOrientationSyntheticInvalid(t *testing.T) {
	for _, body := range []SkyBody{SkyEarth, SkySun, SkyMoon, SkyBody(-100), SkyBody(100)} {
		got, err := BodyOrientationAt(body, 2451545)
		if !errors.Is(err, ErrInvalidObservationInput) || got != (BodyOrientation{}) {
			t.Errorf("body=%d got=%+v err=%v", body, got, err)
		}
	}
	for _, jdTT := range []float64{math.NaN(), math.Inf(-1), math.Inf(1), math.MaxFloat64} {
		got, err := BodyOrientationAt(SkyMars, jdTT)
		if !errors.Is(err, ErrInvalidObservationInput) || got != (BodyOrientation{}) {
			t.Errorf("jdTT=%v got=%+v err=%v", jdTT, got, err)
		}
	}
}
