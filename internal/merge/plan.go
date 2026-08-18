package merge

import "github.com/LYH2263/go-logindex/internal/segment"

// BuildPlan 根据策略生成合并计划。
func BuildPlan(p Policy, segs []*segment.Segment, nextID segment.ID) (Plan, bool) {
	if !p.ShouldMerge(segs) {
		return Plan{}, false
	}
	sources := p.PickSources(segs, p.MinMergeSegs)
	if len(sources) < 2 {
		return Plan{}, false
	}
	ids := make([]segment.ID, len(sources))
	for i, s := range sources {
		ids[i] = s.Meta().ID
	}
	return Plan{Sources: ids, Target: nextID}, true
}

// SelectByIDs 按 ID 挑段。
func SelectByIDs(segs []*segment.Segment, ids []segment.ID) []*segment.Segment {
	want := make(map[segment.ID]struct{}, len(ids))
	for _, id := range ids {
		want[id] = struct{}{}
	}
	var out []*segment.Segment
	for _, s := range segs {
		if _, ok := want[s.Meta().ID]; ok {
			out = append(out, s)
		}
	}
	return out
}

// ExcludeIDs 排除若干段。
func ExcludeIDs(segs []*segment.Segment, ids []segment.ID) []*segment.Segment {
	drop := make(map[segment.ID]struct{}, len(ids))
	for _, id := range ids {
		drop[id] = struct{}{}
	}
	out := make([]*segment.Segment, 0, len(segs))
	for _, s := range segs {
		if _, ok := drop[s.Meta().ID]; ok {
			continue
		}
		out = append(out, s)
	}
	return out
}
