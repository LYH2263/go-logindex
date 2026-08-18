package logindex

import (
	"fmt"

	"github.com/LYH2263/go-logindex/internal/segment"
)

// Info 返回可读状态摘要。
func (idx *Index) Info() string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return fmt.Sprintf("root=%s segs=%d mem_docs=%d texts=%d",
		idx.root, len(idx.segments), idx.mem.DocCount(), len(idx.texts))
}

// SegmentMetas 返回各段元数据。
func (idx *Index) SegmentMetas() []segment.Meta {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]segment.Meta, 0, len(idx.segments))
	for _, s := range idx.segments {
		out = append(out, s.Meta())
	}
	return out
}

// Compact 手动触发合并（若段数足够）。
func (idx *Index) Compact() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if idx.closed {
		return ErrClosed
	}
	return idx.maybeMergeLocked()
}
