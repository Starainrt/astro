package svg

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/svgmap"
)

type projectedFillCase struct {
	name       string
	date       time.Time
	pole       float64
	projection geodata.Projection
}

// 偏食可见域的填充必须与球面并集同域：极冠环在等经纬图上沿地图上、下边缘闭合、
// 接缝两侧各贴自己那一侧的边缘，在正射球面图上沿视界圆盘闭合；
// 把窗口两端折到同侧会让闭合边横穿整幅图并丢掉极冠。
func TestSolarEclipseProjectedPartialFillMatchesUnion(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	cases := make([]projectedFillCase, 0, 6)
	for _, item := range []projectedFillCase{
		{"2012-05-21 北极冠", time.Date(2012, 5, 21, 12, 0, 0, 0, cst), 90, ""},
		{"2021-12-04 南极冠", time.Date(2021, 12, 4, 12, 0, 0, 0, cst), -90, ""},
		{"2035-09-02 无冠", time.Date(2035, 9, 2, 12, 0, 0, 0, cst), 0, ""},
	} {
		for _, projection := range []geodata.Projection{
			geodata.ProjectionEquirectangular, geodata.ProjectionOrthographic,
		} {
			item := item
			item.projection = projection
			cases = append(cases, item)
		}
	}
	for _, item := range cases {
		item := item
		t.Run(item.name+"/"+string(item.projection), func(t *testing.T) {
			info, ok := eclipse.SolarEclipsePartialFootprints(item.date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 5 * time.Minute, BoundaryPoints: 180,
			})
			if !ok {
				t.Fatal("missing partial footprints")
			}
			polygons, ok := solarEclipsePartialBandPolygons(info)
			if !ok || len(polygons) == 0 {
				t.Fatal("missing partial-band union")
			}
			options := SolarEclipseMapSVGOptions{Width: 1200, Height: 800, Location: cst}
			layout := solarEclipseMapLayoutFor(options, item.projection,
				geodata.GeoPoint{Longitude: info.Eclipse.GreatestLongitude, Latitude: info.Eclipse.GreatestLatitude})
			frame := layout.frame
			rings := projectedPartialRings(t, polygons, frame)

			if item.projection == geodata.ProjectionEquirectangular {
				for _, ring := range rings {
					for index := range ring {
						point, next := ring[index], ring[(index+1)%len(ring)]
						if math.Abs(next[0]-point[0]) <= frame.Width/2 || math.Abs(next[1]-point[1]) >= 1 {
							continue
						}
						if frame.Y-point[1] > 1 && point[1]-(frame.Y+frame.Height) > 1 {
							t.Fatalf("填充边横穿整幅图：y=%.3f x=%.3f→%.3f", point[1], point[0], next[0])
						}
					}
				}
			}

			projected := func(longitude, latitude float64) (float64, float64, bool) {
				x, y, ok := frame.Project(longitude, latitude)
				if !ok {
					return 0, 0, false
				}
				if item.projection != geodata.ProjectionOrthographic {
					return x, y, true
				}
				radius := math.Min(frame.Width, frame.Height) / 2
				if math.Hypot(x-(frame.X+frame.Width/2), y-(frame.Y+frame.Height/2)) > radius+1 {
					return 0, 0, false
				}
				return x, y, true
			}

			polarSamples := 0
			if item.pole != 0 {
				latitude := math.Copysign(89.5, item.pole)
				for longitude := -180.0; longitude < 180; longitude += 5 {
					point := geodata.GeoPoint{Longitude: longitude, Latitude: latitude}
					if !geodata.SphericalPolygonsContainPoints(polygons, []geodata.GeoPoint{point})[0] {
						continue
					}
					x, y, ok := projected(point.Longitude, point.Latitude)
					if !ok {
						continue
					}
					if !planeFillContains(rings, x, y) {
						t.Fatalf("纬度 %.1f 经度 %.1f 在球面并集内但未被填充", latitude, longitude)
					}
					polarSamples++
				}
				if polarSamples < 48 {
					t.Fatalf("极冠样本只有 %d 个", polarSamples)
				}
			}

			checked, skipped := 0, 0
			for latitude := -85.0; latitude <= 85; latitude += 5 {
				for longitude := -180.0; longitude < 180; longitude += 10 {
					point := geodata.GeoPoint{Longitude: longitude, Latitude: latitude}
					x, y, ok := projected(point.Longitude, point.Latitude)
					if !ok {
						continue
					}
					if planeFillBoundaryDistance(rings, x, y) < 3 {
						skipped++
						continue
					}
					spherical := geodata.SphericalPolygonsContainPoints(polygons, []geodata.GeoPoint{point})[0]
					if plane := planeFillContains(rings, x, y); plane != spherical {
						t.Fatalf("%.1f %.1f 填充=%v 球面并集=%v", longitude, latitude, plane, spherical)
					}
					checked++
				}
			}
			if checked < 400 || skipped > checked/4 {
				t.Fatalf("一致性样本 checked=%d skipped=%d", checked, skipped)
			}
		})
	}
}

func projectedPartialRings(t *testing.T, polygons [][]geodata.GeoPoint, frame svgmap.Frame) [][][2]float64 {
	t.Helper()
	var rings [][][2]float64
	for _, polygon := range polygons {
		points := make([]geodata.GeoPoint, len(polygon))
		copy(points, polygon)
		for _, fragment := range svgmap.PolygonFragments(points, frame.Clip()) {
			ring := make([][2]float64, 0, len(fragment))
			for _, point := range fragment {
				x, y, ok := frame.Project(point.Longitude, point.Latitude)
				if !ok {
					ring = nil
					break
				}
				ring = append(ring, [2]float64{x, y})
			}
			if len(ring) >= 3 {
				rings = append(rings, ring)
			}
		}
	}
	if len(rings) == 0 {
		t.Fatal("no projected fragment")
	}
	return rings
}

func planeFillContains(rings [][][2]float64, x, y float64) bool {
	winding := 0
	for _, ring := range rings {
		for index := range ring {
			first, second := ring[index], ring[(index+1)%len(ring)]
			if first[1] <= y {
				if second[1] > y && (second[0]-first[0])*(y-first[1])-(x-first[0])*(second[1]-first[1]) > 0 {
					winding++
				}
			} else if second[1] <= y && (second[0]-first[0])*(y-first[1])-(x-first[0])*(second[1]-first[1]) < 0 {
				winding--
			}
		}
	}
	return winding != 0
}

func planeFillBoundaryDistance(rings [][][2]float64, x, y float64) float64 {
	distance := math.Inf(1)
	for _, ring := range rings {
		for index := range ring {
			first, second := ring[index], ring[(index+1)%len(ring)]
			distance = math.Min(distance, planeSegmentDistance(x, y, first, second))
		}
	}
	return distance
}

func planeSegmentDistance(x, y float64, first, second [2]float64) float64 {
	dx, dy := second[0]-first[0], second[1]-first[1]
	length := dx*dx + dy*dy
	fraction := 0.0
	if length > 0 {
		fraction = math.Max(0, math.Min(1, ((x-first[0])*dx+(y-first[1])*dy)/length))
	}
	return math.Hypot(x-(first[0]+fraction*dx), y-(first[1]+fraction*dy))
}
