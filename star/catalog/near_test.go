package catalog

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/starainrt/astro/basic"
)

func TestNearMatchesBruteForce(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	catalog := newCatalog()
	for i := 0; i < 4000; i++ {
		ra := random.Float64() * 360
		dec := random.Float64()*180 - 90
		if i < 40 {
			// 贴近 RA 0/360 与天极的记录。
			ra = []float64{0, 359.999, 0.001, 180}[i%4]
			dec = []float64{0, 89.9, -89.9, 45}[i/4%4]
		}
		id := fmt.Sprintf("S %d", i)
		catalog.add(Entry{ID: id, Coordinate: parseCoordinate(id, ra, dec, 0, 0, 0, 0, julianYearTime(2000), basic.CoordinateFrameICRS)})
	}
	catalog.finalize()
	queries := [][3]float64{
		{0, 0, 1}, {359.9, 0, 0.5}, {0.05, 60, 3}, {180, 89.5, 2}, {10, -88, 5},
		{270, 30, 20}, {90, -45, 60}, {0, 0, 120}, {359.999, 89.9, 0.2},
		{0, 0, 60}, {30, 20, 65}, {200, -10, 75}, {0, 0, 180}, {10, 10, 200},
	}
	for _, query := range queries {
		got := catalog.Near(query[0], query[1], query[2])
		var want []string
		for _, entry := range catalog.entries {
			if angularDistanceDeg(normalizeRA(query[0]), query[1], entry.Coordinate.RA, entry.Coordinate.Dec) <= query[2]+1e-12 {
				want = append(want, entry.ID)
			}
		}
		ids := make([]string, len(got))
		for i, entry := range got {
			ids[i] = entry.ID
			if i > 0 && angularDistanceDeg(query[0], query[1], got[i-1].Coordinate.RA, got[i-1].Coordinate.Dec) >
				angularDistanceDeg(query[0], query[1], entry.Coordinate.RA, entry.Coordinate.Dec) {
				t.Fatalf("query %v: result not sorted by distance", query)
			}
		}
		sort.Strings(ids)
		sort.Strings(want)
		if fmt.Sprint(ids) != fmt.Sprint(want) {
			t.Fatalf("query %v: got %d, want %d", query, len(ids), len(want))
		}
	}
}
