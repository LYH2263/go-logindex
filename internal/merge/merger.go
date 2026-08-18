package merge

import (
	"fmt"
	"sort"
	"time"

	"github.com/LYH2263/go-logindex/internal/posting"
	"github.com/LYH2263/go-logindex/internal/segment"
)

// Policy 控制何时合并。
type Policy struct {
	MaxSegments   int
	MaxDocsPerSeg int
	MinMergeSegs  int
}

// DefaultPolicy 默认策略。
func DefaultPolicy() Policy {
	return Policy{
		MaxSegments:   8,
		MaxDocsPerSeg: 10_000,
		MinMergeSegs:  2,
	}
}

// Plan 描述一次合并计划。
type Plan struct {
	Sources []segment.ID
	Target  segment.ID
}

// ShouldMerge 根据策略判断是否需要合并。
func (p Policy) ShouldMerge(segs []*segment.Segment) bool {
	if len(segs) >= p.MaxSegments && p.MaxSegments > 0 {
		return true
	}
	return false
}

// PickSources 选择待合并的段（优先小段）。
func (p Policy) PickSources(segs []*segment.Segment, n int) []*segment.Segment {
	if n <= 0 {
		n = p.MinMergeSegs
	}
	if n < 2 {
		n = 2
	}
	if len(segs) < n {
		return nil
	}
	cp := make([]*segment.Segment, len(segs))
	copy(cp, segs)
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].Meta().DocCount < cp[j].Meta().DocCount
	})
	if len(cp) > n {
		cp = cp[:n]
	}
	return cp
}

// Merger 执行段合并。
type Merger struct{}

// Merge 合并多个段为新段；同 term 同 doc 不丢不重。
func (Merger) Merge(id segment.ID, parts []*segment.Segment) (*segment.Segment, error) {
	if len(parts) == 0 {
		return segment.New(id, nil, nil, time.Now().Unix()), nil
	}
	live := make(map[uint64]struct{})
	termDocs := make(map[string]map[uint64]struct{})
	for _, seg := range parts {
		if seg == nil {
			continue
		}
		for d := range seg.LiveDocs() {
			live[d] = struct{}{}
		}
		for term, list := range seg.AllPostings() {
			set, ok := termDocs[term]
			if !ok {
				set = make(map[uint64]struct{})
				termDocs[term] = set
			}
			for _, d := range list.Docs() {
				live[d] = struct{}{}
				set[d] = struct{}{}
			}
		}
	}
	flat := make(map[string][]uint64, len(termDocs))
	for term, set := range termDocs {
		docs := make([]uint64, 0, len(set))
		for d := range set {
			docs = append(docs, d)
		}
		flat[term] = docs
	}
	liveSlice := make([]uint64, 0, len(live))
	for d := range live {
		liveSlice = append(liveSlice, d)
	}
	out := segment.New(id, flat, liveSlice, time.Now().Unix())
	if err := segment.ValidateNoDup(out); err != nil {
		return nil, err
	}
	// 额外校验：合并后每个源活 doc 仍在。
	for _, seg := range parts {
		for d := range seg.LiveDocs() {
			if !out.IsLive(d) {
				return nil, fmt.Errorf("merge: lost live doc %d", d)
			}
		}
	}
	return out, nil
}

// MergePostingMaps 合并多张倒排表。
func MergePostingMaps(maps []map[string]*posting.List) map[string]*posting.List {
	out := make(map[string]*posting.List)
	for _, m := range maps {
		for term, list := range m {
			if existing, ok := out[term]; ok {
				out[term] = posting.Union(existing, list)
			} else {
				out[term] = list.Clone()
			}
		}
	}
	return out
}
