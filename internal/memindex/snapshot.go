package memindex

import "github.com/LYH2263/go-logindex/internal/posting"

// Snapshot 是内存索引的不可变视图。
type Snapshot struct {
	Terms map[string]*posting.List
	Live  map[uint64]struct{}
}

// Posting 查询词项。
func (s *Snapshot) Posting(term string) *posting.List {
	if s == nil || s.Terms == nil {
		return posting.New()
	}
	if l, ok := s.Terms[term]; ok {
		return l.Clone()
	}
	return posting.New()
}

// DocCount 文档数。
func (s *Snapshot) DocCount() int {
	if s == nil {
		return 0
	}
	return len(s.Live)
}

// IsLive 判断。
func (s *Snapshot) IsLive(doc uint64) bool {
	if s == nil {
		return false
	}
	_, ok := s.Live[doc]
	return ok
}
