package seedsearch_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func TestEndDemandSummary(t *testing.T) {
	o := seedsearch.Outcome{
		AvgEndDemand: seedsearch.ZoneTotals{
			board.ZoneIndustry:    2,
			board.ZoneResidential: 1.5,
			board.ZoneRail:        0,
			board.ZonePlant:       2,
		},
	}
	got := o.EndDemandSummary()
	want := "I2 W1.5 b0 R2"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEndDamageSummaryEmpty(t *testing.T) {
	got := seedsearch.Outcome{}.EndDamageSummary()
	if got != "-" {
		t.Fatalf("got %q, want \"-\"", got)
	}
}
