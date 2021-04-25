package x

import (
	"fmt"
	"testing"
	"time"
)

//立春布尔值
func TestFixLiChun(t *testing.T) {
	year := 2021
	year = 2033
	cust := time.Date(year, time.Month(2), 3, 22, 0, 0, 0, time.Local) //精确到时
	cust = time.Date(year, time.Month(2), 3, 21, 0, 0, 0, time.Local)

	cust = time.Date(year, time.Month(2), 3, 19, 0, 0, 0, time.Local)
	cust = time.Date(year, time.Month(2), 3, 20, 0, 0, 0, time.Local)

	lcb := fixLiChun(year, cust)
	fmt.Println(lcb)
}

//年干 年支
func TestYearGZ(t *testing.T) {
	year := 2021
	lcb := true //干:辛 支:丑
	lcb = false //干:庚 支:子

	year = 2033
	lcb = true  //干:癸 支:丑
	lcb = false //干:壬 支:子

	g, z := YearGZ(year, lcb)
	fmt.Printf("干:%s 支:%s\n", g, z)
}

//年干支
func TestGetYGZ(t *testing.T) {
	y, m, d, h := 2021, 2, 3, 21 //庚子
	y, m, d, h = 2021, 2, 3, 22  //辛丑
	y, m, d, h = 2033, 2, 3, 19  //壬子
	y, m, d, h = 2033, 2, 3, 20  //癸丑
	gz := GetYGZ(y, m, d, h)
	fmt.Println(gz)
}
