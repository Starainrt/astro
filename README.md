# Astro

[![Go Reference](https://pkg.go.dev/badge/github.com/starainrt/astro.svg)](https://pkg.go.dev/github.com/starainrt/astro)

自用多年的天文算法库，用于个人天文历法爱好。

>📚 本项目主要用于天文算法学习与验证，计算结果满足业余爱好级别需求。

基于《天文算法》（Astronomical Algorithms）一书实现，提供历法转换、行星位置、月相、日出日落，月出月落等天文计算功能。包含VSOP87行星算法和ELP2000/82月球算法。

没有特殊标注时，本程序所提供的坐标均为瞬时天球坐标。

精度说明：VSOP87[-2000年，6000年]行星位置精度约为1角秒


## 目录

- [安装](#安装)
- [功能概览](#功能概览)
- [快速开始](#快速开始)
   - [历法转换与节气](#历法转换与节气)
   - [太阳与月亮](#太阳与月亮)
   - [行星](#行星)
   - [恒星](#恒星)
- [TODO](#todo)

## 安装

```bash
go get github.com/starainrt/astro
```

## 功能概览

- 📅 **历法转换**：公历与农历互转（公元前104年-公元3000年或更久）、节气时刻
- 🌞 **太阳计算**：天球位置、日出日落、日地距离、真太阳时等
- 🌙 **月亮计算**：天球位置、月出月落、日月距离、月相、朔望时间等
- 🪐 **行星计算**：七大行星天球位置、升落时间、合冲留等特殊天象时间
- ⭐ **恒星计算**：指定天球坐标所属星座；同时包含9100颗恒星数据库，可计算升降时间，获取指定日期的恒星坐标信息

## 快速开始

### 历法转换与节气

本工具支持公历与农历日期之间的互相转换，支持年份范围为公元前104年正月至公元3000年（即[-103, 3000]）。

#### 数据来源与校对

- **[-103, 1912] 年**：基于《寿星天文历》数据，并依据 [ytliu0 教授整理的历表](https://ytliu0.github.io/ChineseCalendar/index_simp.html) 进行修正，目前已完成校对。
- **[1913, 3000] 年**：依据 VSOP87 定气、ELP2000 定朔，按照现行标准GB/T 33661-2017中农历算法计算。

---

#### 使用须知

##### 1. 同一公历日期可能对应多个农历日期
在多个政权并存的历史时期（如三国时期），不同政权可能使用不同历法，造成同一公历日期对应多个农历日期。本程序尽可能提供所有可能的转换结果。

##### 2. 同一农历日期可能对应多个公历日期
不仅因多个政权历法不同，同一政权在历法改革中也可能出现此类情况。例如，武则天改历后，圣历三年出现了两个腊月。

##### 3. 公历历法处理规则
本程序基于儒略日进行计算，公历部分处理规则如下：
- 1582年10月15日之后：使用格里高利历
- 1582年10月4日之前：使用儒略历
- 公元8年之前：使用逆推儒略历
- 1582年10月4日的下一天为1582年10月15日
- 年份表示：0年表示公元前1年，-1年表示公元前2年，以此类推

##### 4. Go 语言特别注意

⚠️ **Go 标准库 `time.Time` 在历法处理上与本程序存在差异：**

- Go 语言在1582年10月15日之前使用逆推格里高利历，而非儒略历。若不使用 `Add` 方法，一般可正常使用。
- 因此，**在1582年10月15日之前，`time.Time.Weekday()` 返回结果与本程序计算结果不一致**。  
  例如：1582年10月4日，本程序为星期四，Go 语言判断为星期一。

#### 建议解决方案：
如需获得与本程序一致的星期数，可使用如下方法：

```go
// date 应为当日0时的 time.Time
weekday := int(calendar.Date2JDE(date)+1.5) % 7
// 0表示星期日，1表示星期一，……，6表示星期六
```
若在1582年之前使用 time.Time 的 Add 或 AddDate 方法，请注意其在某些年份可能不准确。
例如：700年儒略历为闰年，而 Go 使用的逆推格里高利历中700年不是闰年。

#### 历法转换

##### 公历转农历

- **输入**：公历日期 (`time.Time`)
- **输出**：`calendar.Time` 对象，可能包含多个对应的农历日期
- **功能**：可从返回对象中获取：
   - 农历日期的详细描述
   - 年、月、日的天干地支
   - 所属朝代、皇帝、年号等信息
   - 完整的结构化农历信息

##### 农历转公历

支持两种调用方式：

###### 方式一：传入农历字符串
支持以下格式（示例）：
1. `年号+年+月+日`：如 **`"元丰六年十月十二"`**（闰月前加"闰"，日期格式为"初一"、"二十"等）
2. `年号+年+月+干支日`：如 **`"元嘉二十七年七月庚午"`**
3. `年份+月+日`：如 **`"二零二五年正月初一"`**（闰月前加"闰"，适用于现代日期）
4. `年份+月+干支日`：如 **`"二零二五年正月戊戌日"`**
5. `阿拉伯数字+月+日`：可以将中文数字替换为阿拉伯数字，如 **`"2025年1月1日"`**，代表`二零二五年正月初一`
6. **注意历史场景**：历史上月份名称可能与现代不同（如武则天时期“正月”与“一月”代表不同月份），请使用汉字数字确保准确性

> ⚠️ **特别提醒**：  
> 农历年份与公历年份并非完全重合。例如：公历2025年1月28日（除夕）对应农历2024年腊月二十九，应传入 `"二零二四年腊月廿九"`。

###### 方式二：传入数字参数
- **参数**：年份 (`int`)、月份 (`int`)、日期 (`int`)、是否闰月 (`bool`)
- **特点**：简单直接，适用于现代农历日期转换

##### 代码示例

```go
package main

import (
   "encoding/json"
   "fmt"
   "github.com/starainrt/astro/calendar"
   "time"
)

func main() {
   cst := time.FixedZone("CST", 8*3600)

   // 示例1：公历转农历
   date := time.Date(240, 1, 1, 8, 8, 8, 8, cst)
   lunar, _ := calendar.SolarToLunar(date)
   fmt.Println(lunar.LunarDescWithEmperor()) // 输出农历详细描述

   info := lunar.LunarInfo()
   data, _ := json.MarshalIndent(info, "", "  ")
   fmt.Println(string(data)) // 输出结构化农历信息

   // 示例2：农历转公历（字符串格式）
   solar, _ := calendar.LunarToSolar("元丰六年十月十二日")
   for _, v := range solar {
      fmt.Println(v.Time())
      fmt.Println(v.LunarDescWithEmperor())
   }

   // 示例3：农历转公历（数字参数格式）
   modernDate, _ := calendar.LunarToSolarSingle(2025, 1, 1, false)
   fmt.Println(modernDate.Time())
}
```

输出结果：

```aiignore
// 三国时期同一公历日期对应多个农历结果
[魏明帝 景初三年腊月二十 蜀后主 延熙二年冬月十九 吴大帝 赤乌二年冬月二十]

// 结构化农历信息输出
[
  {
    "solarDate": "0240-01-01T08:08:08.000000008+08:00",
    "lunarYear": 239,
    "lunarYearChn": "二三九",
    "lunarMonth": 12,
    "lunarDay": 20,
    "isLeap": false,
    "lunarMonthDayDesc": "腊月二十",
    "ganzhiYear": "己未",
    "ganzhiMonth": "丙子",
    "ganzhiDay": "辛未",
    "dynasty": "魏",
    "emperor": "魏明帝",
    "nianhao": "景初",
    "yearOfNianhao": 3,
    "eraDesc": "景初三年",
    "lunarWithNianhaoDesc": "景初三年腊月二十",
    "chineseZodiac": "羊"
  },
  {
    "solarDate": "0240-01-01T08:08:08.000000008+08:00",
    "lunarYear": 239,
    "lunarYearChn": "二三九",
    "lunarMonth": 11,
    "lunarDay": 19,
    "isLeap": false,
    "lunarMonthDayDesc": "冬月十九",
    "ganzhiYear": "己未",
    "ganzhiMonth": "丙子",
    "ganzhiDay": "辛未",
    "dynasty": "",
    "emperor": "蜀后主",
    "nianhao": "延熙",
    "yearOfNianhao": 2,
    "eraDesc": "延熙二年",
    "lunarWithNianhaoDesc": "延熙二年冬月十九",
    "chineseZodiac": "羊"
  },
  {
    "solarDate": "0240-01-01T08:08:08.000000008+08:00",
    "lunarYear": 239,
    "lunarYearChn": "二三九",
    "lunarMonth": 11,
    "lunarDay": 20,
    "isLeap": false,
    "lunarMonthDayDesc": "冬月二十",
    "ganzhiYear": "己未",
    "ganzhiMonth": "丙子",
    "ganzhiDay": "辛未",
    "dynasty": "吴",
    "emperor": "吴大帝",
    "nianhao": "赤乌",
    "yearOfNianhao": 2,
    "eraDesc": "赤乌二年",
    "lunarWithNianhaoDesc": "赤乌二年冬月二十",
    "chineseZodiac": "羊"
  }
]

// 苏轼，记承天寺夜游，元丰六年十月十二日对应的公历日期
1083-11-24 00:00:00 +0800 CST
[宋神宗 元丰六年十月十二 辽道宗 大康九年十月十二]

// 现代农历日期转换结果
2025-01-29 00:00:00 +0800 CST
```

#### 节气

```golang

package main

import (
	"fmt"
	"github.com/starainrt/astro/calendar"
)

func main() {
  //计算2020年立春时间
	fmt.Println(calendar.JieQi(2020, calendar.JQ_立春))
  //计算2020年冬至时间
	fmt.Println(calendar.JieQi(2020, calendar.JQ_冬至))
  //计算2020年春分时间
	fmt.Println(calendar.JieQi(2020, calendar.JQ_春分))
  //也可传入节气对应太阳黄经，如春分时太阳黄经为0度，这里计算2020年春分时间
	fmt.Println(calendar.JieQi(2020, 0))
}

```

输出结果

```
2020-02-04 17:03:17.820854187 +0800 CST //2020年立春时间
2020-12-21 18:02:17.568823993 +0800 CST //2020年冬至时间
2020-03-20 11:49:34.502393603 +0800 CST //2020年春分时间
2020-03-20 11:49:34.502393603 +0800 CST //2020年春分时间

```


### 太阳与月亮

#### 日出日落/月出月落

> ⚠️ **重要说明**：  
> 月球升降时间计算基于当天日期，升降时间点之间不一定具有连续性。
>
> **可能出现的情况**：
> - 月亮可能在凌晨1点落下，中午12点再次升起，此时升起时间会晚于降落时间；要获取此场景晚上的月落时间，需要传入次日日期进行计算
> 
> **如需获取完整升降周期，需要自行通过判断升起时间是否在降落时间之后来确定后续的正确时间点**

```golang
package main

import (
	"fmt"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
	"time"
)

func main() {
	// 以陕西省西安市为例，设置西安市经纬度,设置地平高度为0米
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	//指定2020年1月1日8时8分8秒
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	// 西安市2020年1月1日民用晨朦影开始时间
	// 民用朦影，太阳位于地平线下6度，航海朦影=地平线下12度，天文朦影=地平线下18度
	fmt.Println(sun.MorningTwilight(date, lon, lat, -6))
	// 西安市2020年1月1日日出时间，计算大气影响
	fmt.Println(sun.RiseTime(date, lon, lat, height, true))
	// 西安市2020年1月1日太阳上中天时间
	fmt.Println(sun.CulminationTime(date, lon))
	// 西安市2020年1月1日日落时间，计算大气影响
	fmt.Println(sun.SetTime(date, lon, lat, height, true))
	// 西安市2020年1月1日民用昏朦影结束时间
	fmt.Println(sun.EveningTwilight(date, lon, lat, -6))

	// 西安市2020年1月1日月出时间，计算大气影响
	fmt.Println(moon.RiseTime(date, lon, lat, height, true))
	// 西安市2020年1月1日月亮上中天时间
	fmt.Println(moon.CulminationTime(date, lon, lat))
	// 西安市2020年1月1日月落时间，计算大气影响
	fmt.Println(moon.SetTime(date, lon, lat, height, true))
}
```


输出结果

```
2020-01-01 07:22:27.964431345 +0800 CST <nil> //西安市1月1日晨朦影开始时间
2020-01-01 07:50:14.534510672 +0800 CST <nil> //西安市1月1日日出时间
2020-01-01 12:47:35.933117866 +0800 CST       //西安市1月1日太阳上中天时间
2020-01-01 17:44:47.076647579 +0800 CST <nil> //西安市1月1日日落时间
2020-01-01 18:12:33.629668056 +0800 CST <nil> //西安市1月1日昏朦影结束时间
2020-01-01 11:52:44.643359184 +0800 CST <nil> //西安市1月1日月出时间
2020-01-01 17:38:03.879639208 +0800 CST       //西安市1月1日月亮上中天时间
2020-01-01 23:26:52.202896177 +0800 CST <nil> //西安市1月1日月落时间


```

#### 日月位置


```golang
package main

import (
	"fmt"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/sun"
	"github.com/starainrt/astro/tools"
	"time"
)

func main() {
	// 以陕西省西安市为例，设置西安市经纬度,设置地平高度为0米
	var lon, lat float64 = 108.93, 34.27
	cst := time.FixedZone("CST", 8*3600)
	//指定2020年1月1日8时8分8秒
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	//太阳此刻黄经
	fmt.Println(sun.ApparentLo(date))
	//黄赤交角
	fmt.Println(sun.EclipticObliquity(date, true))
	//太阳此刻视赤经、视赤纬
	ra, dec := sun.ApparentRaDec(date)
	fmt.Println("赤经：", tools.Format(ra/15, 1), "赤纬：", tools.Format(dec, 0))
	//太阳当前所在星座
	fmt.Println(star.Constellation(ra, dec, date))
	//此刻西安市的太阳方位角与高度角
	fmt.Println("方位角：", sun.Azimuth(date, lon, lat), "高度角：", sun.Zenith(date, lon, lat))
    //此刻日地距离，单位为天文单位（AU）
	fmt.Println(sun.EarthDistance(date))

	//月亮此刻站心视赤经、视赤纬
	ra, dec = moon.ApparentRaDec(date, lon, lat)
	fmt.Println("赤经：", tools.Format(ra/15, 1), "赤纬：", tools.Format(dec, 0))
	//月亮当前所在星座
	fmt.Println(star.Constellation(ra, dec, date))
	//此刻西安市的月亮方位角与高度角
	fmt.Println("方位角：", moon.Azimuth(date, lon, lat), "高度角：", moon.Zenith(date, lon, lat))
    //此刻地月距离，单位为千米（AU）
	fmt.Println(moon.EarthDistance(date))
}
```

输出

```
280.0152925179703 //太阳黄经
23.436215552851408 //黄赤交角
赤经： 18h43m34.83s 赤纬： -23°3′30.25″
人马座 //太阳所在星座
方位角： 120.19483856399326 高度角： 2.4014324584398516
0.9832929365443133 //日地距离

赤经： 23h17m51.93s 赤纬： -10°19′17.02″
宝瓶座 //月亮所在星座
方位角： 67.84449893794012 高度角： -45.13018696439911
404238.6354387698 //地月距离
```

#### 月相

```golang
package main

import (
	"fmt"
	"github.com/starainrt/astro/moon"
	"time"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	//指定2020年1月1日8时8分8秒
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	//月亮此刻被照亮的比例（月相）
	fmt.Println(moon.Phase(date))
	//月相具体描述
	fmt.Println(moon.PhaseDesc(date))
	//下次朔月时间
	fmt.Println(moon.NextShuoYue(date))
	//下次上弦月时间
	fmt.Println(moon.NextShangXianYue(date))
	//下次望月时间
	fmt.Println(moon.NextWangYue(date))
	//下次下弦月时间
	fmt.Println(moon.NextXiaXianYue(date))
}
```

输出

```
0.3000437415436273 //被照亮30%
上峨眉月 //月相描述
2020-01-25 05:41:55.820311009 +0800 CST //下次朔月
2020-01-03 12:45:20.809730887 +0800 CST //下次上弦
2020-01-11 03:21:14.729664623 +0800 CST //下次满月
2020-01-17 20:58:20.955985486 +0800 CST //下次下弦
```



### 行星

#### 内行星

```golang
package main

import (
	"fmt"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
	"time"
)

func main() {
	// 以陕西省西安市为例，设置西安市经纬度,设置地平高度为0米
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	//指定2020年1月1日8时8分8秒
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	//水星上次下合时间
	fmt.Println(mercury.LastInferiorConjunction(date))
	//金星下次上合时间
	fmt.Println(venus.NextSuperiorConjunction(date))
	//水星上次留（顺转逆）时间（水逆）
	fmt.Println(mercury.LastProgradeToRetrograde(date))
	//金星下次留（逆转顺）时间
	fmt.Println(venus.NextRetrogradeToPrograde(date))
	//水星上次东大距时间
	fmt.Println(mercury.LastGreatestElongationEast(date))
	//金星下次西大距时间
	fmt.Println(venus.NextGreatestElongationWest(date))
	//西安市今日金星升起，降落时间
	fmt.Println(venus.RiseTime(date, lon, lat, height, true))
	fmt.Println(venus.SetTime(date, lon, lat, height, true))
	//金星当前视星等
	fmt.Println(venus.ApparentMagnitude(date))
	//金地距离
	fmt.Println(venus.EarthDistance(date))
	//金日距离
	fmt.Println(venus.SunDistance(date))
}
```

输出

```
2019-11-11 23:21:39.702344834 +0800 CST //水星上次下合时间
2021-03-26 14:57:38.289429545 +0800 CST //金星下次上合时间
2019-11-01 04:31:47.807287573 +0800 CST //水星上次留（顺转逆）时间（水逆）
2021-12-18 18:59:12.762369811 +0800 CST //金星下次留（逆转顺）时间
2019-10-20 11:59:33.893027007 +0800 CST //水星上次东大距时间
2020-08-13 07:56:02.326616048 +0800 CST //金星下次西大距时间
2020-01-01 10:01:10.821288228 +0800 CST <nil> //西安市今日金星升起时间
2020-01-01 20:27:00.741534233 +0800 CST <nil> //西安市今日金星降落时间
-4 //金星视星等
1.2760033106813273 //金地距离
0.7262288470390035 //金日距离
```

#### 外行星

```golang
package main

import (
	"fmt"
	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/uranus"
	"time"
)

func main() {
	// 以陕西省西安市为例，设置西安市经纬度,设置地平高度为0米
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	//指定2020年1月1日8时8分8秒
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	//火星下次冲日时间
	fmt.Println(mars.NextOpposition(date))
	//木星下次合日时间
	fmt.Println(jupiter.NextConjunction(date))
	//土星上次留（顺转逆）时间（水逆）
	fmt.Println(saturn.LastProgradeToRetrograde(date))
	//天王星下次留（逆转顺）时间
	fmt.Println(uranus.NextRetrogradeToPrograde(date))
	//海王星上次东方照时间
	fmt.Println(neptune.LastEasternQuadrature(date))
	//火星下次西方照时间
	fmt.Println(mars.NextWesternQuadrature(date))
	//西安市今日火星升起，降落时间
	fmt.Println(mars.RiseTime(date, lon, lat, height, true))
	fmt.Println(mars.SetTime(date, lon, lat, height, true))
	//火星当前视星等
	fmt.Println(mars.ApparentMagnitude(date))
	//地火距离
	fmt.Println(mars.EarthDistance(date))
	//日火距离
	fmt.Println(mars.SunDistance(date))
}

```
输出

```
2020-10-14 07:25:47.740884125 +0800 CST //火星下次冲日时间
2021-01-29 09:39:30.916356146 +0800 CST //木星下次合日时间
2019-04-30 10:28:27.453395426 +0800 CST //土星上次留（顺转逆）时间（水逆）
2021-01-14 21:35:01.269377768 +0800 CST //天王星下次留（逆转顺）时间
2019-12-08 17:00:13.772284984 +0800 CST //海王星上次东方照时间
2020-06-07 03:10:57.179121673 +0800 CST //火星下次西方照时间
2020-01-01 04:40:05.409269034 +0800 CST <nil> //西安市今日火星升起时间
2020-01-01 14:56:57.175483703 +0800 CST <nil> //西安市今日火星降落时间
1.57 //火星当前视星等
2.1820316323604088 //地火距离
1.5894169865107062 //日火距离

```

### 恒星

1. 本程序自带9100颗恒星的数据库，已经自动计算自行

```golang
package main

import (
	"fmt"
	"github.com/starainrt/astro/star"
	"time"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	//指定2020年1月1日8时8分8秒
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)

	//初始化恒星数据库
	star.InitStarDatabase()
	sirius, _ := star.StarDataByName("天狼")
	//天狼星升起时间
	riseDate, _ := star.RiseTime(date, sirius.Ra, sirius.Dec, 115, 40, 0, true)
	fmt.Println(riseDate)
	//天狼星降落时间
	setDate, _ := star.SetTime(date, sirius.Ra, sirius.Dec, 115, 40, 0, true)
	fmt.Println(setDate)
}
```

```
2019-12-31 19:21:56.993647813 +0800 CST //天狼星升起时间
2020-01-01 05:29:53.535125255 +0800 CST  //天狼星降落时间
```

## 已实现

- ✅ 太阳位置、高度角、方位角、中天、晨昏朦影、日出日落、节气
- ✅ 月亮位置、高度角、方位角、中天、升落、月相
- ✅ 地球偏心率、日地距离
- ✅ 真平恒星时、星座计算
- ✅ 七大行星坐标、距日距地距离、特殊天象计算
- ✅ 公农历转换（公元前104年-公元3000年）
- ✅ 9100+恒星数据库

## TODO

- 🔄 代码规范化与性能优化
- 🔄 增强恒星计算功能
- 🔄 日食月食计算
- 🔄 更多天文现象计算
