package logindex

import (
	"github.com/LYH2263/go-logindex/internal/analyze"
	"github.com/LYH2263/go-logindex/internal/posting"
)

// readSource 聚合 mem + segments 的查询视图。
type readSource struct {
	idx *Index
}

func (idx *Index) readSource() readSource {
	return readSource{idx: idx}
}

// Posting 合并各层倒排。
func (r readSource) Posting(term string) *posting.List {
	term = analyze.Normalize(term)
	lists := make([]*posting.List, 0, 1+len(r.idx.segments))
	lists = append(lists, r.idx.mem.Posting(term))
	for _, seg := range r.idx.segments {
		lists = append(lists, seg.Posting(term))
	}
	return posting.UnionMany(lists)
}

// Universe 全部存活文档。
func (r readSource) Universe() *posting.List {
	ids := make([]uint64, 0)
	seen := make(map[uint64]struct{})
	for d := range r.idx.mem.LiveDocs() {
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		ids = append(ids, d)
	}
	for _, seg := range r.idx.segments {
		for d := range seg.LiveDocs() {
			if _, ok := seen[d]; ok {
				continue
			}
			seen[d] = struct{}{}
			ids = append(ids, d)
		}
	}
	return posting.New(ids...)
}
