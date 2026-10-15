package geojson

import (
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/moon"
)

// Generator 指定星历实例的计算模块 / computation module bound to an ephemeris instance.
type Generator struct{ inst *ephem.Instance }

// NewWithInstance 创建指定星历实例的模块，nil 使用全局模型 / create a module; nil uses the global model.
func NewWithInstance(inst *ephem.Instance) *Generator { return &Generator{inst: inst} }

var defaultModule Generator

// MarshalPlanetOccultation 将月掩行星的部分掩、全掩和中心线编码为 GeoJSON / encode a planetary occultation path as GeoJSON.
func (module *Generator) MarshalPlanetOccultation(path moon.PlanetOccultationPath) ([]byte, error) {
	return MarshalPlanetOccultation(path)
}

// MarshalPlanetOccultationFootprints 将指定时刻可用的精确行星外切和内切足迹编码为 GeoJSON FeatureCollection；事件外返回空集合 / encode planetary occultation footprints at an instant; empty outside the event.
func (module *Generator) MarshalPlanetOccultationFootprints(instant moon.PlanetOccultationInstant) ([]byte, error) {
	return MarshalPlanetOccultationFootprints(instant)
}

// MarshalPlanetOccultationWithTimeMarkers 编码行星月掩，并沿中心线按固定间隔追加 Point 要素 / encode a planetary occultation with time markers along the centre line.
func (module *Generator) MarshalPlanetOccultationWithTimeMarkers(
	path moon.PlanetOccultationPath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return MarshalPlanetOccultationWithTimeMarkers(path, options)
}

// MarshalStarOccultation 将月掩恒星的全球掩带和中心线编码为 GeoJSON / encode a stellar occultation path as GeoJSON.
func (module *Generator) MarshalStarOccultation(path moon.StarOccultationPath) ([]byte, error) {
	return MarshalStarOccultation(path)
}

// MarshalStarOccultationFootprint 将指定时刻的精确恒星月掩足迹编码为 GeoJSON FeatureCollection；事件外返回空集合 / encode a stellar occultation footprint at an instant; empty outside the event.
func (module *Generator) MarshalStarOccultationFootprint(instant moon.StarOccultationInstant) ([]byte, error) {
	return MarshalStarOccultationFootprint(instant)
}

// MarshalStarOccultationWithTimeMarkers 编码恒星月掩，并沿中心线按固定间隔追加 Point 要素 / encode a stellar occultation with time markers along the centre line.
func (module *Generator) MarshalStarOccultationWithTimeMarkers(
	path moon.StarOccultationPath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return MarshalStarOccultationWithTimeMarkers(path, options)
}
