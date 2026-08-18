package stats

import "sync"

// DocStats 文档级统计。
type DocStats struct {
	mu       sync.RWMutex
	lengths  map[uint64]int
	totalLen int64
}

// NewDocStats 构造。
func NewDocStats() *DocStats {
	return &DocStats{lengths: make(map[uint64]int)}
}

// Set 设置文档长度（token 数）。
func (d *DocStats) Set(docID uint64, length int) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if old, ok := d.lengths[docID]; ok {
		d.totalLen -= int64(old)
	}
	d.lengths[docID] = length
	d.totalLen += int64(length)
}

// Remove 删除。
func (d *DocStats) Remove(docID uint64) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if old, ok := d.lengths[docID]; ok {
		d.totalLen -= int64(old)
		delete(d.lengths, docID)
	}
}

// Length 文档长度。
func (d *DocStats) Length(docID uint64) int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lengths[docID]
}

// AvgLength 平均长度。
func (d *DocStats) AvgLength() float64 {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if len(d.lengths) == 0 {
		return 0
	}
	return float64(d.totalLen) / float64(len(d.lengths))
}

// Count 文档数。
func (d *DocStats) Count() int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.lengths)
}
