package calendar

func nanMingCals() map[int]uint32 {
	return map[int]uint32{
		1645: 1232804864,
		1646: 1232088832,
		1647: 1689265152,
		1648: 1783732480,
		1649: 3662686720,
		1650: 1800249344,
		1651: 1456485120,
		1652: 719333632,
		1653: 2465438976,
		1654: 2464165888,
		1655: 3378521344,
		1656: 3567950336,
		1657: 3567266816,
		1658: 3662684416,
		1659: 1520998144,
		1660: 1453337088,
		1661: 2799508992,
		1662: 634401280,
		1663: 2463115008,
		1664: 2841320448,
		1665: 2840604160,
		1666: 3030393600,
		1667: 3042056192,
		1668: 2907712256,
		1669: 1437474816,
		1670: 1269838592,
		1671: 632301824,
		1672: 1388060160,
		1673: 1387278336,
		1674: 1689265408,
		1675: 1957370368,
		1676: 1788882176,
		1677: 2907709696,
		1678: 1302927104,
		1679: 1264593408,
		1680: 2775916288,
		1681: 2766156032,
		1682: 3529516544,
		1683: 3912440576,
	}
}

func nanMingEras01() []Era {
	return []Era{
		{
			Year:              1646,
			Emperor:           "南明鲁王",
			OtherNianHaoStart: "鲁王监国",
			Dynasty:           "南明",
		},
		{
			Year:              1645,
			Emperor:           "南明隆武帝",
			OtherNianHaoStart: "隆武",
			Dynasty:           "南明",
		}, {
			Year:    1645,
			Emperor: "南明弘光帝",
			Nianhao: "弘光",
			Dynasty: "南明",
		},
		{
			Year:    1628,
			Emperor: "明思宗",
			Nianhao: "崇祯",
			Dynasty: "明",
		},
	}
}

func nanMingEras02() []Era {
	return []Era{
		{
			Year:              1647,
			Emperor:           "南明/明郑",
			OtherNianHaoStart: "永历",
			Dynasty:           "南明",
		},
	}
}

func nanMingEraMap() map[string][][]int {
	return map[string][][]int{
		"永历":   [][]int{{1647, 1683}},
		"鲁王监国": [][]int{{1646, 1653}},
		"隆武":   [][]int{{1645, 1646}},
		"弘光":   [][]int{{1645, 1645}},
	}
}

func innerSolarToLunarNanMing(date Time) Time {
	year := date.solarTime.Year()
	month := int(date.solarTime.Month())
	day := date.solarTime.Day()
	if year > 1644 && year < 1654 {
		lyear, lmonth, ganzhiMonth, lday, isLeap, ldesc := rapidLunarHan2Qing(year, month, day, 0, nanMingCals)
		date.lunars = append(date.lunars, LunarTime{
			year:        lyear,
			month:       lmonth,
			day:         lday,
			leap:        isLeap,
			desc:        ldesc,
			ganzhiMonth: ganzhiMonth,
			comment:     "",
			eras:        innerEras(lyear, nanMingEras01),
		})
	}
	if year > 1646 && year < 1684 {
		lyear, lmonth, ganzhiMonth, lday, isLeap, ldesc := rapidLunarHan2Qing(year, month, day, 0, nanMingCals)
		date.lunars = append(date.lunars, LunarTime{
			year:        lyear,
			month:       lmonth,
			day:         lday,
			leap:        isLeap,
			desc:        ldesc,
			ganzhiMonth: ganzhiMonth,
			comment:     "",
			eras:        innerEras(lyear, nanMingEras02),
		})
	}
	return date
}
