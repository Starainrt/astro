package svg

import (
	"github.com/starainrt/astro/satellite"
	"time"
)

// MapProjection 地面图投影 / ground-map projection.
type MapProjection string

const (
	// MapProjectionAuto 按路径纬度选择投影 / select a projection from path latitudes.
	MapProjectionAuto MapProjection = ""
	// MapProjectionEquirectangular 等经纬投影 / equirectangular projection.
	MapProjectionEquirectangular MapProjection = "equirectangular"
	// MapProjectionNorthPolar 北极方位等距投影 / north-polar azimuthal equidistant projection.
	MapProjectionNorthPolar MapProjection = "north-polar"
	// MapProjectionSouthPolar 南极方位等距投影 / south-polar azimuthal equidistant projection.
	MapProjectionSouthPolar MapProjection = "south-polar"
	// MapProjectionOrthographic 正射投影 / orthographic projection.
	MapProjectionOrthographic MapProjection = "orthographic"
)

// TransitMapSVGOptions 凌现象地面图选项 / transit ground-map options.
type TransitMapSVGOptions struct {
	Width, Height     int
	Language          string
	Location          *time.Location
	Projection        MapProjection
	Nearby            *satellite.NearbyTransit
	Observer          *satellite.Observer
	TimeLabelStep     time.Duration
	Title, FooterNote string
}

// TransitLocalSVGOptions 站心凌现象盘面图选项 / local transit disk-chart options.
type TransitLocalSVGOptions struct {
	Width, Height     int
	Language          string
	Location          *time.Location
	Title, FooterNote string
}

// CoverageSVGOptions 卫星可见范围图选项 / satellite coverage-chart options.
type CoverageSVGOptions struct {
	Width, Height     int
	Language          string
	Location          *time.Location
	Projection        MapProjection
	Highlight         time.Time
	Observer          *satellite.Observer
	TimeLabelStep     time.Duration
	Title, FooterNote string
	Passes            []CoveragePass
}

// CoveragePass 一场预计算的站心过境与可选星等、仰角采样 / a computed station pass with optional magnitude and elevation samples.
type CoveragePass struct {
	satellite.Pass
	Samples []satellite.Look
}
