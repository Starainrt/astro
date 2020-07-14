package star

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// Constellation
// 计算date对应UTC世界时给定Date坐标赤经、赤纬所在的星座

func Constellation(ra, dec float64, date time.Time) string {
	jde := basic.Date2JDE(date.UTC())
	return basic.WhichCst(ra, dec, jde)
}
