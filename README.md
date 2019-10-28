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

	"github.com/starainrt/astro/venus"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"

	"github.com/starainrt/astro/basic"

	"github.com/starainrt/astro"
)

func main() {
	fmt.Println(astro.Lunar(2019, 10, 24))  //2019年10月24日农历
	fmt.Println(astro.Solar(2020, 1, 1, false))  //2020年大年初一对应的公历
	fmt.Println(astro.JDE2Date(basic.GetJQTime(2019, 270))) //2019年冬至日时刻
	fmt.Println(astro.SunRiseTime(astro.Date2JDE(time.Now()), 117, 40, 8, true)) //东经117北纬40度今日日出时刻
	fmt.Println(astro.SunDownTime(astro.Date2JDE(time.Now()), 117, 40, 8, true))  //东经117北纬40度今日日落时刻
	fmt.Println(astro.MoonRiseTime(astro.Date2JDE(time.Now()), 117, 40, 8, true))  //东经117北纬40度今日月出时刻
	fmt.Println(astro.MoonDownTime(astro.Date2JDE(time.Now()), 117, 40, 8, true))  //东经117北纬40度今日月落时刻
	fmt.Println(sun.Zenith(astro.Date2JDE(time.Now()), 117, 40, 8))  //当前太阳高度角
	ra, dec := sun.SeeRaDec(astro.NowJDE() - 8.0/24.0)   //当前太阳天文坐标（赤经赤纬）
	fmt.Println(basic.WhichCst(ra, dec, astro.NowJDE()-8.0/24.0))  //当前太阳所在星座
	ra, dec = moon.SeeRaDec(astro.NowJDE()-8.0/24.0, 117, 40, 8) 
	fmt.Println(basic.WhichCst(ra, dec, astro.NowJDE()-8.0/24.0))  //当前月亮所在星座
	fmt.Println(venus.SeeRaDec(astro.NowJDE() - 8.0/24.0))  //当前金星天文坐标
}
```

输出如下

```
9 26 false 九月廿六  //2019年10月24日农历
2020-01-25 00:00:00 +0800 CST  //2020年大年初一对应的公历
2019-12-22 04:19:19.162276983 +0800 CST  //2019年冬至日时刻
2019-10-24 06:31:15.768019258 +0800 CST <nil> //东经117北纬40度今日日出时刻
2019-10-24 17:20:43.362249433 +0800 CST <nil>  //东经117北纬40度今日日落时刻
2019-10-24 01:20:47.607341408 +0800 CST <nil>  //东经117北纬40度今日月出时刻
2019-10-24 15:27:32.907188236 +0800 CST <nil>   //东经117北纬40度今日月落时刻
37.21988664913089  //当前太阳高度角
室女座  //当前太阳所在星座
狮子座  //当前月亮所在星座
226.7261547571923 -17.473439221980396  //当前金星天文坐标
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

