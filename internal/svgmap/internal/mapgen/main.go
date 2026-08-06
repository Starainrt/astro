// Command mapgen 将 Natural Earth 陆地多边形转换为供共享世界地图渲染器使用的紧凑 SVG 路径 /
// Command mapgen converts Natural Earth land polygons into compact SVG paths for the shared world-map renderer.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	naturalEarthURL = "https://raw.githubusercontent.com/nvkelso/natural-earth-vector/v5.1.2/geojson/ne_50m_land.geojson"
	naturalEarthSHA = "e874b27a51d146452be360cafb3cc50c86001074a67d534113e6534682f9826b"
	worldWidth      = 5760
	worldHeight     = 2880
	polarSize       = 2880
)

type featureCollection struct {
	Features []feature `json:"features"`
}

type feature struct {
	Geometry geometry `json:"geometry"`
}

type geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type geoPoint struct {
	lon float64
	lat float64
}

type point struct {
	x int
	y int
}

type projection struct {
	name       string
	hemisphere int
	project    func(geoPoint) point
}

func main() {
	source := flag.String("source", "", "optional local Natural Earth GeoJSON source")
	outputDir := flag.String("output-dir", ".", "generated SVG path destination directory")
	flag.Parse()

	data, err := readSource(*source)
	if err != nil {
		fatal(err)
	}
	if err := verifySource(data); err != nil {
		fatal(err)
	}
	rings, err := decodeRings(data)
	if err != nil {
		fatal(err)
	}
	projections := []projection{
		{name: "land_equirectangular.path", project: projectEquirectangular},
		{name: "land_north_polar.path", hemisphere: 1, project: func(value geoPoint) point { return projectPolar(value, 1) }},
		{name: "land_south_polar.path", hemisphere: -1, project: func(value geoPoint) point { return projectPolar(value, -1) }},
	}
	for _, current := range projections {
		path, ringCount, pointCount := buildPath(rings, current)
		output := filepath.Join(*outputDir, current.name)
		if err := os.WriteFile(output, path, 0644); err != nil {
			fatal(fmt.Errorf("write %s: %w", output, err))
		}
		fmt.Printf("generated %s: %d rings, %d quantized points, %d bytes\n", output, ringCount, pointCount, len(path))
	}
}

func readSource(path string) ([]byte, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		return data, nil
	}
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Get(naturalEarthURL)
	if err != nil {
		return nil, fmt.Errorf("download Natural Earth data: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download Natural Earth data: %s", response.Status)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read Natural Earth response: %w", err)
	}
	return data, nil
}

func verifySource(data []byte) error {
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	if actual != naturalEarthSHA {
		return fmt.Errorf("Natural Earth source checksum = %s, want %s", actual, naturalEarthSHA)
	}
	return nil
}

func decodeRings(data []byte) ([][]geoPoint, error) {
	var collection featureCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("decode Natural Earth GeoJSON: %w", err)
	}
	var rings [][]geoPoint
	appendPolygon := func(polygon [][][]float64) {
		for _, rawRing := range polygon {
			ring := make([]geoPoint, 0, len(rawRing))
			for _, coordinate := range rawRing {
				if len(coordinate) >= 2 {
					ring = append(ring, geoPoint{lon: coordinate[0], lat: coordinate[1]})
				}
			}
			if len(ring) >= 3 {
				rings = append(rings, ring)
			}
		}
	}
	for index, current := range collection.Features {
		switch current.Geometry.Type {
		case "Polygon":
			var polygon [][][]float64
			if err := json.Unmarshal(current.Geometry.Coordinates, &polygon); err != nil {
				return nil, fmt.Errorf("decode feature %d polygon: %w", index, err)
			}
			appendPolygon(polygon)
		case "MultiPolygon":
			var multiPolygon [][][][]float64
			if err := json.Unmarshal(current.Geometry.Coordinates, &multiPolygon); err != nil {
				return nil, fmt.Errorf("decode feature %d multipolygon: %w", index, err)
			}
			for _, polygon := range multiPolygon {
				appendPolygon(polygon)
			}
		default:
			return nil, fmt.Errorf("feature %d has unsupported geometry %q", index, current.Geometry.Type)
		}
	}
	if len(rings) == 0 {
		return nil, fmt.Errorf("Natural Earth source contains no usable rings")
	}
	return rings, nil
}

func buildPath(rings [][]geoPoint, current projection) ([]byte, int, int) {
	path := make([]byte, 0, 256000)
	ringCount := 0
	pointCount := 0
	for _, sourceRing := range rings {
		ring := sourceRing
		if current.hemisphere != 0 {
			ring = clipHemisphere(ring, current.hemisphere)
		}
		projected := projectRing(ring, current)
		var added int
		path, added = appendRing(path, projected)
		if added > 0 {
			ringCount++
			pointCount += added
		}
	}
	return path, ringCount, pointCount
}

func clipHemisphere(points []geoPoint, hemisphere int) []geoPoint {
	if len(points) == 0 {
		return nil
	}
	inside := func(value geoPoint) bool { return value.lat*float64(hemisphere) >= 0 }
	intersection := func(a, b geoPoint) geoPoint {
		dLon := normalizeLongitudeDelta(b.lon - a.lon)
		fraction := -a.lat / (b.lat - a.lat)
		return geoPoint{lon: normalizeLongitude(a.lon + fraction*dLon), lat: 0}
	}
	result := make([]geoPoint, 0, len(points)+2)
	previous := points[len(points)-1]
	previousInside := inside(previous)
	for _, current := range points {
		currentInside := inside(current)
		if currentInside != previousInside {
			result = append(result, intersection(previous, current))
		}
		if currentInside {
			result = append(result, current)
		}
		previous = current
		previousInside = currentInside
	}
	return result
}

func projectRing(ring []geoPoint, current projection) []point {
	if len(ring) < 3 {
		return nil
	}
	if current.hemisphere == 0 {
		result := make([]point, 0, len(ring))
		for _, value := range ring {
			result = appendUniquePoint(result, current.project(value))
		}
		return result
	}

	result := make([]point, 0, len(ring)*2)
	for index, start := range ring {
		end := ring[(index+1)%len(ring)]
		dLon := normalizeLongitudeDelta(end.lon - start.lon)
		dLat := end.lat - start.lat
		steps := int(math.Ceil(math.Max(math.Abs(dLon), math.Abs(dLat)) / 1.5))
		if steps < 1 {
			steps = 1
		}
		for step := 0; step < steps; step++ {
			fraction := float64(step) / float64(steps)
			value := geoPoint{
				lon: normalizeLongitude(start.lon + fraction*dLon),
				lat: start.lat + fraction*dLat,
			}
			result = appendUniquePoint(result, current.project(value))
		}
	}
	return result
}

func appendUniquePoint(points []point, value point) []point {
	if len(points) == 0 || points[len(points)-1] != value {
		return append(points, value)
	}
	return points
}

func appendRing(path []byte, points []point) ([]byte, int) {
	if len(points) > 1 && points[0] == points[len(points)-1] {
		points = points[:len(points)-1]
	}
	if len(points) < 3 {
		return path, 0
	}
	path = append(path, 'M')
	path = appendNumber(path, points[0].x)
	path = appendNumber(path, points[0].y)
	previous := points[0]
	previousCommand := byte(0)
	for _, current := range points[1:] {
		dx := current.x - previous.x
		dy := current.y - previous.y
		command := byte('l')
		if dx == 0 {
			command = 'v'
		} else if dy == 0 {
			command = 'h'
		}
		if command != previousCommand {
			path = append(path, command)
			previousCommand = command
		}
		switch command {
		case 'h':
			path = appendNumber(path, dx)
		case 'v':
			path = appendNumber(path, dy)
		default:
			path = appendNumber(path, dx)
			path = appendNumber(path, dy)
		}
		previous = current
	}
	path = append(path, 'z')
	return path, len(points)
}

func projectEquirectangular(value geoPoint) point {
	return point{
		x: int((value.lon+180)/360*worldWidth + 0.5),
		y: int((90-value.lat)/180*worldHeight + 0.5),
	}
}

func projectPolar(value geoPoint, hemisphere int) point {
	radius := (90 - float64(hemisphere)*value.lat) / 90 * polarSize / 2
	longitude := value.lon * math.Pi / 180
	return point{
		x: int(polarSize/2 + radius*math.Sin(longitude) + 0.5),
		y: int(polarSize/2 - radius*math.Cos(longitude) + 0.5),
	}
}

func normalizeLongitude(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}

func normalizeLongitudeDelta(value float64) float64 {
	return normalizeLongitude(value)
}

func appendNumber(path []byte, value int) []byte {
	if len(path) > 0 && value >= 0 {
		last := path[len(path)-1]
		if last >= '0' && last <= '9' {
			path = append(path, ' ')
		}
	}
	return strconv.AppendInt(path, int64(value), 10)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mapgen:", err)
	os.Exit(1)
}
