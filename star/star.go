package star

import (
	"errors"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

var (
	ERR_STAR_NEVER_RISE = errors.New("ERROR:极夜，星星在今日永远在地平线下！")
	ERR_STAR_NEVER_SET  = errors.New("ERROR:极昼，星星在今日永远在地平线上！")
	// ERR_STAR_NEVER_DOWN deprecated: -- use ERR_STAR_NEVER_SET instead
	ERR_STAR_NEVER_DOWN = ERR_STAR_NEVER_SET
)

func riseSetResult(date time.Time, jde float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_STAR_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_STAR_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	return basic.JDE2DateByZone(jde, date.Location(), true), nil
}

// Constellation 星座中文名 / Chinese constellation name.
//
// ra/dec 为给定时刻的赤经赤纬，单位度；date 作为所属历元使用。
// ra/dec are equatorial coordinates in degrees and date provides the epoch used by the constellation boundaries.
func Constellation(ra, dec float64, date time.Time) string {
	jde := basic.Date2JDE(date.UTC())
	return basic.ConstellationNameZH(ra, dec, jde)
}

// ConstellationCode IAU 星座代码 / IAU constellation code.
//
// ra/dec 为给定时刻的赤经赤纬，单位度；date 作为所属历元使用。
// ra/dec are equatorial coordinates in degrees and date provides the epoch used by the constellation boundaries.
func ConstellationCode(ra, dec float64, date time.Time) string {
	jde := basic.Date2JDE(date.UTC())
	return basic.ConstellationCode(ra, dec, jde)
}

// ConstellationEN 星座英文名 / English constellation name.
//
// ra/dec 为给定时刻的赤经赤纬，单位度；date 作为所属历元使用。
// ra/dec are equatorial coordinates in degrees and date provides the epoch used by the constellation boundaries.
func ConstellationEN(ra, dec float64, date time.Time) string {
	jde := basic.Date2JDE(date.UTC())
	return basic.ConstellationNameEN(ra, dec, jde)
}

// MeanSiderealTime 平恒星时 / mean sidereal time.
//
// 返回 date 对应绝对时刻的格林尼治平恒星时，单位小时。
// Returns Greenwich mean sidereal time at the instant represented by date, in hours.
func MeanSiderealTime(date time.Time) float64 {
	return basic.MeanSiderealTime(basic.Date2JDE(date.UTC()))
}

// ApparentSiderealTime 真恒星时 / apparent sidereal time.
//
// 返回 date 对应绝对时刻的格林尼治真恒星时，单位小时。
// Returns Greenwich apparent sidereal time at the instant represented by date, in hours.
func ApparentSiderealTime(date time.Time) float64 {
	return basic.ApparentSiderealTime(basic.Date2JDE(date.UTC()))
}

// RiseTime 恒星升起时刻 / stellar rise time.
//
// date 取其所在时区的当地日期，返回值保持相同时区；ra/dec 为该日期附近使用的瞬时赤经赤纬，单位度；
// lon/lat 为观测者经纬度，东正西负、北正南负；height 为海拔高度，单位米；aero 为 true 时加入标准大气折射。
// date is interpreted on its local civil day and the result keeps the same time zone. ra/dec are apparent coordinates in degrees;
// lon/lat are east-positive and north-positive, height is observer elevation in meters, and aero enables standard atmospheric refraction.
func RiseTime(date time.Time, ra, dec, lon, lat, height float64, aero bool) (time.Time, error) {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde, err := basic.StarRiseTime(jde, ra, dec, lon, lat, height, timezone, aero)
	return riseSetResult(date, riseJde, err)
}

// DownTime 恒星落下时刻别名 / deprecated stellar set-time alias.
//
// Deprecated: use SetTime instead.
//
// 参数与 SetTime 相同，仅为兼容旧接口保留。
// Same as SetTime and kept only for backward compatibility.
func DownTime(date time.Time, ra, dec, lon, lat, height float64, aero bool) (time.Time, error) {
	return SetTime(date, ra, dec, lon, lat, height, aero)
}

// SetTime 恒星落下时刻 / stellar set time.
//
// 参数与 RiseTime 相同，返回给定当地日期内的落下时刻。
// Uses the same inputs as RiseTime and returns the set time on the corresponding local civil day.
func SetTime(date time.Time, ra, dec, lon, lat, height float64, aero bool) (time.Time, error) {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde, err := basic.StarSetTime(jde, ra, dec, lon, lat, height, timezone, aero)
	return riseSetResult(date, riseJde, err)
}

// HourAngle 恒星时角 / hour angle.
//
// ra 为瞬时赤经，单位度；lon 为观测者经度，东正西负；date 为观测时刻，会读取其时区参与地方时计算。
// ra is the apparent right ascension in degrees; lon is east-positive longitude; date is the observing instant and its zone offset participates in local-time calculations.
func HourAngle(date time.Time, ra, lon float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarHourAngle(jde, ra, lon, timezone)
}

// Azimuth 恒星方位角 / azimuth.
//
// ra/dec 为瞬时赤经赤纬，单位度；lon/lat 为观测者经纬度，东正西负、北正南负；返回值按正北为 0°、向东增加。
// ra/dec are apparent equatorial coordinates in degrees; lon/lat are east-positive and north-positive; azimuth is measured from north toward east.
func Azimuth(date time.Time, ra, dec, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarAzimuth(jde, ra, dec, lon, lat, timezone)
}

// Altitude 恒星高度角 / stellar altitude.
//
// ra/dec 为瞬时赤经赤纬，单位度；lon/lat 为观测者经纬度，东正西负、北正南负；返回值单位度。
// ra/dec are apparent equatorial coordinates in degrees; lon/lat are east-positive and north-positive; the result is in degrees.
func Altitude(date time.Time, ra, dec, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarHeight(jde, ra, dec, lon, lat, timezone)
}

// Zenith 恒星天顶距 / stellar zenith distance.
//
// 参数与 Altitude 相同，返回值为对应时刻的天顶距，单位度。
// Uses the same inputs as Altitude and returns the zenith distance in degrees.
func Zenith(date time.Time, ra, dec, lon, lat float64) float64 {
	return 90 - Altitude(date, ra, dec, lon, lat)
}

// CulminationTime 恒星中天时刻 / culmination time.
//
// date 取其所在时区的当地日期，返回值保持相同时区；ra 为瞬时赤经，单位度；lon 为观测者经度，东正西负。
// date is interpreted on its local civil day and the result keeps the same time zone. ra is the apparent right ascension in degrees and lon is east-positive longitude.
func CulminationTime(date time.Time, ra, lon float64) time.Time {
	jde := basic.Date2JDE(date)
	if jde-math.Floor(jde) < 0.5 {
		jde--
	}
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.StarCulminationTime(jde, ra, lon, timezone) - timezone/24.00
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// InitStarDatabase 初始化恒星数据库 / initializes the embedded star catalog.
func InitStarDatabase() error {
	return basic.LoadStarData()
}

// StarDataByHR 按 HR 编号查询恒星数据 / returns star data by HR number.
func StarDataByHR(hr int) (basic.StarData, error) {
	return basic.StarDataByHR(hr)
}

// StarDataByName 按中文名查询恒星数据 / returns star data by Chinese name.
func StarDataByName(name string) (basic.StarData, error) {
	return basic.StarDataByChinese(name)
}

// TopBrightStars 明亮恒星样本 / bright-star sample list.
//
// 返回按视星等大致由亮到暗排列、视星等约不高于 3 的内置恒星数据。
// Returns the built-in bright-star sample, roughly ordered from brighter to dimmer and limited to stars around magnitude 3 or brighter.
func TopBrightStars() ([]basic.StarData, error) {
	var brightStars = make([]basic.StarData, 0, 170)
	for _, star := range []int{2491, 2326, 5340, 5459, 7001, 1708, 1713, 2943, 472, 2061, 5267, 7557, 1457, 6134, 5056, 2990, 8728, 4853, 7924, 4730, 5460, 3982, 2618, 6527, 4763, 1790, 1791, 3685, 1903, 4731, 8425, 4905, 3207, 4301, 1017, 2693, 6879, 3307, 5191, 6553, 2088, 6217, 2421, 7790, 3485, 2294, 2891, 3748, 617, 7121, 424, 188, 1948, 337, 15, 2004, 5288, 6556, 5563, 8636, 936, 4534, 4819, 7796, 3634, 5793, 1852, 168, 6705, 3165, 3699, 603, 21, 5054, 6241, 5469, 5132, 5440, 5953, 4295, 99, 8308, 6580, 8775, 6378, 4554, 8162, 2827, 7949, 264, 8781, 3734, 911, 5231, 4357, 6175, 1865, 4662, 4621, 7194, 5685, 4057, 2095, 5984, 1956, 553, 4786, 5854, 5235, 403, 5571, 4216, 4798, 1577, 6508, 6859, 2773, 5506, 7525, 6056, 6132, 5531, 5028, 4199, 1899, 6603, 6148, 5776, 6536, 1666, 4656, 98, 6913, 3185, 6212, 6165, 4932, 39, 1829, 1203, 6461, 5897, 8502, 591, 8322, 1165, 7528, 2286, 2890, 7264, 6084, 5944, 5671, 1220, 2845, 4915, 8232, 2553, 915, 8650, 1231, 4757, 6510, 8414, 2473, 3873, 6746, 7235, 1605} {
		info, err := basic.StarDataByHR(star)
		if err != nil {
			return nil, err
		}
		brightStars = append(brightStars, info)
	}
	return brightStars, nil
}
