package basic

import (
	"math"
	"testing"
)

type lunarEclipseBaseline struct {
	name string
	jde  float64

	expectedType LunarEclipseType
	expectedMax  float64
	expectedMag  float64

	expectedPenumbralStart float64
	expectedPenumbralEnd   float64
	expectedPartialStart   float64
	expectedPartialEnd     float64
	expectedTotalStart     float64
	expectedTotalEnd       float64
}

func TestLunarEclipseChauvenetAgainstLegacyBaseline(t *testing.T) {
	// 这些基准值来自历史本地月食基线，
	// 其阴影口径对应当前保留的 Chauvenet 模型。
	testCases := []lunarEclipseBaseline{
		{
			name:                   "2022-11-08 total",
			jde:                    JDECalc(2022, 11, 8),
			expectedType:           LunarEclipseTotal,
			expectedMax:            2459891.9585873615,
			expectedMag:            1.3635170051692678,
			expectedPenumbralStart: 2459891.8346063416,
			expectedPenumbralEnd:   2459892.0826413140,
			expectedPartialStart:   2459891.8820205650,
			expectedPartialEnd:     2459892.0351211606,
			expectedTotalStart:     2459891.9288277230,
			expectedTotalEnd:       2459891.9883249460,
		},
		{
			name:                   "2023-05-05 penumbral",
			jde:                    JDECalc(2023, 5, 5),
			expectedType:           LunarEclipsePenumbral,
			expectedPenumbralStart: 2460070.1342392800,
			expectedPenumbralEnd:   2460070.3159191823,
		},
		{
			name:                   "2023-10-28 partial",
			jde:                    JDECalc(2023, 10, 28),
			expectedType:           LunarEclipsePartial,
			expectedMax:            2460246.3439460830,
			expectedMag:            0.12723850274626405,
			expectedPenumbralStart: 2460246.2507697106,
			expectedPenumbralEnd:   2460246.4370874465,
			expectedPartialStart:   2460246.3164327650,
			expectedPartialEnd:     2460246.3713359070,
		},
		{
			name:                   "2024-03-25 penumbral",
			jde:                    JDECalc(2024, 3, 25),
			expectedType:           LunarEclipsePenumbral,
			expectedPenumbralStart: 2460394.7028870000,
			expectedPenumbralEnd:   2460394.8999071894,
		},
		{
			name:                   "2024-09-18 partial",
			jde:                    JDECalc(2024, 9, 18),
			expectedType:           LunarEclipsePartial,
			expectedMax:            2460571.6148748010,
			expectedMag:            0.09042791952817894,
			expectedPenumbralStart: 2460571.5281155687,
			expectedPenumbralEnd:   2460571.7016473800,
			expectedPartialStart:   2460571.5923644140,
			expectedPartialEnd:     2460571.6374154520,
		},
		{
			name:                   "2025-03-14 total",
			jde:                    JDECalc(2025, 3, 14),
			expectedType:           LunarEclipseTotal,
			expectedMax:            2460748.7916214615,
			expectedMag:            1.1828107517800281,
			expectedPenumbralStart: 2460748.6645233813,
			expectedPenumbralEnd:   2460748.9187600957,
			expectedPartialStart:   2460748.7156107454,
			expectedPartialEnd:     2460748.8676076555,
			expectedTotalStart:     2460748.7685903380,
			expectedTotalEnd:       2460748.8146345600,
		},
		{
			name:                   "2025-09-07 total",
			jde:                    JDECalc(2025, 9, 7),
			expectedType:           LunarEclipseTotal,
			expectedMax:            2460926.2590034613,
			expectedMag:            1.3672329695760280,
			expectedPenumbralStart: 2460926.1445036167,
			expectedPenumbralEnd:   2460926.3734498024,
			expectedPartialStart:   2460926.1860739910,
			expectedPartialEnd:     2460926.3319619163,
			expectedTotalStart:     2460926.2302397094,
			expectedTotalEnd:       2460926.2877871464,
		},
		{
			name:                   "2026-03-03 total",
			jde:                    JDECalc(2026, 3, 3),
			expectedType:           LunarEclipseTotal,
			expectedMax:            2461102.9825476190,
			expectedMag:            1.1556387222746651,
			expectedPenumbralStart: 2461102.8638708987,
			expectedPenumbralEnd:   2461103.1012840020,
			expectedPartialStart:   2461102.9103626400,
			expectedPartialEnd:     2461103.0546894810,
			expectedTotalStart:     2461102.9619182530,
			expectedTotalEnd:       2461103.0031447060,
		},
	}

	const timeTolerance = 1e-6
	const magnitudeTolerance = 2e-5

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := LunarEclipseChauvenet(tc.jde)

			if result.Type != tc.expectedType {
				t.Fatalf("Type mismatch: got %s want %s", result.Type, tc.expectedType)
			}

			if tc.expectedMax != 0 && math.Abs(result.Maximum-tc.expectedMax) > timeTolerance {
				t.Fatalf("Maximum mismatch: got %.12f want %.12f", result.Maximum, tc.expectedMax)
			}
			if tc.expectedMag != 0 && math.Abs(result.Magnitude-tc.expectedMag) > magnitudeTolerance {
				t.Fatalf("Magnitude mismatch: got %.12f want %.12f", result.Magnitude, tc.expectedMag)
			}

			assertCloseJD(t, "PenumbralStart", result.PenumbralStart, tc.expectedPenumbralStart, timeTolerance)
			assertCloseJD(t, "PenumbralEnd", result.PenumbralEnd, tc.expectedPenumbralEnd, timeTolerance)
			assertCloseJD(t, "PartialStart", result.PartialStart, tc.expectedPartialStart, timeTolerance)
			assertCloseJD(t, "PartialEnd", result.PartialEnd, tc.expectedPartialEnd, timeTolerance)
			assertCloseJD(t, "TotalStart", result.TotalStart, tc.expectedTotalStart, timeTolerance)
			assertCloseJD(t, "TotalEnd", result.TotalEnd, tc.expectedTotalEnd, timeTolerance)

			if result.HasTotal && !(result.TotalStart < result.Maximum && result.Maximum < result.TotalEnd) {
				t.Fatalf("total contact order invalid: start=%.12f max=%.12f end=%.12f", result.TotalStart, result.Maximum, result.TotalEnd)
			}
			if result.HasPartial && !(result.PartialStart < result.Maximum && result.Maximum < result.PartialEnd) {
				t.Fatalf("partial contact order invalid: start=%.12f max=%.12f end=%.12f", result.PartialStart, result.Maximum, result.PartialEnd)
			}
			if result.HasPenumbral && !(result.PenumbralStart < result.Maximum && result.Maximum < result.PenumbralEnd) {
				t.Fatalf("penumbral contact order invalid: start=%.12f max=%.12f end=%.12f", result.PenumbralStart, result.Maximum, result.PenumbralEnd)
			}
		})
	}
}

func TestLunarEclipseDefaultUsesDanjon(t *testing.T) {
	jde := JDECalc(2025, 3, 14)
	defaultResult := LunarEclipse(jde)
	danjonResult := LunarEclipseDanjon(jde)
	chauvenetResult := LunarEclipseChauvenet(jde)

	assertCloseJD(t, "Maximum", defaultResult.Maximum, danjonResult.Maximum, 1e-12)
	assertCloseJD(t, "PenumbralStart", defaultResult.PenumbralStart, danjonResult.PenumbralStart, 1e-12)
	assertCloseJD(t, "PenumbralEnd", defaultResult.PenumbralEnd, danjonResult.PenumbralEnd, 1e-12)
	if math.Abs(defaultResult.PenumbralMagnitude-danjonResult.PenumbralMagnitude) > 1e-12 {
		t.Fatalf("default penumbral magnitude mismatch: got %.12f want %.12f", defaultResult.PenumbralMagnitude, danjonResult.PenumbralMagnitude)
	}
	if math.Abs(defaultResult.PenumbralMagnitude-chauvenetResult.PenumbralMagnitude) < 1e-4 {
		t.Fatalf("default model should not collapse to Chauvenet: default=%.12f chauvenet=%.12f", defaultResult.PenumbralMagnitude, chauvenetResult.PenumbralMagnitude)
	}
}

func TestPenumbralLunarEclipseKeepsNegativeUmbralMagnitude(t *testing.T) {
	testCases := []struct {
		name string
		jde  float64
		calc func(float64) LunarEclipseResult
	}{
		{name: "default 2024-03-25", jde: JDECalc(2024, 3, 25), calc: LunarEclipse},
		{name: "danjon 2024-03-25", jde: JDECalc(2024, 3, 25), calc: LunarEclipseDanjon},
		{name: "chauvenet 2023-05-05", jde: JDECalc(2023, 5, 5), calc: LunarEclipseChauvenet},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.calc(tc.jde)
			if result.Type != LunarEclipsePenumbral {
				t.Fatalf("type mismatch: got %s want %s", result.Type, LunarEclipsePenumbral)
			}
			if !result.HasPenumbral || result.HasPartial || result.HasTotal {
				t.Fatalf("unexpected eclipse flags: %+v", result)
			}
			if !(result.Magnitude < 0) {
				t.Fatalf("expected negative umbral magnitude for penumbral eclipse, got %.12f", result.Magnitude)
			}
			if !(result.PenumbralMagnitude > 0) {
				t.Fatalf("expected positive penumbral magnitude, got %.12f", result.PenumbralMagnitude)
			}
		})
	}
}

func TestLunarEclipseDanjonMagnitudesCloserToNASA(t *testing.T) {
	testCases := []struct {
		name                   string
		jde                    float64
		expectedType           LunarEclipseType
		nasaPenumbralMagnitude float64
		nasaUmbralMagnitude    float64
	}{
		{
			name:                   "2023-10-28 partial",
			jde:                    JDECalc(2023, 10, 28),
			expectedType:           LunarEclipsePartial,
			nasaPenumbralMagnitude: 1.1181,
			nasaUmbralMagnitude:    0.1220,
		},
		{
			name:                   "2025-03-14 total",
			jde:                    JDECalc(2025, 3, 14),
			expectedType:           LunarEclipseTotal,
			nasaPenumbralMagnitude: 2.2595,
			nasaUmbralMagnitude:    1.1784,
		},
		{
			name:                   "2026-03-03 total",
			jde:                    JDECalc(2026, 3, 3),
			expectedType:           LunarEclipseTotal,
			nasaPenumbralMagnitude: 2.1838,
			nasaUmbralMagnitude:    1.1507,
		},
		{
			name:                   "2026-08-28 partial",
			jde:                    JDECalc(2026, 8, 28),
			expectedType:           LunarEclipsePartial,
			nasaPenumbralMagnitude: 1.9645,
			nasaUmbralMagnitude:    0.9299,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			danjonResult := LunarEclipseDanjon(tc.jde)
			chauvenetResult := LunarEclipseChauvenet(tc.jde)

			if danjonResult.Type != tc.expectedType {
				t.Fatalf("Danjon type mismatch: got %s want %s", danjonResult.Type, tc.expectedType)
			}
			if chauvenetResult.Type != tc.expectedType {
				t.Fatalf("Chauvenet type mismatch: got %s want %s", chauvenetResult.Type, tc.expectedType)
			}

			danjonPenumbralError := math.Abs(danjonResult.PenumbralMagnitude - tc.nasaPenumbralMagnitude)
			chauvenetPenumbralError := math.Abs(chauvenetResult.PenumbralMagnitude - tc.nasaPenumbralMagnitude)
			if !(danjonPenumbralError < chauvenetPenumbralError) {
				t.Fatalf("Danjon penumbral magnitude should be closer to NASA: danjon=%.6f chauvenet=%.6f nasa=%.6f", danjonResult.PenumbralMagnitude, chauvenetResult.PenumbralMagnitude, tc.nasaPenumbralMagnitude)
			}

			danjonUmbralError := math.Abs(danjonResult.Magnitude - tc.nasaUmbralMagnitude)
			chauvenetUmbralError := math.Abs(chauvenetResult.Magnitude - tc.nasaUmbralMagnitude)
			if !(danjonUmbralError < chauvenetUmbralError) {
				t.Fatalf("Danjon umbral magnitude should be closer to NASA: danjon=%.6f chauvenet=%.6f nasa=%.6f", danjonResult.Magnitude, chauvenetResult.Magnitude, tc.nasaUmbralMagnitude)
			}
		})
	}
}

func TestLunarEclipseNoEvent(t *testing.T) {
	testCases := []struct {
		name string
		calc func(float64) LunarEclipseResult
	}{
		{name: "default", calc: LunarEclipse},
		{name: "danjon", calc: LunarEclipseDanjon},
		{name: "chauvenet", calc: LunarEclipseChauvenet},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.calc(JDECalc(2023, 6, 4))
			if result.Type != LunarEclipseNone {
				t.Fatalf("Type mismatch: got %s want %s", result.Type, LunarEclipseNone)
			}
			if result.HasPenumbral || result.HasPartial || result.HasTotal {
				t.Fatalf("unexpected contacts: %+v", result)
			}
			if result.PenumbralStart != 0 || result.PenumbralEnd != 0 || result.PartialStart != 0 || result.PartialEnd != 0 || result.TotalStart != 0 || result.TotalEnd != 0 {
				t.Fatalf("expected no contact times, got %+v", result)
			}
			if result.Magnitude != 0 || result.PenumbralMagnitude != 0 {
				t.Fatalf("expected zero magnitudes for non-eclipse, got %+v", result)
			}
		})
	}
}

func assertCloseJD(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if want == 0 {
		if got != 0 {
			t.Fatalf("%s mismatch: got %.12f want 0", name, got)
		}
		return
	}
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.12f want %.12f", name, got, want)
	}
}
