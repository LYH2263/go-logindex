package memindex

// BufferStats 缓冲统计。
type BufferStats struct {
	Docs      int
	Terms     int
	Deletes   int
	ApproxBytes int
}

// Stats 返回粗略统计。
func (m *Index) Stats() BufferStats {
	if m == nil {
		return BufferStats{}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	bytes := 0
	for term, set := range m.terms {
		bytes += len(term) + 8*len(set)
	}
	for _, text := range m.docs {
		bytes += len(text)
	}
	return BufferStats{
		Docs:        len(m.docs),
		Terms:       len(m.terms),
		Deletes:     len(m.deleted),
		ApproxBytes: bytes,
	}
}

// HasDoc 是否包含文档。
func (m *Index) HasDoc(docID uint64) bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.docs[docID]
	return ok
}

// Text 返回原文（若存在）。
func (m *Index) Text(docID uint64) (string, bool) {
	if m == nil {
		return "", false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.docs[docID]
	return t, ok
}
