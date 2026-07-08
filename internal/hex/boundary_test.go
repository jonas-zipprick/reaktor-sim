package hex_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestReflectOffOuterWallReversesTravelDir(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"NW", "SE"},
		{"SE", "NW"},
		{"NE", "SW"},
		{"SW", "NE"},
		{"E", "W"},
		{"W", "E"},
	}
	for _, tc := range cases {
		var inDir int
		for d := 0; d < 6; d++ {
			if hex.DisplayDirName(d) == tc.in {
				inDir = d
				break
			}
		}
		got := hex.DisplayDirName(hex.ReflectOffOuterWall(inDir))
		if got != tc.want {
			t.Fatalf("reflect %s: got %s, want %s", tc.in, got, tc.want)
		}
	}
}
