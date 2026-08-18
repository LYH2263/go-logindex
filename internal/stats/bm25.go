package stats

import "math"

func logNatural(x float64) float64 {
	return math.Log(x)
}

// BM25 简易 BM25 分数。
func BM25(tf, df, numDocs, docLen int, avgLen, k1, b float64) float64 {
	if avgLen <= 0 {
		avgLen = 1
	}
	if k1 <= 0 {
		k1 = 1.2
	}
	if b < 0 {
		b = 0.75
	}
	idf := IDF(numDocs, df)
	tfF := float64(tf)
	dl := float64(docLen)
	denom := tfF + k1*(1-b+b*dl/avgLen)
	if denom == 0 {
		return 0
	}
	return idf * (tfF * (k1 + 1) / denom)
}

// Combine 合并两份 TermCount（同词相加）。
func Combine(a, b []TermCount) []TermCount {
	m := make(map[string]TermCount, len(a)+len(b))
	for _, t := range a {
		m[t.Term] = t
	}
	for _, t := range b {
		old := m[t.Term]
		old.Term = t.Term
		old.Count += t.Count
		old.DF += t.DF
		m[t.Term] = old
	}
	out := make([]TermCount, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
