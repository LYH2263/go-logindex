package logindex

import (
	"fmt"
	"os"
	"sync"

	"github.com/LYH2263/go-logindex/internal/analyze"
	"github.com/LYH2263/go-logindex/internal/memindex"
	"github.com/LYH2263/go-logindex/internal/merge"
	"github.com/LYH2263/go-logindex/internal/segment"
	"github.com/LYH2263/go-logindex/internal/stats"
	"github.com/LYH2263/go-logindex/internal/walix"
)

// Index 日志倒排索引。
type Index struct {
	mu       sync.RWMutex
	root     string
	opts     Options
	mem      *memindex.Index
	segments []*segment.Segment
	catalog  *segment.Catalog
	wal      *walix.WAL
	tf       *stats.TermFreq
	ds       *stats.DocStats
	closed   bool
	// texts 保存已 flush 文档原文，供暴力对照与 NOT 全集；生产可外置。
	texts map[uint64]string
}

// Open 打开或创建索引目录。
func Open(root string, opts ...Option) (*Index, error) {
	if root == "" {
		return nil, fmt.Errorf("logindex: empty root")
	}
	cfg := applyOptions(opts)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	segs, cat, err := (segment.Reader{Root: root}).LoadAll()
	if err != nil {
		return nil, err
	}
	idx := &Index{
		root:     root,
		opts:     cfg,
		mem:      memindex.New(cfg.Analyzer),
		segments: segs,
		catalog:  cat,
		texts:    make(map[uint64]string),
	}
	if cfg.EnableStats {
		idx.tf = stats.NewTermFreq()
		idx.ds = stats.NewDocStats()
	}
	if cfg.EnableWAL {
		w, err := walix.Open(root)
		if err != nil {
			return nil, err
		}
		idx.wal = w
		if err := idx.recoverWAL(); err != nil {
			_ = w.Close()
			return nil, err
		}
	}
	// 从段重建 texts 不可用（段不存原文）；仅恢复 WAL 中的。
	return idx, nil
}

// Close 关闭索引。
func (idx *Index) Close() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if idx.closed {
		return nil
	}
	idx.closed = true
	if idx.wal != nil {
		return idx.wal.Close()
	}
	return nil
}

// Root 返回数据目录。
func (idx *Index) Root() string {
	return idx.root
}

// Add 写入文档；未 Flush 前对 Search 也可见（读视图含 mem）。
func (idx *Index) Add(docID uint64, text string) error {
	if docID == 0 {
		return ErrInvalidDocID
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if idx.closed {
		return ErrClosed
	}
	if idx.wal != nil {
		if err := idx.wal.Append(walix.Record{Op: walix.OpAdd, DocID: docID, Text: text}); err != nil {
			return err
		}
	}
	idx.mem.Add(docID, text)
	idx.texts[docID] = text
	// 从旧段墓碑：覆盖写入时删除旧段中同 ID。
	for _, seg := range idx.segments {
		seg.MarkDeleted(docID)
	}
	if idx.tf != nil {
		idx.tf.AddDoc(docID, analyze.TermFreq(idx.opts.Analyzer, text))
	}
	if idx.ds != nil {
		idx.ds.Set(docID, len(analyze.Terms(idx.opts.Analyzer, text)))
	}
	if idx.mem.DocCount() >= idx.opts.FlushThreshold {
		if err := idx.flushLocked(); err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除文档。
func (idx *Index) Delete(docID uint64) error {
	if docID == 0 {
		return ErrInvalidDocID
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if idx.closed {
		return ErrClosed
	}
	if idx.wal != nil {
		if err := idx.wal.Append(walix.Record{Op: walix.OpDelete, DocID: docID}); err != nil {
			return err
		}
	}
	idx.mem.Delete(docID)
	delete(idx.texts, docID)
	if idx.tf != nil {
		idx.tf.RemoveDoc(docID)
	}
	if idx.ds != nil {
		idx.ds.Remove(docID)
	}
	return nil
}

// Flush 将内存缓冲落成不可变段，并切换读视图。
func (idx *Index) Flush() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if idx.closed {
		return ErrClosed
	}
	return idx.flushLocked()
}

func (idx *Index) flushLocked() error {
	if idx.mem.DocCount() == 0 && len(idx.mem.PendingDeletes()) == 0 {
		return nil
	}
	if idx.wal != nil {
		if err := idx.wal.Append(walix.Record{Op: walix.OpFlush}); err != nil {
			return err
		}
	}
	id := idx.catalog.AllocID()
	seg, pending := idx.mem.FlushToSegment(id)
	if seg.Meta().DocCount > 0 {
		if err := seg.Persist(idx.root); err != nil {
			return err
		}
		idx.segments = append(idx.segments, seg)
		idx.catalog.Add(id)
		if err := segment.SaveCatalog(idx.root, idx.catalog); err != nil {
			return err
		}
	}
	for doc := range pending {
		delete(idx.texts, doc)
	}
	if idx.wal != nil {
		if err := idx.wal.Truncate(); err != nil {
			return err
		}
	}
	return idx.maybeMergeLocked()
}

func (idx *Index) maybeMergeLocked() error {
	if !idx.opts.MergePolicy.ShouldMerge(idx.segments) && !merge.NeedsCompaction(idx.opts.MergePolicy, idx.segments) {
		return nil
	}
	plan, ok := merge.BuildPlan(idx.opts.MergePolicy, idx.segments, idx.catalog.AllocID())
	if !ok {
		// 强制挑最小两段。
		srcs := idx.opts.MergePolicy.PickSources(idx.segments, 2)
		if len(srcs) < 2 {
			return nil
		}
		ids := make([]segment.ID, len(srcs))
		for i, s := range srcs {
			ids[i] = s.Meta().ID
		}
		plan = merge.Plan{Sources: ids, Target: idx.catalog.AllocID()}
	}
	parts := merge.SelectByIDs(idx.segments, plan.Sources)
	merged, err := (merge.Merger{}).Merge(plan.Target, parts)
	if err != nil {
		return err
	}
	if err := merged.Persist(idx.root); err != nil {
		return err
	}
	idx.segments = merge.ExcludeIDs(idx.segments, plan.Sources)
	idx.segments = append(idx.segments, merged)
	for _, id := range plan.Sources {
		idx.catalog.Remove(id)
		_ = segment.RemoveDir(idx.root, id)
	}
	idx.catalog.Add(plan.Target)
	// AllocID 已推进；确保 NextID 正确。
	if idx.catalog.NextID <= plan.Target {
		idx.catalog.NextID = plan.Target + 1
	}
	return segment.SaveCatalog(idx.root, idx.catalog)
}

func (idx *Index) recoverWAL() error {
	recs, err := walix.Replay(idx.wal.Path())
	if err != nil {
		return err
	}
	for _, r := range recs {
		switch r.Op {
		case walix.OpAdd:
			idx.mem.Add(r.DocID, r.Text)
			idx.texts[r.DocID] = r.Text
			for _, seg := range idx.segments {
				seg.MarkDeleted(r.DocID)
			}
			if idx.tf != nil {
				idx.tf.AddDoc(r.DocID, analyze.TermFreq(idx.opts.Analyzer, r.Text))
			}
			if idx.ds != nil {
				idx.ds.Set(r.DocID, len(analyze.Terms(idx.opts.Analyzer, r.Text)))
			}
		case walix.OpDelete:
			idx.mem.Delete(r.DocID)
			delete(idx.texts, r.DocID)
			for _, seg := range idx.segments {
				seg.MarkDeleted(r.DocID)
			}
			if idx.tf != nil {
				idx.tf.RemoveDoc(r.DocID)
			}
			if idx.ds != nil {
				idx.ds.Remove(docID(r.DocID))
			}
		case walix.OpFlush:
			// 崩溃前 flush 标记：内存仍应落盘。
			if err := idx.flushLocked(); err != nil {
				return err
			}
		}
	}
	return nil
}

func docID(id uint64) uint64 { return id }

// SegmentCount 返回段数量。
func (idx *Index) SegmentCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.segments)
}

// MemDocCount 内存文档数。
func (idx *Index) MemDocCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.mem.DocCount()
}

// StatsSnapshot 返回统计快照。
func (idx *Index) StatsSnapshot() stats.Snapshot {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return stats.Capture(idx.tf, idx.ds, 20)
}

// DocText 返回已知原文（若有）。
func (idx *Index) DocText(docID uint64) (string, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	t, ok := idx.texts[docID]
	return t, ok
}
