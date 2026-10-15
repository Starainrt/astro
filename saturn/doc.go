// Package saturn 土星位置、升落、合冲、留、方照、相位、视直径、物理星历与土星环观测参数；角度单位为度，环长短轴单位为角秒。
// Package saturn covers Saturn position, rise/set, conjunction, opposition, station, quadrature, phase, diameter, physical ephemeris and ring parameters; angles are degrees and ring axes are arcseconds.
// 求值接口的 N 后缀是 n 变体：n < 0 取完整级数；半解析模型下所有 n >= 0 使用同一张快速表；Legacy 按 n 截断。
// The N suffix marks an n-variant: n < 0 uses the full series, every n >= 0 uses one fast table under the semianalytic model, and Legacy truncates by n.
// 时间参数：位置、距离、星等与物理量接口的 date 是绝对时刻，其所在时区参与地方时换算；升落、中天、晨昏与 Next/Last 事件接口按 date 所在时区的当地民用日求解，返回时刻保持该时区。
// Time arguments: date is an absolute instant for position, distance, magnitude and physical-quantity entry points, and its zone takes part in local-time conversion; rise/set, culmination, twilight and Next/Last event entry points solve for the local civil day in date's zone and return instants keeping that zone.
package saturn
