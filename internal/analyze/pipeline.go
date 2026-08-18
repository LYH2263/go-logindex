package analyze

// Pipeline 串联多个 Analyzer，后级吃前级输出的拼接文本。
type Pipeline struct {
	Stages []Analyzer
}

// Analyze 依次执行。
func (p Pipeline) Analyze(text string) []Token {
	if len(p.Stages) == 0 {
		return NewDefault().Analyze(text)
	}
	cur := text
	var last []Token
	for i, stage := range p.Stages {
		if stage == nil {
			continue
		}
		last = stage.Analyze(cur)
		if i == len(p.Stages)-1 {
			return last
		}
		// 中间级：用空格拼接再喂给下一级。
		parts := make([]string, len(last))
		for j, t := range last {
			parts[j] = t.Term
		}
		cur = joinSpace(parts)
	}
	return last
}

func joinSpace(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	n := 0
	for _, p := range parts {
		n += len(p) + 1
	}
	buf := make([]byte, 0, n)
	for i, p := range parts {
		if i > 0 {
			buf = append(buf, ' ')
		}
		buf = append(buf, p...)
	}
	return string(buf)
}

// Identity 原样按空白切分的分析器。
type Identity struct{}

// Analyze 使用空白分词 + 小写。
func (Identity) Analyze(text string) []Token {
	return New(WhitespaceTokenizer{}, LowercaseFilter{}).Analyze(text)
}

// CountTokens 统计 token 数。
func CountTokens(a Analyzer, text string) int {
	return len(Terms(a, text))
}

// HasTerm 判断文本是否含某词。
func HasTerm(a Analyzer, text, term string) bool {
	term = Normalize(term)
	for _, t := range Terms(a, text) {
		if t == term {
			return true
		}
	}
	return false
}
