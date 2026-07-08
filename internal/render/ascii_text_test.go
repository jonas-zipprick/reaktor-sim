package render

import "testing"

func TestASCII(t *testing.T) {
	in := "Spieler 1 schießt Wärme vom Zünder — überschüssig"
	want := "Spieler 1 schiesst Waerme vom Zuender - ueberschuessig"
	if got := ASCII(in); got != want {
		t.Fatalf("ASCII() = %q, want %q", got, want)
	}
}

func TestWrapCaption(t *testing.T) {
	long := "Spieler 1 schiesst eine Waerme vom Zuender in Richtung E. Waerme trifft das Erdgas. Das Erdgas schiesst 4 Waerme ab."
	lines := WrapCaption(long, 280)
	if len(lines) < 2 {
		t.Fatalf("expected wrapped lines, got %d: %v", len(lines), lines)
	}
	for _, line := range lines {
		if len(line)*labelCharWidth > 280 {
			t.Fatalf("line too wide: %q", line)
		}
	}
}
