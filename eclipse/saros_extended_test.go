package eclipse

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestSarosNumberAgainstAllCatalogAnchors(t *testing.T) {
	for phase, anchors := range [][]sarosMagic{solarSarosAnchors[:], lunarSarosAnchors[:]} {
		for index, magic := range anchors {
			anchor := decodeSarosMagic(magic, phase+index)
			k, ok := sarosLunation(basic.JDCalc(int(anchor.Year), int(anchor.Month), float64(anchor.Day)), phase)
			if !ok {
				t.Fatalf("invalid catalog anchor: %+v", anchor)
			}
			for member := 1; member <= int(anchor.Count); member++ {
				got := sarosNumber(k+(member-1)*sarosCycleLunations, phase)
				if got != int(anchor.Series) {
					t.Fatalf("phase=%d series=%d member=%d got series=%d", phase, anchor.Series, member, got)
				}
			}
		}
	}
}

func TestExtendedSarosUsesPrecomputedSeries(t *testing.T) {
	got := ClosestSolarEclipse(time.Date(3288, 11, 15, 12, 0, 0, 0, time.UTC))
	if !got.HasSaros || got.Saros != (SarosInfo{Series: 202, Member: 1, Count: 71}) {
		t.Fatalf("unexpected precomputed solar Saros: has=%v info=%+v", got.HasSaros, got.Saros)
	}
	next := ClosestSolarEclipse(time.Date(3306, 11, 28, 12, 0, 0, 0, time.UTC))
	if !next.HasSaros || next.Saros != (SarosInfo{Series: 202, Member: 2, Count: 71}) {
		t.Fatalf("unexpected second precomputed member: %+v", next.Saros)
	}
	previous := ClosestSolarEclipse(time.Date(3288, 10, 17, 12, 0, 0, 0, time.UTC))
	assertSarosInfo(t, previous.HasSaros, previous.Saros, 164, 67, 80)
}

func TestExtendedSarosSpansAreComplete(t *testing.T) {
	for phase, spans := range [][]sarosSpan{solarSarosExtended[:], lunarSarosExtended[:]} {
		seen := make(map[int]bool)
		for i, span := range spans {
			if span.First > span.Last || (span.Last-span.First)%sarosCycleLunations != 0 || span.Member < 1 || span.Member > span.Count {
				t.Fatalf("phase=%d invalid span: %+v", phase, span)
			}
			if i == 0 || spans[i-1].Series != span.Series {
				if span.Member != 1 || (i > 0 && spans[i-1].Series > span.Series) {
					t.Fatalf("phase=%d invalid series start: %+v", phase, span)
				}
			} else {
				previous := spans[i-1]
				nextMember := previous.Member + (previous.Last-previous.First)/sarosCycleLunations + 1
				if span.First <= previous.Last+sarosCycleLunations || (span.First-previous.First)%sarosCycleLunations != 0 || span.Member != nextMember || span.Count != previous.Count {
					t.Fatalf("phase=%d invalid gap: previous=%+v current=%+v", phase, previous, span)
				}
				for k := previous.Last + sarosCycleLunations; k < span.First; k += sarosCycleLunations {
					if info, ok := matchSarosSpans(spans, k); ok {
						t.Fatalf("phase=%d missing return k=%d matched %+v", phase, k, info)
					}
				}
			}
			lastMember := span.Member + (span.Last-span.First)/sarosCycleLunations
			if (i == len(spans)-1 || spans[i+1].Series != span.Series) && lastMember != span.Count {
				t.Fatalf("phase=%d incomplete series: %+v", phase, span)
			}
			for k := span.First; k <= span.Last; k += sarosCycleLunations {
				if seen[k] || sarosNumber(k, phase) != span.Series {
					t.Fatalf("phase=%d duplicated or misnumbered return: k=%d span=%+v", phase, k, span)
				}
				seen[k] = true
				want := SarosInfo{Series: span.Series, Member: span.Member + (k-span.First)/sarosCycleLunations, Count: span.Count}
				if got, ok := matchSarosSpans(spans, k); !ok || got != want {
					t.Fatalf("phase=%d k=%d got=%+v has=%v want=%+v", phase, k, got, ok, want)
				}
			}
		}
	}
}

func TestExtendedSarosMatchesLiveSeries(t *testing.T) {
	for _, tc := range []struct {
		phase  int
		series int
	}{
		{phase: 0, series: -39},
		{phase: 0, series: 181},
		{phase: 0, series: 202},
		{phase: 0, series: 293},
		{phase: 1, series: -27},
		{phase: 1, series: 0},
		{phase: 1, series: 4}, // This series contains three missing returns.
		{phase: 1, series: 202},
	} {
		t.Run(fmt.Sprintf("phase%d/series%d", tc.phase, tc.series), func(t *testing.T) {
			spans := solarSarosExtended[:]
			if tc.phase == 1 {
				spans = lunarSarosExtended[:]
			}
			found := false
			for _, span := range spans {
				if span.Series != tc.series {
					continue
				}
				found = true
				for _, k := range []int{span.First, span.First + (span.Last-span.First)/sarosCycleLunations/2*sarosCycleLunations, span.Last} {
					want, _ := matchSarosSpans(spans, k)
					got, ok := extrapolateSaros(k, tc.phase)
					if !ok || got != want {
						t.Fatalf("k=%d live=%+v has=%v table=%+v", k, got, ok, want)
					}
					// 锚定序列的总成员数由锚表决定，扩展表只提供成员位置。
					if anchor, anchored := sarosAnchorRangeFor(tc.phase, want.Series); anchored {
						want.Count = anchor.count
					}
					tt, exists, valid := sarosEclipse(k, tc.phase)
					if !exists || !valid {
						t.Fatalf("k=%d missing eclipse: exists=%v valid=%v", k, exists, valid)
					}
					if got, ok := extendedSarosInfo(tt, tc.phase); !ok || got != want {
						t.Fatalf("k=%d tt=%f dispatched=%+v has=%v want=%+v", k, tt, got, ok, want)
					}
				}
			}
			if !found {
				t.Fatal("missing test series")
			}
		})
	}
}

func TestExtendedSarosRangeAndInvalidInput(t *testing.T) {
	pinSarosWindowPolicy(t)
	if sarosExtendedStartTT != timeToTTJDE(time.Date(-3000, 1, 1, 0, 0, 0, 0, time.UTC)) || sarosExtendedEndTT != timeToTTJDE(time.Date(6001, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("precomputed range must include both end years")
	}
	for phase, spans := range [][]sarosSpan{solarSarosExtended[:], lunarSarosExtended[:]} {
		for _, tt := range []float64{sarosExtendedStartTT, sarosExtendedEndTT - 1.0/86400} {
			k, _ := sarosLunation(tt, phase)
			want, wantOK := matchSarosSpans(spans, k)
			got, ok := extendedSarosInfo(tt, phase)
			if got != want || ok != wantOK {
				t.Fatalf("phase=%d boundary lookup got=%+v/%v want=%+v/%v", phase, got, ok, want, wantOK)
			}
		}
		for _, tt := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), math.MaxFloat64} {
			if got, ok := extendedSarosInfo(tt, phase); ok || got != (SarosInfo{}) {
				t.Fatalf("phase=%d invalid epoch returned %+v/%v", phase, got, ok)
			}
		}
	}
	if got, ok := extrapolateSaros(-56340, 1); ok || got != (SarosInfo{}) {
		t.Fatalf("missing lunar return produced metadata: %+v/%v", got, ok)
	}
}

// 直映缓存只复用结果：命中与未命中都必须与未缓存求值逐字段相同，外推结果也必须与
// 逐字重写的未缓存扫描一致。
func TestSarosEclipseMemoMatchesUncachedEvaluation(t *testing.T) {
	for _, phase := range []int{0, 1} {
		for _, k := range []int{-57901, -56340, 15941, 31551, 19004, 34391, 120000} {
			for attempt := 0; attempt < 2; attempt++ {
				maximum, exists, valid := sarosEclipse(k, phase)
				wantMaximum, wantExists, wantValid := sarosEclipseUncached(k, phase)
				if maximum != wantMaximum || exists != wantExists || valid != wantValid {
					t.Fatalf("phase=%d k=%d attempt=%d cached=%v/%v/%v uncached=%v/%v/%v",
						phase, k, attempt, maximum, exists, valid, wantMaximum, wantExists, wantValid)
				}
			}
		}
		for _, k := range []int{-56340, 120000} {
			for attempt := 0; attempt < 2; attempt++ {
				got, ok := extrapolateSaros(k, phase)
				want, wantOK := sarosExtrapolationUncachedForTest(k, phase)
				if ok != wantOK || got != want {
					t.Fatalf("phase=%d k=%d attempt=%d memoized=%+v/%v uncached scan=%+v/%v",
						phase, k, attempt, got, ok, want, wantOK)
				}
			}
		}
	}
}

func sarosExtrapolationUncachedForTest(k, phase int) (SarosInfo, bool) {
	info := SarosInfo{Series: sarosNumber(k, phase)}
	for offset := -sarosExtrapolationWindow; offset <= sarosExtrapolationWindow; offset++ {
		_, exists, valid := sarosEclipseUncached(k+offset*sarosCycleLunations, phase)
		if !valid || (offset == 0 && !exists) {
			return SarosInfo{}, false
		}
		if !exists {
			continue
		}
		if offset == -sarosExtrapolationWindow || offset == sarosExtrapolationWindow {
			return SarosInfo{}, false
		}
		info.Count++
		if offset <= 0 {
			info.Member++
		}
	}
	return info, true
}
