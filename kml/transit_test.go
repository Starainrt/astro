package kml

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/mercury"
)

// 凌日 GeoJSON 转 KML：图层名本地化、分区带填充、等时线逐段带 TimeStamp。
func TestPlanetTransitConversion(t *testing.T) {
	info := mercury.ClosestTransit(time.Date(2032, 11, 13, 0, 0, 0, 0, time.UTC))
	data, err := geojson.MarshalMercuryTransit(info, geojson.PlanetTransitOptions{IsochroneStep: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	root := convert(t, string(data), Options{})
	if !strings.HasPrefix(root.Document.Name, "凌日 2032-11-13") {
		t.Fatalf("Document 名 = %q", root.Document.Name)
	}
	if strings.Contains(root.Document.Description, "日食") || !strings.Contains(root.Document.Description, "凌日") {
		t.Fatalf("凌日说明不匹配：%q", root.Document.Description)
	}
	folders := map[string]folder{}
	for _, item := range root.Document.Folders {
		folders[item.Name] = item
	}
	for _, name := range []string{"全程可见", "日出时已在凌日", "日落时仍在凌日", "一触视地平线", "二触等时线", "四触等时线", "凌甚"} {
		if _, ok := folders[name]; !ok {
			t.Fatalf("缺少图层 %s", name)
		}
	}
	for _, placemark := range folders["二触等时线"].Placemarks {
		if placemark.TimeStamp == nil {
			t.Fatal("等时线 Placemark 必须带 TimeStamp")
		}
	}
	if style := paletteStyle(eventPlanetTransit, "transit-entire"); style.FillColor != transitEntireFill {
		t.Fatalf("全程可见区填充 = %q", style.FillColor)
	}
}
