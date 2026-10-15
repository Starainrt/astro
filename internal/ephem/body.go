package ephem

import (
	"errors"
	"reflect"
)

// Body NAIF 天体编号 / NAIF body code.
type Body int

const (
	SSB     Body = 0
	Mercury Body = 1
	Venus   Body = 2
	EMB     Body = 3
	Mars    Body = 4
	Jupiter Body = 5
	Saturn  Body = 6
	Uranus  Body = 7
	Neptune Body = 8
	Sun     Body = 10
	Moon    Body = 301
	Earth   Body = 399
)

var (
	// ErrOutOfRange 时刻超出星历覆盖 / epoch outside the ephemeris coverage.
	ErrOutOfRange = errors.New("ephem: epoch outside coverage")
	// ErrUnsupportedBody 星历不含该天体 / body not provided by the ephemeris.
	ErrUnsupportedBody = errors.New("ephem: unsupported body")
	// ErrNoEngine 实例未配置引擎 / instance has no engine.
	ErrNoEngine = errors.New("ephem: instance has no engine")
	// ErrUnsupportedApparent 引擎不支持视位置 / engine does not support apparent positions.
	ErrUnsupportedApparent = errors.New("ephem: apparent positions are unsupported")
)

// Engine 几何星历：ICRF 赤道，km 与 km/日，时间 TDB；只读、可并发 / geometric ephemeris in ICRF, km and km/day, TDB.
type Engine interface {
	State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error)
	Coverage(target Body) (start, end float64, ok bool)
}

// IsNilEngine 引擎为 nil 或持有 nil 指针等空值 / engine is nil or wraps a nil value.
func IsNilEngine(engine Engine) bool {
	if engine == nil {
		return true
	}
	value := reflect.ValueOf(engine)
	switch value.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.Interface:
		return value.IsNil()
	}
	return false
}
