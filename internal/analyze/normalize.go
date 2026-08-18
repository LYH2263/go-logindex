package analyze

import "strings"

// Normalize 对词项做归一化（小写 + 去首尾空白）。
func Normalize(term string) string {
	return strings.ToLower(strings.TrimSpace(term))
}

// NormalizeAll 批量归一化并去掉空串。
func NormalizeAll(terms []string) []string {
	out := make([]string, 0, len(terms))
	for _, t := range terms {
		n := Normalize(t)
		if n == "" {
			continue
		}
		out = append(out, n)
	}
	return out
}

// IsIdentRune 判断是否为标识符字符。
func IsIdentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}

// SplitIdent 将类似 "foo_bar-baz" 的串拆成更细片段。
func SplitIdent(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var (
		out   []string
		start = -1
	)
	flush := func(end int) {
		if start < 0 {
			return
		}
		part := s[start:end]
		if part != "" {
			out = append(out, part)
		}
		start = -1
	}
	for i, r := range s {
		if IsIdentRune(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		flush(i)
	}
	flush(len(s))
	return out
}

// CamelSplit 拆分驼峰：ErrorCode → Error, Code。
func CamelSplit(s string) []string {
	if s == "" {
		return nil
	}
	runes := []rune(s)
	var (
		out   []string
		start int
	)
	isUpper := func(r rune) bool { return r >= 'A' && r <= 'Z' }
	for i := 1; i < len(runes); i++ {
		prev := runes[i-1]
		cur := runes[i]
		boundary := false
		if isUpper(cur) && !isUpper(prev) {
			boundary = true
		}
		if isUpper(cur) && i+1 < len(runes) && !isUpper(runes[i+1]) && isUpper(prev) {
			boundary = true
		}
		if boundary {
			out = append(out, string(runes[start:i]))
			start = i
		}
	}
	out = append(out, string(runes[start:]))
	return out
}
