package posting

// Filter 按谓词过滤列表。
func Filter(l *List, keep func(DocID) bool) *List {
	if l == nil || l.Len() == 0 {
		return New()
	}
	out := make([]DocID, 0, l.Len())
	for _, id := range l.docs {
		if keep(id) {
			out = append(out, id)
		}
	}
	return FromSorted(out)
}

// Range 返回 [lo, hi] 闭区间内的文档。
func Range(l *List, lo, hi DocID) *List {
	if l == nil || l.Len() == 0 {
		return New()
	}
	out := make([]DocID, 0)
	for _, id := range l.docs {
		if id < lo {
			continue
		}
		if id > hi {
			break
		}
		out = append(out, id)
	}
	return FromSorted(out)
}

// Equal 比较两列表是否相等。
func Equal(a, b *List) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Len() != b.Len() {
		return false
	}
	for i := range a.docs {
		if a.docs[i] != b.docs[i] {
			return false
		}
	}
	return true
}

// EqualSet 忽略顺序比较（内部会排序语义已保证）。
func EqualSet(a, b *List) bool {
	return Equal(a, b)
}

// CountOverlap 计算交集大小。
func CountOverlap(a, b *List) int {
	return Intersect(a, b).Len()
}

// Prefetch 将列表复制到新底层数组（减少共享）。
func Prefetch(l *List) *List {
	return l.Clone()
}

// AppendSorted 假定 ids 有序，与现有列表归并。
func AppendSorted(l *List, ids []DocID) *List {
	if l == nil {
		return FromSorted(ids)
	}
	return Union(l, FromSorted(ids))
}
