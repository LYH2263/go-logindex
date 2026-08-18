package analyze

import (
	"strings"
	"unicode"
)

// TokenFilter 过滤或改写 token 序列。
type TokenFilter interface {
	Filter(tokens []Token) []Token
}

// LowercaseFilter 转小写。
type LowercaseFilter struct{}

// Filter 转小写。
func (LowercaseFilter) Filter(tokens []Token) []Token {
	out := make([]Token, len(tokens))
	for i, t := range tokens {
		t.Term = strings.ToLower(t.Term)
		out[i] = t
	}
	return out
}

// MinLengthFilter 丢弃过短词。
type MinLengthFilter struct {
	Min int
}

// Filter 按最小长度过滤。
func (f MinLengthFilter) Filter(tokens []Token) []Token {
	min := f.Min
	if min <= 0 {
		min = 1
	}
	out := make([]Token, 0, len(tokens))
	for _, t := range tokens {
		if utf8Len(t.Term) >= min {
			out = append(out, t)
		}
	}
	return out
}

// MaxLengthFilter 丢弃过长词。
type MaxLengthFilter struct {
	Max int
}

// Filter 按最大长度过滤。
func (f MaxLengthFilter) Filter(tokens []Token) []Token {
	if f.Max <= 0 {
		return tokens
	}
	out := make([]Token, 0, len(tokens))
	for _, t := range tokens {
		if utf8Len(t.Term) <= f.Max {
			out = append(out, t)
		}
	}
	return out
}

// StopwordFilter 丢弃停用词。
type StopwordFilter struct {
	Words map[string]struct{}
}

// Filter 去停用词。
func (f StopwordFilter) Filter(tokens []Token) []Token {
	if len(f.Words) == 0 {
		return tokens
	}
	out := make([]Token, 0, len(tokens))
	for _, t := range tokens {
		if _, hit := f.Words[t.Term]; hit {
			continue
		}
		out = append(out, t)
	}
	return out
}

// DedupAdjacentFilter 去掉相邻重复词。
type DedupAdjacentFilter struct{}

// Filter 去相邻重复。
func (DedupAdjacentFilter) Filter(tokens []Token) []Token {
	if len(tokens) == 0 {
		return tokens
	}
	out := make([]Token, 0, len(tokens))
	var prev string
	for i, t := range tokens {
		if i > 0 && t.Term == prev {
			continue
		}
		out = append(out, t)
		prev = t.Term
	}
	return out
}

// ASCIIFoldFilter 将常见重音字符折叠为 ASCII 近似。
type ASCIIFoldFilter struct{}

// Filter 做简易折叠。
func (ASCIIFoldFilter) Filter(tokens []Token) []Token {
	out := make([]Token, len(tokens))
	for i, t := range tokens {
		t.Term = foldASCII(t.Term)
		out[i] = t
	}
	return out
}

// PrefixFilter 仅保留给定前缀的词。
type PrefixFilter struct {
	Prefix string
}

// Filter 按前缀保留。
func (f PrefixFilter) Filter(tokens []Token) []Token {
	if f.Prefix == "" {
		return tokens
	}
	out := make([]Token, 0, len(tokens))
	for _, t := range tokens {
		if strings.HasPrefix(t.Term, f.Prefix) {
			out = append(out, t)
		}
	}
	return out
}

func utf8Len(s string) int {
	return len([]rune(s))
}

func foldASCII(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'À' && r <= 'Å':
			b.WriteByte('A')
		case r >= 'à' && r <= 'å':
			b.WriteByte('a')
		case r == 'Ç':
			b.WriteByte('C')
		case r == 'ç':
			b.WriteByte('c')
		case r >= 'È' && r <= 'Ë':
			b.WriteByte('E')
		case r >= 'è' && r <= 'ë':
			b.WriteByte('e')
		case r >= 'Ì' && r <= 'Ï':
			b.WriteByte('I')
		case r >= 'ì' && r <= 'ï':
			b.WriteByte('i')
		case r == 'Ñ':
			b.WriteByte('N')
		case r == 'ñ':
			b.WriteByte('n')
		case r >= 'Ò' && r <= 'Ö':
			b.WriteByte('O')
		case r >= 'ò' && r <= 'ö':
			b.WriteByte('o')
		case r >= 'Ù' && r <= 'Ü':
			b.WriteByte('U')
		case r >= 'ù' && r <= 'ü':
			b.WriteByte('u')
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_':
			b.WriteRune(r)
		default:
			// drop
		}
	}
	return b.String()
}
