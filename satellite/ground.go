package satellite

import (
	"errors"
	"math"
	"time"
)

// ErrInvalidInput 输入无效 / invalid input.
var ErrInvalidInput = errors.New("invalid satellite input")

// GroundPoint 卫星的WGS84星下点与地固系速度 / WGS84 subsatellite point and Earth-fixed speed.
type GroundPoint struct {
	Time      time.Time
	Longitude float64 // Longitude 东经，(−180, 180]度 / east longitude in degrees.
	Latitude  float64 // Latitude WGS84大地纬度，度 / WGS84 geodetic latitude in degrees.
	HeightKM  float64
	SpeedKMS  float64 // SpeedKMS 地固系速度模 / Earth-fixed velocity magnitude.
}

// GroundPointAt 计算星下点，保留输入时区 / compute a subsatellite point preserving the input time zone.
func GroundPointAt(e Elements, t time.Time) (GroundPoint, error) {
	o, err := newSatelliteOrbit(e)
	if err != nil {
		return GroundPoint{}, err
	}
	return o.groundAt(t)
}

func (o *satelliteOrbit) groundAt(t time.Time) (GroundPoint, error) {
	s, err := o.stateAt(t)
	if err != nil {
		return GroundPoint{}, err
	}
	r, v := earthFixed(s)
	lon, lat, h := geographicPosition(r)
	return GroundPoint{Time: t, Longitude: lon, Latitude: lat, HeightKM: h, SpeedKMS: vectorNorm(v)}, nil
}

func geographicPosition(r [3]float64) (lon, lat, height float64) {
	e2 := earthFlattening * (2 - earthFlattening)
	p := math.Hypot(r[0], r[1])
	lon = math.Atan2(r[1], r[0]) / degreesToRadians
	if lon <= -180 {
		lon = 180
	}
	lat = math.Atan2(r[2], p*(1-e2))
	for i := 0; i < 10; i++ {
		s := math.Sin(lat)
		n := earthRadiusKM / math.Sqrt(1-e2*s*s)
		next := math.Atan2(r[2]+e2*n*s, p)
		delta := math.Abs(next - lat)
		lat = next
		if delta < 1e-12 {
			break
		}
	}
	sin, cos := math.Sincos(lat)
	n := earthRadiusKM / math.Sqrt(1-e2*sin*sin)
	if math.Abs(cos) > 0.5 {
		height = p/cos - n
	} else {
		height = r[2]/sin - n*(1-e2)
	}
	lat /= degreesToRadians
	return
}

// GroundTrack 按正步长采样闭区间，含首尾且至多100000点 / sample a closed interval including both ends, up to 100000 points.
func GroundTrack(e Elements, start, end time.Time, step time.Duration) ([]GroundPoint, error) {
	span := end.Sub(start)
	if start.IsZero() || end.IsZero() || end.Before(start) || step <= 0 || !start.Add(span).Equal(end) {
		return nil, ErrInvalidInput
	}
	if span/step >= 100000 {
		return nil, ErrInvalidInput
	}
	count := int64(span/step) + 1
	if span%step != 0 {
		count++
	}
	if count > 100000 {
		return nil, ErrInvalidInput
	}
	o, err := newSatelliteOrbit(e)
	if err != nil {
		return nil, err
	}
	points := make([]GroundPoint, 0, int(count))
	for t := start; ; {
		p, err := o.groundAt(t)
		if err != nil {
			return nil, err
		}
		points = append(points, p)
		if t.Equal(end) {
			break
		}
		if end.Sub(t) <= step {
			t = end
		} else {
			t = t.Add(step)
		}
	}
	return points, nil
}
