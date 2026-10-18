package phenomena

import "github.com/starainrt/astro/basic"

// StarData 内嵌 9100 星表中的一条记录 / one entry from the embedded 9100-star catalog.
type StarData = basic.StarData

// ErrStarNotFound 未找到匹配的恒星 / no matching star.
var ErrStarNotFound = basic.ErrStarNotFound

// ErrStarNameAmbiguous 星名对应多颗恒星 / the star name matches more than one star.
var ErrStarNameAmbiguous = basic.ErrStarNameAmbiguous

// StarCoordinateFromStarData 将星表记录转换为恒星坐标 / converts a catalog entry to a star coordinate.
func StarCoordinateFromStarData(star StarData) (StarCoordinate, error) {
	return basic.StarCoordinateFromStarData(star)
}

// StarCoordinateByHR 按 HR 编号从内嵌星表得到恒星坐标 / returns a star coordinate by HR number.
func StarCoordinateByHR(hr int) (StarCoordinate, error) {
	star, err := basic.StarDataByHR(hr)
	if err != nil {
		return StarCoordinate{}, err
	}
	return StarCoordinateFromStarData(star)
}

// StarCoordinateByName 按内嵌星表中文名、别名或英文常用名得到唯一恒星坐标；同名多星返回 ErrStarNameAmbiguous / returns the unique star coordinate by Chinese name, alias or English common name, or ErrStarNameAmbiguous for duplicate names.
func StarCoordinateByName(name string) (StarCoordinate, error) {
	star, err := basic.StarDataByName(name)
	if err != nil {
		return StarCoordinate{}, err
	}
	return StarCoordinateFromStarData(star)
}
