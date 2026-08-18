package analyze

import (
	"unicode"
	"unicode/utf8"
)

// Tokenizer 将文本切成原始 token。
type Tokenizer interface {
	Tokenize(text string) []Token
}

// SimpleTokenizer 按非字母数字 rune 切分，保留偏移。
type SimpleTokenizer struct {
	KeepUnicodeLetters bool
	KeepDigits         bool
	KeepUnderscore     bool
}

// NewSimpleTokenizer 返回面向日志的简单分词器。
func NewSimpleTokenizer() *SimpleTokenizer {
	return &SimpleTokenizer{
		KeepUnicodeLetters: true,
		KeepDigits:         true,
		KeepUnderscore:     true,
	}
}

// Tokenize 切分文本。
func (t *SimpleTokenizer) Tokenize(text string) []Token {
	if text == "" {
		return nil
	}
	if t == nil {
		t = NewSimpleTokenizer()
	}
	var (
		out      []Token
		start    = -1
		pos      int
		bytePos  int
	)
	flush := func(endByte int) {
		if start < 0 {
			return
		}
		term := text[start:endByte]
		out = append(out, Token{
			Term:     term,
			Position: pos,
			Start:    start,
			End:      endByte,
		})
		pos++
		start = -1
	}
	for bytePos < len(text) {
		r, size := utf8.DecodeRuneInString(text[bytePos:])
		if t.isWordRune(r) {
			if start < 0 {
				start = bytePos
			}
		} else {
			flush(bytePos)
		}
		bytePos += size
	}
	flush(len(text))
	return out
}

func (t *SimpleTokenizer) isWordRune(r rune) bool {
	if r == utf8.RuneError {
		return false
	}
	if t.KeepUnderscore && r == '_' {
		return true
	}
	if t.KeepDigits && unicode.IsDigit(r) {
		return true
	}
	if t.KeepUnicodeLetters && unicode.IsLetter(r) {
		return true
	}
	return false
}

// WhitespaceTokenizer 仅按空白切分。
type WhitespaceTokenizer struct{}

// Tokenize 按空白切分。
func (WhitespaceTokenizer) Tokenize(text string) []Token {
	var (
		out   []Token
		start = -1
		pos   int
	)
	flush := func(end int) {
		if start < 0 {
			return
		}
		out = append(out, Token{Term: text[start:end], Position: pos, Start: start, End: end})
		pos++
		start = -1
	}
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		if unicode.IsSpace(r) {
			flush(i)
		} else if start < 0 {
			start = i
		}
		i += size
	}
	flush(len(text))
	return out
}

// NGramTokenizer 对字母数字序列生成字符 n-gram。
type NGramTokenizer struct {
	MinN int
	MaxN int
}

// NewNGramTokenizer 构造 n-gram 分词器。
func NewNGramTokenizer(minN, maxN int) *NGramTokenizer {
	if minN < 1 {
		minN = 1
	}
	if maxN < minN {
		maxN = minN
	}
	return &NGramTokenizer{MinN: minN, MaxN: maxN}
}

// Tokenize 生成 n-gram。
func (n *NGramTokenizer) Tokenize(text string) []Token {
	base := NewSimpleTokenizer().Tokenize(text)
	if n == nil {
		return base
	}
	var out []Token
	pos := 0
	for _, tok := range base {
		runes := []rune(tok.Term)
		L := len(runes)
		for size := n.MinN; size <= n.MaxN; size++ {
			if size > L {
				break
			}
			for i := 0; i+size <= L; i++ {
				gram := string(runes[i : i+size])
				out = append(out, Token{
					Term:     gram,
					Position: pos,
					Start:    tok.Start,
					End:      tok.End,
				})
				pos++
			}
		}
	}
	return out
}
