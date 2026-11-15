package kit

import (
	"fmt"
	"strings"
	"unicode"
)

func generateServiceName(s *Service) string {
	if s == nil {
		return "service"
	}
	if s.name.ok() {
		return s.name.val
	}
	if s.image.ok() {
		if n := sanitizeName(imageBasename(s.image.val)); n != "" {
			return n
		}
	}
	if s.build.ok() {
		b := s.build.val
		if b.target.ok() {
			if n := sanitizeName(b.target.val); n != "" {
				return n
			}
		}
		if b.source.ok() {
			if n := sanitizeName(b.source.val); n != "" && n != "." {
				return n
			}
		}
	}
	return "service"
}

func imageBasename(image string) string {
	ref := image
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		ref = ref[i+1:]
	}
	if i := strings.IndexAny(ref, "@:"); i >= 0 {
		ref = ref[:i]
	}
	return ref
}

func sanitizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == '_' || r == '-' || r == '.':
			if b.Len() == 0 || lastDash {
				continue
			}
			b.WriteByte('-')
			lastDash = true
		default:
			if b.Len() == 0 || lastDash {
				continue
			}
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return ""
	}
	if c := out[0]; c < 'a' || c > 'z' {
		if c < '0' || c > '9' {
			return ""
		}
		out = "s" + out
	}
	return out
}

func uniqueName(base string, taken func(string) bool) string {
	if base == "" {
		base = "service"
	}
	if !taken(base) {
		return base
	}
	for i := 2; ; i++ {
		n := fmt.Sprintf("%s-%d", base, i)
		if !taken(n) {
			return n
		}
	}
}
