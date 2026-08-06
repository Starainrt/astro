package moon

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestOccultationCommonInputsValidate(t *testing.T) {
	if err := (Observer{Longitude: 121.5, Latitude: 31.2, Height: 10}).Validate(); err != nil {
		t.Fatalf("valid observer rejected: %v", err)
	}
	if err := (OccultationSearchOptions{SafetyMarginArcsec: 2, MaxEvents: 3}).Validate(); err != nil {
		t.Fatalf("valid options rejected: %v", err)
	}
	if err := (OccultationSearchOptions{}).Validate(); err != nil {
		t.Fatalf("zero options rejected: %v", err)
	}
	if err := (OccultationSearchOptions{MaxStep: 250 * time.Millisecond}).Validate(); err != nil {
		t.Fatalf("minimum search step rejected: %v", err)
	}

	invalid := []error{
		(Observer{Latitude: 91}).Validate(),
		(Observer{Longitude: math.NaN()}).Validate(),
		(OccultationSearchOptions{SafetyMarginArcsec: -1}).Validate(),
		(OccultationSearchOptions{MaxEvents: -1}).Validate(),
		(OccultationSearchOptions{MaxStep: 100 * time.Millisecond}).Validate(),
	}
	for index, err := range invalid {
		if !errors.Is(err, ErrInvalidOccultationInput) {
			t.Fatalf("invalid input %d error = %v, want ErrInvalidOccultationInput", index, err)
		}
	}
}

func TestStarCoordinateValidate(t *testing.T) {
	valid := StarCoordinate{
		ID:                             "star",
		RA:                             101.28715533,
		Dec:                            -16.71611586,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          CoordinateFrameICRS,
		ProperMotionRACosDecMasPerYear: 10,
		ProperMotionDecMasPerYear:      -20,
		ParallaxMas:                    7,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid star coordinate rejected: %v", err)
	}

	tests := []struct {
		name   string
		modify func(*StarCoordinate)
	}{
		{"RA=360", func(s *StarCoordinate) { s.RA = 360 }},
		{"Dec out of range", func(s *StarCoordinate) { s.Dec = -91 }},
		{"missing epoch", func(s *StarCoordinate) { s.Epoch = time.Time{} }},
		{"unsupported frame", func(s *StarCoordinate) { s.Frame = CoordinateFrame("galactic") }},
		{"non-finite proper motion", func(s *StarCoordinate) { s.ProperMotionDecMasPerYear = math.NaN() }},
		{"negative parallax", func(s *StarCoordinate) { s.ParallaxMas = -1 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := valid
			tc.modify(&candidate)
			if err := candidate.Validate(); !errors.Is(err, ErrInvalidOccultationInput) {
				t.Fatalf("StarCoordinate.Validate() error = %v, want ErrInvalidOccultationInput", err)
			}
		})
	}
}

func TestStarCoordinateFromStarData(t *testing.T) {
	data, err := basic.StarDataByChinese("进贤增九")
	if err != nil {
		t.Fatalf("load HR 4799: %v", err)
	}
	data.Pc = 50

	coordinate, err := StarCoordinateFromStarData(data)
	if err != nil {
		t.Fatalf("StarCoordinateFromStarData() error = %v", err)
	}
	if coordinate.ID != "进贤增九" {
		t.Fatalf("ID = %q, want %q", coordinate.ID, "进贤增九")
	}
	if coordinate.RA != data.Ra || coordinate.Dec != data.Dec {
		t.Fatalf("RA/Dec = %.12f %.12f, want %.12f %.12f", coordinate.RA, coordinate.Dec, data.Ra, data.Dec)
	}
	if coordinate.Frame != CoordinateFrameJ2000 || !coordinate.Epoch.Equal(time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("frame/epoch = %q %v, want J2000", coordinate.Frame, coordinate.Epoch)
	}
	if coordinate.ProperMotionRACosDecMasPerYear != data.PmRA*1000 ||
		coordinate.ProperMotionDecMasPerYear != data.PmDec*1000 {
		t.Fatalf("proper motion = %.3f %.3f mas/year", coordinate.ProperMotionRACosDecMasPerYear, coordinate.ProperMotionDecMasPerYear)
	}
	if coordinate.ParallaxMas != 20 {
		t.Fatalf("parallax = %.6f mas, want 20", coordinate.ParallaxMas)
	}
}

func TestStarCoordinateFromStarDataRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name string
		data StarData
	}{
		{name: "zero value"},
		{name: "non-finite distance", data: StarData{InnerStarData: basic.InnerStarData{HR: 1, Ra: 10, Dec: 20, Pc: math.NaN()}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := StarCoordinateFromStarData(tc.data); !errors.Is(err, ErrInvalidOccultationInput) {
				t.Fatalf("StarCoordinateFromStarData() error = %v, want ErrInvalidOccultationInput", err)
			}
		})
	}
}
