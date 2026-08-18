package collect

// Scorer 为文档打分。
type Scorer interface {
	Score(docID uint64) float64
}

// ConstScorer 常数分。
type ConstScorer float64

// Score 返回常数。
func (c ConstScorer) Score(uint64) float64 { return float64(c) }

// TFMapScorer 使用预先算好的 TF 作分。
type TFMapScorer map[uint64]float64

// Score 查表，缺失为 0。
func (m TFMapScorer) Score(docID uint64) float64 {
	if m == nil {
		return 0
	}
	return m[docID]
}

// IDFBoostScorer 简易 idf * tf。
type IDFBoostScorer struct {
	TF   map[uint64]float64
	IDF  float64
	Base float64
}

// Score 计算。
func (s IDFBoostScorer) Score(docID uint64) float64 {
	tf := 0.0
	if s.TF != nil {
		tf = s.TF[docID]
	}
	base := s.Base
	if base == 0 {
		base = 1
	}
	return base + tf*s.IDF
}

// MultiTermScorer 对多个查询词累加 TF。
type MultiTermScorer struct {
	PerTerm []map[uint64]int
}

// Score 累加。
func (m MultiTermScorer) Score(docID uint64) float64 {
	sum := 0.0
	for _, tf := range m.PerTerm {
		if tf == nil {
			continue
		}
		sum += float64(tf[docID])
	}
	return sum
}
