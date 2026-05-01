package basic

import "sync"

type constellationMeta struct {
	code string
	zh   string
	en   string
}

type constellationBoundary struct {
	code   string
	points []constellationPoint
	minRA  float64
	maxRA  float64
	minDec float64
	maxDec float64
}

var (
	constellationInitOnce sync.Once
	constellationMetas    = [...]constellationMeta{
		{code: "AND", zh: "仙女座", en: "Andromeda"},
		{code: "ANT", zh: "唧筒座", en: "Antlia"},
		{code: "APS", zh: "天燕座", en: "Apus"},
		{code: "AQR", zh: "宝瓶座", en: "Aquarius"},
		{code: "AQL", zh: "天鹰座", en: "Aquila"},
		{code: "ARA", zh: "天坛座", en: "Ara"},
		{code: "ARI", zh: "白羊座", en: "Aries"},
		{code: "AUR", zh: "御夫座", en: "Auriga"},
		{code: "BOO", zh: "牧夫座", en: "Bootes"},
		{code: "CAE", zh: "雕具座", en: "Caelum"},
		{code: "CAM", zh: "鹿豹座", en: "Camelopardalis"},
		{code: "CNC", zh: "巨蟹座", en: "Cancer"},
		{code: "CVN", zh: "猎犬座", en: "Canes Venatici"},
		{code: "CMA", zh: "大犬座", en: "Canis Major"},
		{code: "CMI", zh: "小犬座", en: "Canis Minor"},
		{code: "CAP", zh: "摩羯座", en: "Capricornus"},
		{code: "CAR", zh: "船底座", en: "Carina"},
		{code: "CAS", zh: "仙后座", en: "Cassiopeia"},
		{code: "CEN", zh: "半人马座", en: "Centaurus"},
		{code: "CEP", zh: "仙王座", en: "Cepheus"},
		{code: "CET", zh: "鲸鱼座", en: "Cetus"},
		{code: "CHA", zh: "蝘蜓座", en: "Chamaeleon"},
		{code: "CIR", zh: "圆规座", en: "Circinus"},
		{code: "COL", zh: "天鸽座", en: "Columba"},
		{code: "COM", zh: "后发座", en: "Coma Berenices"},
		{code: "CRA", zh: "南冕座", en: "Corona Australis"},
		{code: "CRB", zh: "北冕座", en: "Corona Borealis"},
		{code: "CRV", zh: "乌鸦座", en: "Corvus"},
		{code: "CRT", zh: "巨爵座", en: "Crater"},
		{code: "CRU", zh: "南十字座", en: "Crux"},
		{code: "CYG", zh: "天鹅座", en: "Cygnus"},
		{code: "DEL", zh: "海豚座", en: "Delphinus"},
		{code: "DOR", zh: "剑鱼座", en: "Dorado"},
		{code: "DRA", zh: "天龙座", en: "Draco"},
		{code: "EQU", zh: "小马座", en: "Equuleus"},
		{code: "ERI", zh: "波江座", en: "Eridanus"},
		{code: "FOR", zh: "天炉座", en: "Fornax"},
		{code: "GEM", zh: "双子座", en: "Gemini"},
		{code: "GRU", zh: "天鹤座", en: "Grus"},
		{code: "HER", zh: "武仙座", en: "Hercules"},
		{code: "HOR", zh: "时钟座", en: "Horologium"},
		{code: "HYA", zh: "长蛇座", en: "Hydra"},
		{code: "HYI", zh: "水蛇座", en: "Hydrus"},
		{code: "IND", zh: "印第安座", en: "Indus"},
		{code: "LAC", zh: "蝎虎座", en: "Lacerta"},
		{code: "LEO", zh: "狮子座", en: "Leo"},
		{code: "LMI", zh: "小狮座", en: "Leo Minor"},
		{code: "LEP", zh: "天兔座", en: "Lepus"},
		{code: "LIB", zh: "天秤座", en: "Libra"},
		{code: "LUP", zh: "豺狼座", en: "Lupus"},
		{code: "LYN", zh: "天猫座", en: "Lynx"},
		{code: "LYR", zh: "天琴座", en: "Lyra"},
		{code: "MEN", zh: "山案座", en: "Mensa"},
		{code: "MIC", zh: "显微镜座", en: "Microscopium"},
		{code: "MON", zh: "麒麟座", en: "Monoceros"},
		{code: "MUS", zh: "苍蝇座", en: "Musca"},
		{code: "NOR", zh: "矩尺座", en: "Norma"},
		{code: "OCT", zh: "南极座", en: "Octans"},
		{code: "OPH", zh: "蛇夫座", en: "Ophiuchus"},
		{code: "ORI", zh: "猎户座", en: "Orion"},
		{code: "PAV", zh: "孔雀座", en: "Pavo"},
		{code: "PEG", zh: "飞马座", en: "Pegasus"},
		{code: "PER", zh: "英仙座", en: "Perseus"},
		{code: "PHE", zh: "凤凰座", en: "Phoenix"},
		{code: "PIC", zh: "绘架座", en: "Pictor"},
		{code: "PSC", zh: "双鱼座", en: "Pisces"},
		{code: "PSA", zh: "南鱼座", en: "Piscis Austrinus"},
		{code: "PUP", zh: "船尾座", en: "Puppis"},
		{code: "PYX", zh: "罗盘座", en: "Pyxis"},
		{code: "RET", zh: "网罟座", en: "Reticulum"},
		{code: "SGE", zh: "天箭座", en: "Sagitta"},
		{code: "SGR", zh: "人马座", en: "Sagittarius"},
		{code: "SCO", zh: "天蝎座", en: "Scorpius"},
		{code: "SCL", zh: "玉夫座", en: "Sculptor"},
		{code: "SCT", zh: "盾牌座", en: "Scutum"},
		{code: "SER1", zh: "巨蛇座", en: "Serpens Caput"},
		{code: "SER2", zh: "巨蛇座", en: "Serpens Cauda"},
		{code: "SEX", zh: "六分仪座", en: "Sextans"},
		{code: "TAU", zh: "金牛座", en: "Taurus"},
		{code: "TEL", zh: "望远镜座", en: "Telescopium"},
		{code: "TRI", zh: "三角座", en: "Triangulum"},
		{code: "TRA", zh: "南三角座", en: "Triangulum Australe"},
		{code: "TUC", zh: "杜鹃座", en: "Tucana"},
		{code: "UMA", zh: "大熊座", en: "Ursa Major"},
		{code: "UMI", zh: "小熊座", en: "Ursa Minor"},
		{code: "VEL", zh: "船帆座", en: "Vela"},
		{code: "VIR", zh: "室女座", en: "Virgo"},
		{code: "VOL", zh: "飞鱼座", en: "Volans"},
		{code: "VUL", zh: "狐狸座", en: "Vulpecula"},
	}
	constellationNameZH     map[string]string
	constellationNameEN     map[string]string
	constellationBoundaries []constellationBoundary
	constellationRayStart   = constellationPoint{RA: 277.5, DEC: -100}
)

func initConstellationData() {
	constellationInitOnce.Do(func() {
		initConstellationPolygons()
		constellationNameZH = make(map[string]string, len(constellationMetas))
		constellationNameEN = make(map[string]string, len(constellationMetas))
		constellationBoundaries = make([]constellationBoundary, 0, len(constellationMetas)-2)
		for _, meta := range constellationMetas {
			constellationNameZH[meta.code] = meta.zh
			constellationNameEN[meta.code] = meta.en
			if meta.code == "UMI" || meta.code == "OCT" {
				continue
			}
			points := constellationPolygons[meta.code]
			if len(points) == 0 {
				continue
			}
			boundary := constellationBoundary{
				code:   meta.code,
				points: points,
				minRA:  points[0].RA,
				maxRA:  points[0].RA,
				minDec: points[0].DEC,
				maxDec: points[0].DEC,
			}
			for _, point := range points[1:] {
				if point.RA < boundary.minRA {
					boundary.minRA = point.RA
				}
				if point.RA > boundary.maxRA {
					boundary.maxRA = point.RA
				}
				if point.DEC < boundary.minDec {
					boundary.minDec = point.DEC
				}
				if point.DEC > boundary.maxDec {
					boundary.maxDec = point.DEC
				}
			}
			constellationBoundaries = append(constellationBoundaries, boundary)
		}
	})
}

// ConstellationCode returns the IAU constellation code / 返回 IAU 星座代码。
func ConstellationCode(ra, dec, jde float64) string {
	return resolveConstellationCode(ra, dec, jde)
}

// ConstellationNameEN returns the English constellation name / 返回英文星座名。
func ConstellationNameEN(ra, dec, jde float64) string {
	return ConstellationNameByCodeEN(ConstellationCode(ra, dec, jde))
}

// ConstellationNameByCodeZH returns the Chinese name for a code / 返回星座代码对应的中文名。
func ConstellationNameByCodeZH(code string) string {
	initConstellationData()
	return constellationNameZH[code]
}

// ConstellationNameByCodeEN returns the English name for a code / 返回星座代码对应的英文名。
func ConstellationNameByCodeEN(code string) string {
	initConstellationData()
	return constellationNameEN[code]
}
