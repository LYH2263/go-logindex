package posting

// Deltas 返回相邻差分序列（首元素为绝对值）。
func Deltas(l *List) []DocID {
	if l == nil || l.Len() == 0 {
		return nil
	}
	out := make([]DocID, l.Len())
	var prev DocID
	for i, id := range l.docs {
		out[i] = id - prev
		prev = id
	}
	return out
}

// FromDeltas 从差分还原。
func FromDeltas(deltas []DocID) *List {
	if len(deltas) == 0 {
		return New()
	}
	ids := make([]DocID, len(deltas))
	var prev DocID
	for i, d := range deltas {
		id := prev + d
		ids[i] = id
		prev = id
	}
	return FromSorted(ids)
}

// EstimateEncodedSize 粗估编码大小。
func EstimateEncodedSize(l *List) int {
	if l == nil || l.Len() == 0 {
		return 1
	}
	// 平均每 doc 约 2~3 字节 varint + 长度前缀。
	return 8 + l.Len()*3
}
