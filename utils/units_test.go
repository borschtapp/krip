package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindNumberWithUnit(t *testing.T) {
	tests := []struct {
		give     string
		wantVal  *float64
		wantUnit Unit
	}{
		{"0.6 g", new(0.6), UnitGram},
		{"240 kcal", new(240.0), UnitKilocalorie},
		{"240 calories", new(240.0), UnitKilocalorie},
		{"12 grams", new(12.0), UnitGram},
		{"600mg", new(600.0), UnitMilligram},
		{"1004 kJ", new(1004.0), UnitKilojoule},
		{"12 г", new(12.0), UnitGram},
		{"600 мг", new(600.0), UnitMilligram},
		{"33", new(33.0), ""},
		{"hello world", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			gotVal, gotUnit := FindNumberWithUnit(tt.give)
			assert.Equal(t, tt.wantVal, gotVal)
			assert.Equal(t, tt.wantUnit, gotUnit)
		})
	}
}

func TestParseMass(t *testing.T) {
	tests := []struct {
		give   string
		target Unit
		want   *float64
	}{
		{"0.6 g", UnitMilligram, new(600.0)},
		{"600 mg", UnitMilligram, new(600.0)},
		{"12 grams", UnitGram, new(12.0)},
		{"0.012 kg", UnitGram, new(12.0)},
		{"0.6 г", UnitMilligram, new(600.0)},
		{"500 mcg", UnitMilligram, new(0.5)},
		{"33", UnitGram, new(33.0)}, // no unit: trust documented default
		{"12 kcal", UnitGram, nil},  // recognized but wrong category: drop
		{"hello world", UnitGram, nil},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := ParseMass(tt.give, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseKilocalories(t *testing.T) {
	tests := []struct {
		give string
		want *float64
	}{
		{"240 kcal", new(240.0)},
		{"240 calories", new(240.0)},
		{"240", new(240.0)}, // no unit: trust documented default (kcal)
		{"1004 kJ", new(239.96)},
		{"12 g", nil}, // recognized but wrong category: drop
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := ParseKilocalories(tt.give)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.InDelta(t, *tt.want, *got, 0.01)
			}
		})
	}
}
