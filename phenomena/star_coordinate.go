package phenomena

import "github.com/starainrt/astro/basic"

// StarData 内嵌 9100 星表中的一条记录 / one entry from the embedded 9100-star catalog.
type StarData = basic.StarData

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

// StarCoordinateByName 按内嵌星表中文名得到恒星坐标 / returns a star coordinate by Chinese catalog name.
func StarCoordinateByName(name string) (StarCoordinate, error) {
	star, err := basic.StarDataByChinese(name)
	if err != nil {
		return StarCoordinate{}, err
	}
	return StarCoordinateFromStarData(star)
}
