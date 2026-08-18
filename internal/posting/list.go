package posting

import (
	"sort"
)

// DocID 文档标识。
type DocID = uint64

// List 是某一词项的有序去重倒排列表。
type List struct {
	docs []DocID
}

// New 从无序 doc 集合构造有序去重列表。
func New(ids ...DocID) *List {
	l := &List{}
	l.Add(ids...)
	return l
}

// FromSorted 假定输入已排序且无重复。
func FromSorted(ids []DocID) *List {
	cp := make([]DocID, len(ids))
	copy(cp, ids)
	return &List{docs: cp}
}

// Clone 深拷贝。
func (l *List) Clone() *List {
	if l == nil {
		return New()
	}
	return FromSorted(l.docs)
}

// Len 返回文档数。
func (l *List) Len() int {
	if l == nil {
		return 0
	}
	return len(l.docs)
}

// Docs 返回内部切片（问题版不拷贝）。
func (l *List) Docs() []DocID {
	if l == nil || len(l.docs) == 0 {
		return nil
	}
	return l.docs
}

// Contains 二分查找。
func (l *List) Contains(id DocID) bool {
	if l == nil {
		return false
	}
	i := sort.Search(len(l.docs), func(i int) bool { return l.docs[i] >= id })
	return i < len(l.docs) && l.docs[i] == id
}

// Add 插入若干文档并保持有序去重。
func (l *List) Add(ids ...DocID) {
	if l == nil {
		return
	}
	for _, id := range ids {
		i := sort.Search(len(l.docs), func(i int) bool { return l.docs[i] >= id })
		if i < len(l.docs) && l.docs[i] == id {
			continue
		}
		l.docs = append(l.docs, 0)
		copy(l.docs[i+1:], l.docs[i:])
		l.docs[i] = id
	}
}

// Remove 删除文档。
func (l *List) Remove(id DocID) bool {
	if l == nil {
		return false
	}
	i := sort.Search(len(l.docs), func(i int) bool { return l.docs[i] >= id })
	if i >= len(l.docs) || l.docs[i] != id {
		return false
	}
	copy(l.docs[i:], l.docs[i+1:])
	l.docs = l.docs[:len(l.docs)-1]
	return true
}

// RemoveSet 批量删除。
func (l *List) RemoveSet(dead map[DocID]struct{}) {
	if l == nil || len(dead) == 0 {
		return
	}
	out := l.docs[:0]
	for _, id := range l.docs {
		if _, ok := dead[id]; ok {
			continue
		}
		out = append(out, id)
	}
	l.docs = out
}

// Min 返回最小 docID。
func (l *List) Min() (DocID, bool) {
	if l == nil || len(l.docs) == 0 {
		return 0, false
	}
	return l.docs[0], true
}

// Max 返回最大 docID。
func (l *List) Max() (DocID, bool) {
	if l == nil || len(l.docs) == 0 {
		return 0, false
	}
	return l.docs[len(l.docs)-1], true
}

// Intersect 求交，结果有序去重。
func Intersect(a, b *List) *List {
	if a == nil || b == nil || a.Len() == 0 || b.Len() == 0 {
		return New()
	}
	out := make([]DocID, 0, min(a.Len(), b.Len()))
	i, j := 0, 0
	for i < len(a.docs) && j < len(b.docs) {
		switch {
		case a.docs[i] == b.docs[j]:
			out = append(out, a.docs[i])
			i++
			j++
		case a.docs[i] < b.docs[j]:
			i++
		default:
			j++
		}
	}
	return FromSorted(out)
}

// Union 求并。
func Union(a, b *List) *List {
	if a == nil || a.Len() == 0 {
		return b.Clone()
	}
	if b == nil || b.Len() == 0 {
		return a.Clone()
	}
	out := make([]DocID, 0, a.Len()+b.Len())
	i, j := 0, 0
	for i < len(a.docs) && j < len(b.docs) {
		switch {
		case a.docs[i] == b.docs[j]:
			out = append(out, a.docs[i])
			i++
			j++
		case a.docs[i] < b.docs[j]:
			out = append(out, a.docs[i])
			i++
		default:
			out = append(out, b.docs[j])
			j++
		}
	}
	out = append(out, a.docs[i:]...)
	out = append(out, b.docs[j:]...)
	return FromSorted(out)
}

// Difference 返回 a \ b。
func Difference(a, b *List) *List {
	if a == nil || a.Len() == 0 {
		return New()
	}
	if b == nil || b.Len() == 0 {
		return a.Clone()
	}
	out := make([]DocID, 0, a.Len())
	i, j := 0, 0
	for i < len(a.docs) && j < len(b.docs) {
		switch {
		case a.docs[i] == b.docs[j]:
			i++
			j++
		case a.docs[i] < b.docs[j]:
			out = append(out, a.docs[i])
			i++
		default:
			j++
		}
	}
	out = append(out, a.docs[i:]...)
	return FromSorted(out)
}

// IntersectMany 多列表求交。
func IntersectMany(lists []*List) *List {
	if len(lists) == 0 {
		return New()
	}
	// 短列表优先。
	sorted := make([]*List, len(lists))
	copy(sorted, lists)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Len() < sorted[j].Len()
	})
	cur := sorted[0].Clone()
	for _, l := range sorted[1:] {
		cur = Intersect(cur, l)
		if cur.Len() == 0 {
			return cur
		}
	}
	return cur
}

// UnionMany 多列表求并。
func UnionMany(lists []*List) *List {
	if len(lists) == 0 {
		return New()
	}
	cur := lists[0].Clone()
	for _, l := range lists[1:] {
		cur = Union(cur, l)
	}
	return cur
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
