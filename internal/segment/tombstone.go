package segment

// TombstoneSet 汇总多个段的墓碑/删除集合。
type TombstoneSet struct {
	dead map[uint64]struct{}
}

// NewTombstoneSet 构造。
func NewTombstoneSet() *TombstoneSet {
	return &TombstoneSet{dead: make(map[uint64]struct{})}
}

// Add 标记删除。
func (t *TombstoneSet) Add(doc uint64) {
	if t == nil {
		return
	}
	t.dead[doc] = struct{}{}
}

// Contains 查询。
func (t *TombstoneSet) Contains(doc uint64) bool {
	if t == nil {
		return false
	}
	_, ok := t.dead[doc]
	return ok
}

// Len 数量。
func (t *TombstoneSet) Len() int {
	if t == nil {
		return 0
	}
	return len(t.dead)
}

// ApplyTo 从 posting 列表中剔除。
func (t *TombstoneSet) ApplyTo(docs []uint64) []uint64 {
	if t == nil || len(t.dead) == 0 {
		return docs
	}
	out := docs[:0]
	for _, d := range docs {
		if _, ok := t.dead[d]; ok {
			continue
		}
		out = append(out, d)
	}
	return out
}

// CloneMap 拷贝底层集合。
func (t *TombstoneSet) CloneMap() map[uint64]struct{} {
	if t == nil {
		return nil
	}
	out := make(map[uint64]struct{}, len(t.dead))
	for d := range t.dead {
		out[d] = struct{}{}
	}
	return out
}
