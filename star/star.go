package star

import (
	"errors"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

var (
	ERR_STAR_NEVER_RISE = errors.New("ERROR:极夜，星星在今日永远在地平线下！")
	ERR_STAR_NEVER_DOWN = errors.New("ERROR:极昼，星星在今日永远在地平线上！")
)

// Constellation
// 计算date对应UTC世界时给定Date坐标赤经、赤纬所在的星座
func Constellation(ra, dec float64, date time.Time) string {
	jde := basic.Date2JDE(date.UTC())
	return basic.WhichCst(ra, dec, jde)
}

// MeanSiderealTime UTC 平恒星时
func MeanSiderealTime(date time.Time) float64 {
	return basic.MeanSiderealTime(basic.Date2JDE(date.UTC()))
}

// ApparentSiderealTime UTC真恒星时
func ApparentSiderealTime(date time.Time) float64 {
	return basic.ApparentSiderealTime(basic.Date2JDE(date.UTC()))
}

// RiseTime 星星升起时间
//
//	date, 世界时（忽略此处时区）
//	ra，Date瞬时赤经
//	dec，Date瞬时赤纬
//	lon，经度，东正西负
//	lat，纬度，北正南负
//	height，高度
//	aero,是否进行大气修正
func RiseTime(date time.Time, ra, dec, lon, lat, height float64, aero bool) (time.Time, error) {
	var err error
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde := basic.StarRiseTime(jde, ra, dec, lon, lat, height, timezone, aero)
	if riseJde == -2 {
		err = ERR_STAR_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_STAR_NEVER_DOWN
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// DownTime 星星降落时间
//
//	date, 世界时（忽略此处时区）
//	ra，Date瞬时赤经
//	dec，Date瞬时赤纬
//	lon，经度，东正西负
//	lat，纬度，北正南负
//	height，高度
//	aero,是否进行大气修正
func DownTime(date time.Time, ra, dec, lon, lat, height float64, aero bool) (time.Time, error) {
	var err error
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde := basic.StarDownTime(jde, ra, dec, lon, lat, height, timezone, aero)
	if riseJde == -2 {
		err = ERR_STAR_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_STAR_NEVER_DOWN
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// HourAngle 恒星时角
// 返回给定Date赤经、经度、对应date时区date时刻的太阳时角（
func HourAngle(date time.Time, ra, lon float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarHourAngle(jde, ra, lon, timezone)
}

// Azimuth 恒星方位角
// 返回给定Date赤经赤纬、经纬度、对应date时区date时刻的恒星方位角（正北为0，向东增加）
func Azimuth(date time.Time, ra, dec, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarAzimuth(jde, ra, dec, lon, lat, timezone)
}

// Zenith 恒星高度角
// 返回给定赤经赤纬、经纬度、对应date时区date时刻的太阳高度角
func Zenith(date time.Time, ra, dec, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarHeight(jde, ra, dec, lon, lat, timezone)
}

// CulminationTime 恒星中天时间
// 返回给定赤经赤纬、经纬度、对应date时区date时刻的太阳中天日期
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

// InitStarDatabase 初始化恒星数据库
func InitStarDatabase() error {
	return basic.LoadStarData()
}

// 通过恒星HR编号获取恒星参数
func StarDataByHR(hr int) (basic.StarData, error) {
	return basic.StarDataByHR(hr)
}

// 通过中文名获取恒星参数
func StarDataByName(name string) (basic.StarData, error) {
	return basic.StarDataByChinese(name)
}

// 从亮到暗返回视星等小于3.00的恒星数据
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
