package analyze

// Token 是分词结果。
type Token struct {
	Term     string
	Position int
	Start    int
	End      int
}

// Analyzer 将原始文本转为词项序列。
type Analyzer interface {
	Analyze(text string) []Token
}

// Default 是内置默认分析器：小写、按非字母数字切分、去停用词、可选词干近似。
type Default struct {
	tokenizer Tokenizer
	filters   []TokenFilter
}

// NewDefault 构造默认分析器。
func NewDefault() *Default {
	return &Default{
		tokenizer: NewSimpleTokenizer(),
		filters: []TokenFilter{
			LowercaseFilter{},
			MinLengthFilter{Min: 1},
			StopwordFilter{Words: DefaultEnglishStopwords()},
			DedupAdjacentFilter{},
		},
	}
}

// New 用自定义分词器与过滤器构造分析器。
func New(tok Tokenizer, filters ...TokenFilter) *Default {
	if tok == nil {
		tok = NewSimpleTokenizer()
	}
	return &Default{tokenizer: tok, filters: filters}
}

// Analyze 执行分词与过滤。
func (a *Default) Analyze(text string) []Token {
	if a == nil {
		return NewDefault().Analyze(text)
	}
	tokens := a.tokenizer.Tokenize(text)
	for _, f := range a.filters {
		if f == nil {
			continue
		}
		tokens = f.Filter(tokens)
	}
	// 重新编号 position，保证连续。
	out := make([]Token, 0, len(tokens))
	for i, t := range tokens {
		if t.Term == "" {
			continue
		}
		t.Position = i
		out = append(out, t)
	}
	return out
}

// Terms 仅返回词项字符串列表。
func Terms(a Analyzer, text string) []string {
	if a == nil {
		a = NewDefault()
	}
	toks := a.Analyze(text)
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		out = append(out, t.Term)
	}
	return out
}

// UniqueTerms 返回去重后的词项（保持首次出现顺序）。
func UniqueTerms(a Analyzer, text string) []string {
	terms := Terms(a, text)
	seen := make(map[string]struct{}, len(terms))
	out := make([]string, 0, len(terms))
	for _, t := range terms {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

// TermFreq 统计文本内词项出现次数。
func TermFreq(a Analyzer, text string) map[string]int {
	if a == nil {
		a = NewDefault()
	}
	freq := make(map[string]int)
	for _, t := range a.Analyze(text) {
		freq[t.Term]++
	}
	return freq
}
