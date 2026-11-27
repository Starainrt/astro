package basic

import (
	"math"

	"github.com/starainrt/astro/tools"
)

// BodyObserver 行星椭球面上的观测者 / observer on a planetary reference layer.
type BodyObserver struct {
	// Body 观测者所在行星，水星至海王星 / observer's planet, Mercury through Neptune.
	Body SkyBody
	// EastLongitude 行星中心东经，度 / planetocentric east longitude in degrees.
	EastLongitude float64
	// Latitude 椭球法线的大地纬度，度 / planetodetic latitude of the reference ellipsoid normal in degrees.
	Latitude float64
	// Height 相对椭球面的高度，km，只影响站心位置 / height above the reference layer in km, affecting only observer position.
	Height float64
}

type bodyObserverFrame struct {
	positionKM, east, north, up Vector3
	orientation                 BodyOrientation
}

func validateBodyObserver(observer BodyObserver) (planetPhysicalModel, error) {
	model, _, _, _, ok := bodyOrientationModel(observer.Body)
	if !ok || !isFiniteFloat(observer.EastLongitude) || !isFiniteFloat(observer.Latitude) || math.Abs(observer.Latitude) > 90 || !isFiniteFloat(observer.Height) {
		return planetPhysicalModel{}, ErrInvalidObservationInput
	}
	return model, nil
}

func bodyObserverFrameAt(observer BodyObserver, jdTDB float64) (bodyObserverFrame, error) {
	model, err := validateBodyObserver(observer)
	if err != nil {
		return bodyObserverFrame{}, err
	}
	orientation, err := BodyOrientationAt(observer.Body, jdTDB)
	if err != nil {
		return bodyObserverFrame{}, err
	}
	frame := bodyObserverFrameFromModel(observer, model, orientation)
	for _, vector := range []Vector3{frame.positionKM, frame.east, frame.north, frame.up} {
		for _, value := range vector {
			if !isFiniteFloat(value) {
				return bodyObserverFrame{}, ErrInvalidObservationInput
			}
		}
	}
	return frame, nil
}

func bodyObserverFrameFromModel(observer BodyObserver, model planetPhysicalModel, orientation BodyOrientation) bodyObserverFrame {
	sinLon, cosLon := math.Sincos(tools.Limit360(observer.EastLongitude) * math.Pi / 180)
	sinLat, cosLat := math.Sincos(observer.Latitude * math.Pi / 180)
	atPole := math.Abs(observer.Latitude) == 90
	if atPole {
		sinLat, cosLat = math.Copysign(1, observer.Latitude), 0
	}
	ratio := model.polarRadius * model.polarRadius / (model.equatorialRadius * model.equatorialRadius)
	n := model.equatorialRadius / math.Sqrt(cosLat*cosLat+ratio*sinLat*sinLat)
	position := Vector3{(n + observer.Height) * cosLat * cosLon, (n + observer.Height) * cosLat * sinLon, (ratio*n + observer.Height) * sinLat}
	up := Vector3{cosLat * cosLon, cosLat * sinLon, sinLat}
	// 北向投影的解析切向量避免极区的相减损失，极点另用本初子午线。
	north := Vector3{-sinLat * cosLon, -sinLat * sinLon, cosLat}
	matrix := bodyFixedToICRF(orientation)
	frame := bodyObserverFrame{orientation: orientation}
	for row := range matrix {
		for col := range matrix[row] {
			frame.positionKM[row] += matrix[row][col] * position[col]
			frame.up[row] += matrix[row][col] * up[col]
			frame.north[row] += matrix[row][col] * north[col]
		}
	}
	if atPole {
		prime := Vector3{matrix[0][0], matrix[1][0], matrix[2][0]}
		projection := vectorDot(prime, frame.up)
		for i := range frame.north {
			frame.north[i] = prime[i] - projection*frame.up[i]
		}
	}
	frame.north = normalizeVector(frame.north)
	frame.east = normalizeVector(pxp(frame.north, frame.up))
	return frame
}

func convertBodyLongitude(body SkyBody, longitude float64) (float64, error) {
	model, _, _, _, ok := bodyOrientationModel(body)
	if !ok || !isFiniteFloat(longitude) {
		return 0, ErrInvalidObservationInput
	}
	result := tools.Limit360(longitude)
	if !model.positiveEast {
		result = tools.Limit360(360 - result)
	}
	return result, nil
}

func convertBodyLatitude(body SkyBody, latitude float64, toPlanetodetic bool) (float64, error) {
	model, _, _, _, ok := bodyOrientationModel(body)
	if !ok || !isFiniteFloat(latitude) || math.Abs(latitude) > 90 {
		return 0, ErrInvalidObservationInput
	}
	if math.Abs(latitude) == 90 || model.equatorialRadius == model.polarRadius {
		return latitude, nil
	}
	ratio := model.polarRadius * model.polarRadius / (model.equatorialRadius * model.equatorialRadius)
	if toPlanetodetic {
		sinLat, cosLat := math.Sincos(latitude * math.Pi / 180)
		return math.Atan2(sinLat, cosLat*ratio) * 180 / math.Pi, nil
	}
	return math.Atan(math.Tan(latitude*math.Pi/180)*ratio) * 180 / math.Pi, nil
}
