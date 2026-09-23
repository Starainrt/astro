package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// TestLunarGeoJSONDropsAntimeridianSlivers 固定月食可见区的环卫生契约：日界线剪裁会在相邻
// 世界各输出一次"全部顶点落在同一条子午线、平面面积只剩浮点噪声"的零宽薄片（实测
// 1.8e-12 deg²，刚好越过共享剪裁器 1e-12 的零面积阈值），导出时必须在月食这一层丢弃，
// 同时不能连带丢掉有面积的面。
// TestLunarGeoJSONDropsAntimeridianSlivers pins the ring-hygiene contract of the lunar
// visibility export: the antimeridian split emits one zero-width sliver per adjacent world whose
// vertices all sit on one meridian with a planar area of floating-point noise (measured
// 1.8e-12 deg^2, just past the shared splitter's 1e-12 zero-area floor). The lunar export must
// drop it without dropping polygons that do have area.
func TestLunarGeoJSONDropsAntimeridianSlivers(t *testing.T) {
	for _, day := range []string{
		"2022-11-08", "2026-03-03", "4026-03-16", "1904-09-24", "2025-03-14",
		"2028-12-31", "0275-09-22", "0386-09-24", "1076-09-15", "2396-03-25",
	} {
		t.Run(day, func(t *testing.T) {
			date, err := time.Parse("2006-01-02", day)
			if err != nil {
				t.Fatal(err)
			}
			info, ok := eclipse.LunarEclipseOnDate(date)
			if !ok {
				t.Fatal("missing lunar eclipse")
			}
			data, err := geojson.MarshalLunarEclipse(info, 360)
			if err != nil {
				t.Fatal(err)
			}
			collection := decodeCollection(t, data)
			roles := []string{
				"visible-at-p1", "visible-at-p4", "visible-during-eclipse", "visible-throughout-eclipse",
			}
			// 纯半影月食没有本影接触，不导出两条仅见半影带。
			if info.HasPartial {
				roles = append(roles, "penumbra-moonset", "penumbra-moonrise")
			}
			for _, role := range roles {
				feature := featureWithRole(t, collection, role)
				var polygons [][][][]float64
				if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
					t.Fatal(err)
				}
				if len(polygons) == 0 {
					t.Fatalf("%s has no polygons", role)
				}
				for index, polygon := range polygons {
					area := 0.0
					for _, ring := range polygon {
						if len(ring) < 4 {
							t.Fatalf("%s polygon %d ring has %d positions", role, index, len(ring))
						}
						distinct := 0
						for position, point := range ring {
							if position == 0 || math.Abs(point[0]-ring[position-1][0]) > 1e-9 ||
								math.Abs(point[1]-ring[position-1][1]) > 1e-9 {
								distinct++
							}
						}
						if distinct > 1 &&
							math.Abs(ring[0][0]-ring[len(ring)-1][0]) <= 1e-9 &&
							math.Abs(ring[0][1]-ring[len(ring)-1][1]) <= 1e-9 {
							// 闭合点不是独立顶点 / The closing position is not a distinct vertex.
							distinct--
						}
						if distinct < 3 {
							t.Fatalf("%s polygon %d is a zero-width sliver: %v", role, index, ring)
						}
						for position := 0; position+1 < len(ring); position++ {
							area += ring[position][0]*ring[position+1][1] - ring[position+1][0]*ring[position][1]
						}
					}
					if math.Abs(area/2) < 1e-12 {
						t.Fatalf("%s polygon %d has no area: %v", role, index, polygon)
					}
				}
			}
		})
	}
}
