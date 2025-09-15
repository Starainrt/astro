package calendar

// 蜀汉朔日表
func shuCals() map[int]uint32 {
	return map[int]uint32{
		237: 2862623744,
		238: 3042255104,
		239: 2874160128,
		240: 1432365568,
		241: 2863373824,
		242: 1788883200,
		243: 2907710976,
		244: 1432984576,
		245: 1431318016,
		246: 1789764352,
		247: 1520449024,
		248: 2865769216,
		249: 1432100864,
		250: 1431319296,
		251: 1520445440,
		252: 2865929472,
		253: 2863673088,
		254: 1431315712,
		255: 1520998912,
		256: 1453337856,
		257: 2863964416,
		258: 2862625792,
		259: 3041929472,
		260: 1454087936,
		261: 1432367360,
		262: 2862622208,
		263: 3578927872,
	}
}

func shuEras() []Era {
	return []Era{
		{
			Year:              264,
			Emperor:           "魏元帝",
			OtherNianHaoStart: "咸熙",
		},
		{
			Year:              263,
			Emperor:           "蜀后主",
			OtherNianHaoStart: "炎兴",
		},
		{
			Year:    258,
			Emperor: "蜀后主",
			Nianhao: "景耀",
		},
		{
			Year:    238,
			Emperor: "蜀后主",
			Nianhao: "延熙",
		},
		{
			Year:              223,
			Emperor:           "蜀后主",
			OtherNianHaoStart: "建兴",
		},
		{
			Year:    221,
			Emperor: "蜀昭烈帝",
			Nianhao: "章武",
		},
	}
}

func shuEraMap() map[string][][]int {
	return map[string][][]int{
		"炎兴": [][]int{{263, 263}},
		"景耀": [][]int{{258, 263}},
		"延熙": [][]int{{238, 257}},
		"建兴": [][]int{{223, 237}},
		"章武": [][]int{{221, 223}},
	}
}

func wuCals() map[int]uint32 {
	return map[int]uint32{
		223: 1432367360,
		224: 2862622208,
		225: 3578927616,
		226: 2907712768,
		227: 1433281280,
		228: 1431320064,
		229: 1788881408,
		230: 2907971328,
		231: 2865771008,
		232: 1431316480,
		233: 1789565952,
		234: 1520447232,
		235: 2865767424,
		236: 1431378432,
		237: 3578800896,
		238: 1521295616,
		239: 1436562432,
		240: 2862622976,
		241: 3578993920,
		242: 3041931264,
		243: 1436558848,
		244: 2863766272,
		245: 2862624000,
		246: 3042320896,
		247: 2865771776,
		248: 1431317248,
		249: 2863406848,
		250: 1788883456,
		251: 2874156800,
		252: 1431969024,
		253: 1431318272,
		254: 1520444416,
		255: 2874185984,
		256: 2863672320,
		257: 1431642368,
		258: 3041932032,
		259: 1453336832,
		260: 2863898112,
		261: 2862624768,
		262: 3041928448,
		263: 1453955840,
		264: 1432366592,
		265: 2863505920,
		266: 1788884224,
		267: 2907712000,
		268: 1433149440,
		269: 1431319040,
		270: 1788880640,
		271: 2907872256,
		272: 2865770240,
		273: 1431315456,
		274: 1789434112,
		275: 1453337600,
		276: 2866094336,
		277: 2862625536,
		278: 3578800128,
		279: 1454087680,
		280: 1436561664,
	}
}

func wuEraMap() map[string][][]int {
	return map[string][][]int{
		"天玺": [][]int{{276, 276}},
		"天册": [][]int{{275, 276}},
		"凤凰": [][]int{{272, 275}},
		"建衡": [][]int{{269, 271}},
		"宝鼎": [][]int{{266, 269}},
		"甘露": [][]int{{265, 266}},
		"元兴": [][]int{{264, 265}},
		"永安": [][]int{{258, 264}},
		"太平": [][]int{{256, 258}},
		"五凤": [][]int{{254, 256}},
		"建兴": [][]int{{252, 253}},
		"太元": [][]int{{251, 252}},
		"赤乌": [][]int{{238, 251}},
		"嘉禾": [][]int{{232, 238}},
		"黄龙": [][]int{{229, 231}},
		"黄武": [][]int{{222, 229}},
	}
}

func wuEras() []Era {
	return []Era{
		{
			Year:    277,
			Emperor: "吴末帝",
			Nianhao: "天纪",
			Dynasty: "吴",
		},
		{
			Year:              276,
			Emperor:           "吴末帝",
			OtherNianHaoStart: "天玺",
			Dynasty:           "吴",
		},
		{
			Year:              275,
			Emperor:           "吴末帝",
			OtherNianHaoStart: "天册",
			Dynasty:           "吴",
		},
		{
			Year:    272,
			Emperor: "吴末帝",
			Nianhao: "凤凰",
			Dynasty: "吴",
		},
		{
			Year:              269,
			Emperor:           "吴末帝",
			OtherNianHaoStart: "建衡",
			Dynasty:           "吴",
		},
		{
			Year:              266,
			Emperor:           "吴末帝",
			OtherNianHaoStart: "宝鼎",
			Dynasty:           "吴",
		},
		{
			Year:              265,
			Emperor:           "吴末帝",
			OtherNianHaoStart: "甘露",
			Dynasty:           "吴",
		},
		{
			Year:              264,
			Emperor:           "吴末帝",
			OtherNianHaoStart: "元兴",
			Dynasty:           "吴",
		},
		{
			Year:              258,
			Emperor:           "吴景帝",
			OtherNianHaoStart: "永安",
			Dynasty:           "吴",
		},
		{
			Year:              256,
			Emperor:           "吴景帝",
			OtherNianHaoStart: "太平",
			Dynasty:           "吴",
		},
		{
			Year:    254,
			Emperor: "吴景帝",
			Nianhao: "五凤",
			Dynasty: "吴",
		},
		{
			Year:              252,
			Emperor:           "吴景帝",
			OtherNianHaoStart: "建兴",
			Dynasty:           "吴",
		},
		{
			Year:              251,
			Emperor:           "吴大帝",
			OtherNianHaoStart: "太元",
			Dynasty:           "吴",
		},
		{
			Year:              238,
			Emperor:           "吴大帝",
			OtherNianHaoStart: "赤乌",
			Dynasty:           "吴",
		},
		{
			Year:    232,
			Emperor: "吴大帝",
			Nianhao: "嘉禾",
			Dynasty: "吴",
		},
		{
			Year:              229,
			Emperor:           "吴大帝",
			OtherNianHaoStart: "黄龙",
			Dynasty:           "吴",
		},
		{
			Year:    222,
			Emperor: "吴大帝",
			Nianhao: "黄武",
			Dynasty: "吴",
		},
	}
}

func innerSolarToLunarSanGuo(date Time) Time {
	year := date.solarTime.Year()
	month := int(date.solarTime.Month())
	day := date.solarTime.Day()
	if year >= 221 && year <= 263 {
		lyear, lmonth, ganzhiMonth, lday, isLeap, ldesc := rapidLunarHan2Qing(year, month, day, 0, shuCals)
		date.lunars = append(date.lunars, LunarTime{
			solarDate:   date.solarTime,
			year:        lyear,
			month:       lmonth,
			day:         lday,
			leap:        isLeap,
			desc:        ldesc,
			comment:     "",
			ganzhiMonth: ganzhiMonth,
			eras:        innerEras(lyear, shuEras),
		})
	}
	if year >= 222 && year <= 280 {
		lyear, lmonth, ganzhiMonth, lday, isLeap, ldesc := rapidLunarHan2Qing(year, month, day, 0, wuCals)
		date.lunars = append(date.lunars, LunarTime{
			solarDate:   date.solarTime,
			year:        lyear,
			month:       lmonth,
			day:         lday,
			leap:        isLeap,
			desc:        ldesc,
			ganzhiMonth: ganzhiMonth,
			comment:     "",
			eras:        innerEras(lyear, wuEras),
		})
	}
	return date
}
