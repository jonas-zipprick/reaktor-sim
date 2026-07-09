package field_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/field"
)

func TestAvailableInMonthNoFilter(t *testing.T) {
	if !field.AvailableInMonth(field.UraniumPlate, 0) {
		t.Fatal("filter 0 should allow all fields")
	}
}

func TestAvailableInMonthEarlyCampaign(t *testing.T) {
	cases := []struct {
		t    field.Type
		m    int
		want bool
	}{
		{field.Mirror, 1, true},
		{field.GasBoiler, 1, false},
		{field.GasBoiler, 2, true},
		{field.AbsorberRod, 2, false},
		{field.AbsorberRod, 3, true},
		{field.UraniumPlate, 2, false},
		{field.UraniumPlate, 3, true},
		{field.Tokamak, 3, false},
		{field.Tokamak, 4, true},
		{field.CapacitorBank, 1, false},
		{field.CapacitorBank, 2, true},
		{field.HVCascade, 2, false},
		{field.HVCascade, 3, true},
		{field.Superconductor, 3, false},
		{field.Superconductor, 4, true},
	}
	for _, tc := range cases {
		if got := field.AvailableInMonth(tc.t, tc.m); got != tc.want {
			t.Fatalf("%v month %d = %v, want %v", tc.t, tc.m, got, tc.want)
		}
	}
}

func TestFilterMarketMonth2(t *testing.T) {
	got := field.FilterMarket(field.GridMarket, 2)
	for _, ft := range got {
		if ft == field.HVCascade || ft == field.PumpedStorage {
			t.Fatalf("month 2 market contains %v", ft)
		}
	}
}
