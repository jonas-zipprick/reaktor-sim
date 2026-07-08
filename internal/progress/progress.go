// Package progress renders simple terminal progress bars on stderr.
package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Bar is a single-line terminal progress indicator.
type Bar struct {
	out    io.Writer
	label  string
	total  int64
	done   int64
	width  int
	start  time.Time
	closed bool
}

// NewBar creates a progress bar for total items. width is the bar character count.
func NewBar(label string, total int64, width int) *Bar {
	return NewBarTo(os.Stderr, label, total, width)
}

// NewBarTo creates a progress bar writing to w (mainly for tests).
func NewBarTo(w io.Writer, label string, total int64, width int) *Bar {
	if total < 1 {
		total = 1
	}
	if width < 10 {
		width = 30
	}
	return &Bar{
		out:   w,
		label: label,
		total: total,
		width: width,
		start: time.Now(),
	}
}

// SetTotal updates the expected item count (e.g. when branching work grows).
func (b *Bar) SetTotal(total int64) {
	if b.closed || total < 1 {
		return
	}
	b.total = total
	if b.done > b.total {
		b.done = b.total
	}
	b.render()
}

// Set updates the completed item count and redraws the bar.
func (b *Bar) Set(done int64) {
	if b.closed || done < 0 {
		return
	}
	if done > b.total {
		done = b.total
	}
	b.done = done
	b.render()
}

// Finish marks the bar complete and prints a trailing newline.
func (b *Bar) Finish() {
	if b.closed {
		return
	}
	b.done = b.total
	b.render()
	fmt.Fprintln(b.out)
	b.closed = true
}

func (b *Bar) render() {
	pct := float64(b.done) / float64(b.total)
	filled := int(pct * float64(b.width))
	if filled > b.width {
		filled = b.width
	}
	bar := strings.Repeat("=", filled)
	if filled < b.width {
		bar += ">"
		bar += strings.Repeat(" ", b.width-filled-1)
	}

	eta := ""
	if b.done > 0 && b.done < b.total {
		elapsed := time.Since(b.start)
		remaining := time.Duration(float64(elapsed) / float64(b.done) * float64(b.total-b.done))
		eta = fmt.Sprintf(" ETA %s", remaining.Round(time.Second))
	}

	label := b.label
	if label != "" {
		label += " "
	}
	fmt.Fprintf(b.out, "\r%s[%s] %d/%d (%3.0f%%)%s   ",
		label, bar, b.done, b.total, pct*100, eta)
}
