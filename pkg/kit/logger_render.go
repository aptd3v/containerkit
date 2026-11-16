package kit

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
)

func formatEvent(e LogEvent, color, timestamps bool, glyph string) string {
	var b strings.Builder
	b.Grow(96)
	if timestamps {
		b.WriteString(paint(e.Time.Format("15:04:05"), ansiDim, color))
		b.WriteByte(' ')
	}
	b.WriteString(paint(pad(string(e.Kind), 8), kindColor(e.Kind), color))
	b.WriteByte(' ')
	b.WriteString(paint(glyph, statusColor(e.Status), color))
	b.WriteByte(' ')
	id := e.ID
	if id == "" {
		id = e.Parent
	}
	b.WriteString(paint(pad(trunc(id, 22), 22), ansiBold, color))
	b.WriteByte(' ')
	op := e.Op
	if op == "" {
		op = e.Text
	}
	b.WriteString(paint(pad(trunc(op, 12), 12), ansiDim, color))
	rest := eventRest(e)
	if rest != "" {
		b.WriteByte(' ')
		b.WriteString(rest)
	}
	return b.String()
}

func eventRest(e LogEvent) string {
	var parts []string
	if label := eventLabel(e); label != "" {
		parts = append(parts, label)
	}
	if e.Total > 0 {
		bar := progressBar(e.Current, e.Total, 18)
		pct := int(float64(e.Current) / float64(e.Total) * 100)
		if pct > 100 {
			pct = 100
		}
		parts = append(parts, bar, fmt.Sprintf("%s / %s  %d%%", humanBytes(e.Current), humanBytes(e.Total), pct))
	}
	return strings.Join(parts, "  ")
}

// eventLabel is the human text for a row. A short repeat of the operation
// stays in the operation column, so "building" is not printed twice.
func eventLabel(e LogEvent) string {
	if detail := strings.TrimSpace(e.Detail); detail != "" && !shortEcho(detail, e.Op) && !shortEcho(detail, e.Text) {
		return detail
	}
	if text := strings.TrimSpace(e.Text); text != "" && !shortEcho(text, e.Op) {
		return text
	}
	return ""
}

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// shortEcho reports that candidate only repeats op, so the detail column
// would show "started" and "Started" on the same line. Longer text is kept:
// the op column is 12 columns wide and would hide the rest of the message.
func shortEcho(candidate, op string) bool {
	candidate = strings.TrimSpace(candidate)
	op = strings.TrimSpace(op)
	if candidate == "" || op == "" || !strings.EqualFold(candidate, op) {
		return false
	}
	return utf8.RuneCountInString(candidate) <= 12
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "\r", " ")
}

func statusGlyph(s LogStatus) string {
	switch s {
	case StatusDone:
		return "✓"
	case StatusError:
		return "✖"
	case StatusWarn:
		return "!"
	case StatusWorking:
		return "●"
	default:
		return "·"
	}
}

func statusColor(s LogStatus) string {
	switch s {
	case StatusDone:
		return ansiGreen
	case StatusError:
		return ansiRed
	case StatusWarn:
		return ansiMagenta
	case StatusWorking:
		return ansiYellow
	default:
		return ansiCyan
	}
}

func kindColor(k LogKind) string {
	switch k {
	case KindPull, KindPush:
		return ansiBlue
	case KindBuild:
		return ansiMagenta
	case KindCompose:
		return ansiCyan
	case KindLog:
		return ansiDim
	default:
		return ansiGreen
	}
}

func paint(s, code string, on bool) string {
	if !on || s == "" {
		return s
	}
	return code + s + ansiReset
}

func progressBar(current, total int64, width int) string {
	if width < 4 {
		width = 4
	}
	if total <= 0 {
		return strings.Repeat("─", width)
	}
	filled := int(float64(current) / float64(total) * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func humanBytes(n int64) string {
	if n < 0 {
		n = 0
	}
	const k = 1024.0
	v := float64(n)
	switch {
	case v < k:
		return fmt.Sprintf("%d B", n)
	case v < k*k:
		return fmt.Sprintf("%.1f KB", v/k)
	case v < k*k*k:
		return fmt.Sprintf("%.1f MB", v/(k*k))
	default:
		return fmt.Sprintf("%.1f GB", v/(k*k*k))
	}
}

// visualRows is how many terminal rows line occupies at width.
// width <= 0 means the line is one row, which is what tests and non-TTYs use.
func visualRows(line string, width int) int {
	if width <= 0 {
		return 1
	}
	cols := visibleCols(line)
	if cols == 0 {
		return 1
	}
	return (cols + width - 1) / width
}

// visibleCols counts columns, skipping CSI color sequences.
func visibleCols(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && !((s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')) {
				i++
			}
			if i < len(s) {
				i++
			}
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		n++
	}
	return n
}

func trunc(s string, n int) string {
	if n <= 0 || utf8.RuneCountInString(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	r := []rune(s)
	// Keep both ends. Container names share a long prefix, and the service
	// is the tail: "Container astronomy-shop-email-1" stays distinguishable.
	tail := (n - 1) / 2
	head := n - 1 - tail
	return string(r[:head]) + "…" + string(r[len(r)-tail:])
}

func pad(s string, n int) string {
	w := utf8.RuneCountInString(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
