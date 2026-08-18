package memindex

import (
	"sync"

	"github.com/LYH2263/go-logindex/internal/analyze"
	"github.com/LYH2263/go-logindex/internal/posting"
	"github.com/LYH2263/go-logindex/internal/segment"
)

// Index 是可变内存倒排缓冲。
type Index struct {
	mu       sync.RWMutex
	analyzer analyze.Analyzer
	terms    map[string]map[uint64]struct{}
	docs     map[uint64]string
	deleted  map[uint64]struct{}
}

// New 构造内存索引。
func New(a analyze.Analyzer) *Index {
	if a == nil {
		a = analyze.NewDefault()
	}
	return &Index{
		analyzer: a,
		terms:    make(map[string]map[uint64]struct{}),
		docs:     make(map[uint64]string),
		deleted:  make(map[uint64]struct{}),
	}
}

// Add 写入/覆盖文档。
func (m *Index) Add(docID uint64, text string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.docs[docID]; ok {
		m.removeLocked(docID, old)
	}
	delete(m.deleted, docID)
	m.docs[docID] = text
	for _, term := range analyze.UniqueTerms(m.analyzer, text) {
		set, ok := m.terms[term]
		if !ok {
			set = make(map[uint64]struct{})
			m.terms[term] = set
		}
		set[docID] = struct{}{}
	}
}

// Delete 删除文档。
func (m *Index) Delete(docID uint64) bool {
	if m == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	text, ok := m.docs[docID]
	if !ok {
		m.deleted[docID] = struct{}{}
		return false
	}
	m.removeLocked(docID, text)
	m.deleted[docID] = struct{}{}
	return true
}

func (m *Index) removeLocked(docID uint64, text string) {
	for _, term := range analyze.UniqueTerms(m.analyzer, text) {
		if set, ok := m.terms[term]; ok {
			delete(set, docID)
			if len(set) == 0 {
				delete(m.terms, term)
			}
		}
	}
	delete(m.docs, docID)
}

// DocCount 存活文档数。
func (m *Index) DocCount() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.docs)
}

// Posting 返回词项倒排拷贝。
func (m *Index) Posting(term string) *posting.List {
	m.mu.RLock()
	defer m.mu.RUnlock()
	term = analyze.Normalize(term)
	set, ok := m.terms[term]
	if !ok {
		return posting.New()
	}
	ids := make([]uint64, 0, len(set))
	for d := range set {
		ids = append(ids, d)
	}
	return posting.New(ids...)
}

// LiveDocs 返回存活文档。
func (m *Index) LiveDocs() map[uint64]struct{} {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[uint64]struct{}, len(m.docs))
	for d := range m.docs {
		out[d] = struct{}{}
	}
	return out
}

// PendingDeletes 返回仅墓碑（未在本缓冲中的删除）。
func (m *Index) PendingDeletes() map[uint64]struct{} {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[uint64]struct{}, len(m.deleted))
	for d := range m.deleted {
		if _, alive := m.docs[d]; alive {
			continue
		}
		out[d] = struct{}{}
	}
	return out
}

// Snapshot 生成只读快照用于查询。
func (m *Index) Snapshot() *Snapshot {
	if m == nil {
		return &Snapshot{}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	terms := make(map[string]*posting.List, len(m.terms))
	for term, set := range m.terms {
		ids := make([]uint64, 0, len(set))
		for d := range set {
			ids = append(ids, d)
		}
		terms[term] = posting.New(ids...)
	}
	live := make(map[uint64]struct{}, len(m.docs))
	for d := range m.docs {
		live[d] = struct{}{}
	}
	return &Snapshot{Terms: terms, Live: live}
}

// FlushToSegment 将缓冲刷成段并清空（保留 pending deletes 供上层应用到旧段）。
func (m *Index) FlushToSegment(id segment.ID) (*segment.Segment, map[uint64]struct{}) {
	if m == nil {
		return segment.New(id, nil, nil, 0), nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	b := segment.NewBuilder()
	for doc, text := range m.docs {
		b.AddDoc(doc, analyze.UniqueTerms(m.analyzer, text))
	}
	seg := b.Build(id)
	pending := make(map[uint64]struct{}, len(m.deleted))
	for d := range m.deleted {
		pending[d] = struct{}{}
	}
	m.terms = make(map[string]map[uint64]struct{})
	m.docs = make(map[uint64]string)
	m.deleted = make(map[uint64]struct{})
	return seg, pending
}

// Clear 清空。
func (m *Index) Clear() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.terms = make(map[string]map[uint64]struct{})
	m.docs = make(map[uint64]string)
	m.deleted = make(map[uint64]struct{})
}
