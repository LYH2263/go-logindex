package stats

import (
	"sort"
	"sync"
)

// TermFreq 全局词频与文档频。
type TermFreq struct {
	mu       sync.RWMutex
	tf       map[string]int // 总出现次数
	df       map[string]int // 包含该词的文档数
	docTerms map[uint64]map[string]int
}

// NewTermFreq 构造。
func NewTermFreq() *TermFreq {
	return &TermFreq{
		tf:       make(map[string]int),
		df:       make(map[string]int),
		docTerms: make(map[uint64]map[string]int),
	}
}

// AddDoc 登记文档词频。
func (t *TermFreq) AddDoc(docID uint64, freqs map[string]int) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if old, ok := t.docTerms[docID]; ok {
		t.removeLocked(docID, old)
	}
	cp := make(map[string]int, len(freqs))
	for term, n := range freqs {
		if n <= 0 {
			continue
		}
		cp[term] = n
		t.tf[term] += n
		t.df[term]++
	}
	t.docTerms[docID] = cp
}

// RemoveDoc 删除文档统计。
func (t *TermFreq) RemoveDoc(docID uint64) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if old, ok := t.docTerms[docID]; ok {
		t.removeLocked(docID, old)
	}
}

func (t *TermFreq) removeLocked(docID uint64, freqs map[string]int) {
	for term, n := range freqs {
		t.tf[term] -= n
		if t.tf[term] <= 0 {
			delete(t.tf, term)
		}
		t.df[term]--
		if t.df[term] <= 0 {
			delete(t.df, term)
		}
	}
	delete(t.docTerms, docID)
}

// TF 返回全局词频。
func (t *TermFreq) TF(term string) int {
	if t == nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tf[term]
}

// DF 返回文档频。
func (t *TermFreq) DF(term string) int {
	if t == nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.df[term]
}

// DocTF 返回某文档内词频。
func (t *TermFreq) DocTF(docID uint64, term string) int {
	if t == nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if m, ok := t.docTerms[docID]; ok {
		return m[term]
	}
	return 0
}

// TopTerms 返回 TF 最高的 k 个词。
func (t *TermFreq) TopTerms(k int) []TermCount {
	if t == nil || k <= 0 {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]TermCount, 0, len(t.tf))
	for term, n := range t.tf {
		out = append(out, TermCount{Term: term, Count: n, DF: t.df[term]})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Term < out[j].Term
		}
		return out[i].Count > out[j].Count
	})
	if len(out) > k {
		out = out[:k]
	}
	return out
}

// TermCount 词项计数。
type TermCount struct {
	Term  string
	Count int
	DF    int
}

// NumDocs 已统计文档数。
func (t *TermFreq) NumDocs() int {
	if t == nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.docTerms)
}
