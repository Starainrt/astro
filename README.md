# astro
用了多年的天文自用代码， 基本上按照《天文算法》一书中内容进行实现，行星算法为vsop87

## 使用

go get github.com/starainrt/astro

## 示例


### 历法与节气

按现行农历GB/T 33661-2017算法计算，古代由于定朔定气误差此处计算结果会与古籍不符，推荐使用年份为[1929-6000]年


#### 历法

公农历互转

```golang
package main

import (
	"fmt"
	"github.com/starainrt/astro/calendar"
	"time"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	//指定2020年1月1日8时8分8秒
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	//公历转农历: 公历2020年1月1日转农历
	fmt.Println(calendar.SolarToLunar(date))
	//农历转公历：农历2020年（鼠年）正月初一
	fmt.Println(calendar.LunarToSolar(2020, 1, 1, false))
}

```

输出结果

```
12 7 false 腊月初七       //输出农历月份、日期、是否为闰月、汉字表述。
2020-01-25 00:00:00 +0800 CST //输出本地时区0时公历时间。

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

### 外行星

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


## 已实现

1. 太阳坐标、高度角、方位角、中天、晨昏朦影、日出日落、节气  
2. 月亮坐标、高度角、方位角、中天、升落  
3. 地球偏心率、日地距离
4. 真平恒星时，指定赤经赤纬对应星座  
5. 七大行星坐标、距日、距地距离、内行星上下合/留/最大距角计算、外行星冲/合/留计算
6. 公农历转换  
7. 待续  

## TODO

1. 代码规范化
2. 恒星相关计算
3. 日食、月食相关计算

