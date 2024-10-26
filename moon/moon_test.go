package moon

import (
	"fmt"
	"testing"
	"time"
)

func Test_MoonPhaseDate(t *testing.T) {
	//指定北京时间2022年1月20日
	tz := time.FixedZone("CST", 8*3600)
	date := time.Date(2022, 01, 20, 00, 00, 00, 00, tz)
	//指定日期后的下一个朔月
	moonPhase01 := NextShuoYue(date)
	fmt.Println("下一朔月", moonPhase01)
	if moonPhase01.Unix() != 1643694356 {
		t.Fatal(moonPhase01.Unix())
	}
	//指定日期后的上一个朔月
	moonPhase01 = LastShuoYue(date)
	fmt.Println("上一朔月", moonPhase01)
	if moonPhase01.Unix() != 1641148406 {
		t.Fatal(moonPhase01.Unix())
	}
	//离指定日期最近的朔月
	moonPhase01 = ClosestShuoYue(date)
	fmt.Println("最近朔月", moonPhase01)
	if moonPhase01.Unix() != 1643694356 {
		t.Fatal(moonPhase01.Unix())
	}
	//离指定日期最近的望月时间
	moonPhase01 = ClosestWangYue(date)
	fmt.Println("最近望月", moonPhase01)
	if moonPhase01.Unix() != 1642463301 {
		t.Fatal(moonPhase01.Unix())
	}
	//离指定日期最近的上弦月时间
	moonPhase01 = ClosestShangXianYue(date)
	fmt.Println("最近上弦月", moonPhase01)
	if moonPhase01.Unix() != 1641751871 {
		t.Fatal(moonPhase01.Unix())
	}
	//离指定日期最近的下弦月时间
	moonPhase01 = ClosestXiaXianYue(date)
	fmt.Println("最近下弦月", moonPhase01)
	if moonPhase01.Unix() != 1643118050 {
		t.Fatal(moonPhase01.Unix())
	}
	//-------------------
	for i := 0; i < 26; i++ {
		moonPhase01 = LastShuoYue(moonPhase01)
		fmt.Println("上一朔月", moonPhase01)
	}
}
