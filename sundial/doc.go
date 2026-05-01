// Package sundial provides apparent-solar-time helpers and planar-sundial
// geometry utilities.
//
// 当前提供两层能力：
//   - 真太阳时换算
//   - 平太阳时换算
//   - 太阳时角
//   - 平太阳时 / 区时对应的时角与时间线采样
//   - 平面日晷通用几何（影尖坐标、日晷中心、极轴晷针）
//   - 平面日晷受光时角区间
//   - 赤纬曲线的分段采样
//   - 赤道 / 水平 / 垂直日晷特例
//   - 水平日晷时线角
//
// 对地方平太阳时时间线采样时，传入的 date 应处于目标地点的地方平太阳时区；
// 对区时时间线采样时，date 负责提供民用日期与时区，原有钟面时间会被目标钟面读数替换。
//
// The package covers apparent solar time, mean solar time, hour-angle
// conversions for mean or zone time, general planar sundial geometry,
// illuminated hour-angle intervals, declination-curve sampling, and a few
// common special cases such as equatorial, horizontal, and vertical dials.
package sundial
