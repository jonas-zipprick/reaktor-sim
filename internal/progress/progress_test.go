package progress_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jonas/reaktor-sim/internal/progress"
)

func TestBarRenderAndFinish(t *testing.T) {
	var buf bytes.Buffer
	bar := progress.NewBarTo(&buf, "Seeds", 4, 10)
	bar.Set(2)
	line := buf.String()
	if !strings.Contains(line, "2/4") || !strings.Contains(line, "50%") {
		t.Fatalf("progress line = %q", line)
	}
	bar.Finish()
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Fatalf("expected trailing newline, got %q", buf.String())
	}
}
