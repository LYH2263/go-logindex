package posting

// Iterator 顺序遍历倒排列表。
type Iterator struct {
	docs []DocID
	i    int
}

// NewIterator 构造迭代器。
func NewIterator(l *List) *Iterator {
	if l == nil {
		return &Iterator{}
	}
	return &Iterator{docs: l.docs, i: 0}
}

// Reset 重置到开头。
func (it *Iterator) Reset() {
	if it != nil {
		it.i = 0
	}
}

// Next 前进并返回当前值；耗尽返回 false。
func (it *Iterator) Next() (DocID, bool) {
	if it == nil || it.i >= len(it.docs) {
		return 0, false
	}
	id := it.docs[it.i]
	it.i++
	return id, true
}

// Peek 查看下一值但不消费。
func (it *Iterator) Peek() (DocID, bool) {
	if it == nil || it.i >= len(it.docs) {
		return 0, false
	}
	return it.docs[it.i], true
}

// SeekGE 定位到第一个 >= id 的位置。
func (it *Iterator) SeekGE(id DocID) (DocID, bool) {
	if it == nil {
		return 0, false
	}
	for it.i < len(it.docs) && it.docs[it.i] < id {
		it.i++
	}
	return it.Peek()
}

// Remaining 剩余数量。
func (it *Iterator) Remaining() int {
	if it == nil {
		return 0
	}
	return len(it.docs) - it.i
}

// IntersectIter 双指针求交到回调。
func IntersectIter(a, b *List, fn func(DocID) bool) {
	ia, ib := NewIterator(a), NewIterator(b)
	for {
		da, oka := ia.Peek()
		db, okb := ib.Peek()
		if !oka || !okb {
			return
		}
		switch {
		case da == db:
			if !fn(da) {
				return
			}
			ia.Next()
			ib.Next()
		case da < db:
			ia.SeekGE(db)
		default:
			ib.SeekGE(da)
		}
	}
}
