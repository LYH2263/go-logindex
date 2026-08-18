package segment

import "sort"

// Set 管理一组段指针。
type Set struct {
	items []*Segment
}

// NewSet 构造。
func NewSet(segs ...*Segment) *Set {
	s := &Set{}
	for _, seg := range segs {
		if seg != nil {
			s.items = append(s.items, seg)
		}
	}
	return s
}

// Add 追加。
func (s *Set) Add(seg *Segment) {
	if s == nil || seg == nil {
		return
	}
	s.items = append(s.items, seg)
}

// All 返回拷贝。
func (s *Set) All() []*Segment {
	if s == nil {
		return nil
	}
	out := make([]*Segment, len(s.items))
	copy(out, s.items)
	return out
}

// Len 数量。
func (s *Set) Len() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

// TotalDocs 存活文档总数（跨段可能重复 ID，按并集计）。
func (s *Set) TotalDocs() int {
	if s == nil {
		return 0
	}
	seen := make(map[uint64]struct{})
	for _, seg := range s.items {
		for d := range seg.LiveDocs() {
			seen[d] = struct{}{}
		}
	}
	return len(seen)
}

// IDs 段 ID 列表。
func (s *Set) IDs() []ID {
	if s == nil {
		return nil
	}
	out := make([]ID, len(s.items))
	for i, seg := range s.items {
		out[i] = seg.Meta().ID
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Find 按 ID 查找。
func (s *Set) Find(id ID) *Segment {
	if s == nil {
		return nil
	}
	for _, seg := range s.items {
		if seg.Meta().ID == id {
			return seg
		}
	}
	return nil
}

// Remove 移除 ID。
func (s *Set) Remove(id ID) {
	if s == nil {
		return
	}
	out := s.items[:0]
	for _, seg := range s.items {
		if seg.Meta().ID == id {
			continue
		}
		out = append(out, seg)
	}
	s.items = out
}
