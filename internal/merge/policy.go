package merge

import "github.com/LYH2263/go-logindex/internal/segment"

// ScoreSegment 评估段“合并优先级”（越小越先合并）。
func ScoreSegment(s *segment.Segment) int {
	if s == nil {
		return 0
	}
	m := s.Meta()
	// 小段优先；词项多也略提高优先级（便于压缩）。
	return m.DocCount*10 + m.TermCount
}

// NeedsCompaction 更细的触发条件。
func NeedsCompaction(p Policy, segs []*segment.Segment) bool {
	if p.ShouldMerge(segs) {
		return true
	}
	tiny := 0
	for _, s := range segs {
		if s.Meta().DocCount < 32 {
			tiny++
		}
	}
	return tiny >= p.MinMergeSegs && p.MinMergeSegs >= 2
}

// EstimateMergedDocs 估算合并后文档数（去重后近似为并集）。
func EstimateMergedDocs(segs []*segment.Segment) int {
	seen := make(map[uint64]struct{})
	for _, s := range segs {
		for d := range s.LiveDocs() {
			seen[d] = struct{}{}
		}
	}
	return len(seen)
}
