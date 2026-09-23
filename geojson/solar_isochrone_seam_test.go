package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// 契约：等时线整条支路共用同一时刻，跨反经线时必须照样在两侧补出接缝点——否则换日线处会留下缺口。
func TestSolarEclipseIsochroneMeetsAtAntimeridian(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, location)
	level := time.Date(2009, time.July, 22, 11, 30, 0, 0, location) // 03:30 UT，正好跨反经线
	partial, ok := eclipse.SolarEclipsePartialFootprintsNASABulletinSplitK(date,
		eclipse.SolarEclipsePartialFootprintOptions{
			Step:               20 * time.Minute,
			BoundaryPoints:     24,
			DisableRiseSet:     true,
			GreatestTimeValues: []time.Time{level},
		})
	if !ok {
		t.Fatal("expected the 2009-07-22 partial phase")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	collection := decodeCollection(t, data)
	lines := featuresWithRole(collection, "greatest-time-line")
	if len(lines) != 1 {
		t.Fatalf("greatest-time-line features=%d, want 1", len(lines))
	}
	var coordinates [][][]float64
	if err := json.Unmarshal(lines[0].Geometry.Coordinates, &coordinates); err != nil {
		t.Fatalf("decode coordinates: %v", err)
	}
	if len(coordinates) != 2 {
		t.Fatalf("跨反经线的支路应拆成 2 段，得到 %d", len(coordinates))
	}
	east, west := coordinates[0], coordinates[1]
	if len(east) < 2 || len(west) < 2 {
		t.Fatal("段点数过少")
	}
	if last := east[len(east)-1]; math.Abs(last[0]+180) > 1e-9 {
		t.Fatalf("东段末点经度 %v，应为 −180", last[0])
	}
	if first := west[0]; math.Abs(first[0]-180) > 1e-9 {
		t.Fatalf("西段首点经度 %v，应为 +180", first[0])
	}
	if east[len(east)-1][1] != west[0][1] {
		t.Fatalf("接缝两侧纬度不同：%v vs %v", east[len(east)-1][1], west[0][1])
	}
	for index, segment := range coordinates {
		for point := 1; point < len(segment); point++ {
			if math.Abs(segment[point][0]-segment[point-1][0]) > 180 {
				t.Fatalf("段 %d 内仍有跨 180° 的边：%v → %v", index, segment[point-1], segment[point])
			}
		}
	}
}
