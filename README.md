# astro
用了多年的天文自用代码，最开始用vb写的，后来转成php，最近又转写到golang，代码大多不规范（慢慢改
基本上是参考实现《天文算法》一书中内容，行星算法为vsop87

## 使用

go get github.com/starainrt/astro

## 示例


```golang
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/star"

	"github.com/starainrt/astro/calendar"
	"github.com/starainrt/astro/venus"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)  //指定2020年1月1日8时8分8秒
	fmt.Println(calendar.ChineseLunar(date))                   //2020年1月1日农历
	fmt.Println(calendar.ChineseLunar(date.AddDate(0, 5, 10))) //2020年06月11日农历（闰月）
	fmt.Println(calendar.Solar(2020, 1, 1, false))             //2020年大年初一对应的公历
	fmt.Println(calendar.JieQi(2020, calendar.JQ_大暑))          //2020年大暑时刻

	fmt.Println(sun.RiseTime(date, 117, 40, true)) //东经117北纬40度日出时刻
	fmt.Println(sun.CulminationTime(date, 117))    //东经117北纬40度太阳中天时刻
	fmt.Println(sun.DownTime(date, 117, 40, true)) //东经117北纬40度日落时刻
	fmt.Println(sun.Zenith(date, 117, 40))         //太阳高度角
	fmt.Println(sun.Azimuth(date, 117, 40))        //太阳方位角
	ra, dec := sun.SeeRaDec(date)                     //太阳天文坐标（赤经赤纬）
	fmt.Println(ra, dec)
	fmt.Println(star.Constellation(ra, dec, date)) //太阳所在星座
	fmt.Println(sun.EarthDistance(date))           //日地距离

	fmt.Println(moon.RiseTime(date, 117, 40, true)) //东经117北纬40度月出时刻
	fmt.Println(moon.DownTime(date, 117, 40, true)) //东经117北纬40度月落时刻
	fmt.Println(moon.Zenith(date, 117, 40))         //月球高度角
	fmt.Println(moon.Azimuth(date, 117, 40))        //月球方位角
	fmt.Println(moon.Phase(date))                      //月相
	ra, dec = moon.SeeRaDec(date, 117, 40)
	fmt.Println(star.Constellation(ra, dec, date)) //月亮所在星座
	fmt.Println(moon.EarthDistance(date))          //日月距离

	fmt.Println(venus.SeeRaDec(calendar.Date2JDE(date))) //金星天文坐标
}

```

输出如下

```
12 7 false 腊月初七     //2020年1月1日农历
4 20 true 闰四月二十    //2020年06月11日农历（闰月）
2020-01-25 00:00:00 +0800 CST   //2020年大年初一对应的公历
2020-07-22 16:36:44.138101637 +0800 CST //2020年大暑时刻
2020-01-01 07:33:45.119036436 +0800 CST <nil>  //东经117北纬40度日出时刻
2020-01-01 12:15:19.122876226 +0800 CST         //东经117北纬40度太阳中天时刻
2020-01-01 16:56:47.901035249 +0800 CST <nil>   //东经117北纬40度日落时刻
4.702335599543981   //太阳高度角
125.59379868359358  //太阳方位角
280.8951498494854 -23.05837169975993  //太阳天文坐标（赤经赤纬）
人马座  //太阳所在星座
0.9832858179003018  //日地距离，单位：天文单位（au）


2020-01-01 11:25:58.860839903 +0800 CST <nil>   //东经117北纬40度月出时刻
2020-01-01 22:48:32.983140349 +0800 CST <nil>   //东经117北纬40度月落时刻
-37.1795367151122   //月球高度角
70.09670289513748   //月球方位角
0.3000487434203941  //月相（被照亮的比例）
宝瓶座  //月亮所在星座
404238.68856284436  //日月距离（单位：千米）


317.8537654752149 -18.1404176641921 //金星天文坐标
```


## 实现

1.太阳坐标、高度角、方位角、中天、晨昏朦影、日出日落、节气  
2.月亮坐标、高度角、方位角、中天、升落  
3. 地球偏心率、日地距离
4.真平恒星时，指定赤经赤纬对应星座  
5.七大行星坐标、距日、距地距离  
6.公农历转换  
7.待续  

## TODO

1. 代码规范化
2. 增加内行星上下合/留/最大距角计算
3. 增加外行星冲/合/留计算
4. 恒星相关计算

