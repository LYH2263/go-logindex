package collect

import "sort"

// DedupHits 按 DocID 去重，保留最高分。
func DedupHits(hits []Hit) []Hit {
	if len(hits) == 0 {
		return nil
	}
	best := make(map[uint64]float64, len(hits))
	for _, h := range hits {
		if old, ok := best[h.DocID]; !ok || h.Score > old {
			best[h.DocID] = h.Score
		}
	}
	out := make([]Hit, 0, len(best))
	for id, score := range best {
		out = append(out, Hit{DocID: id, Score: score})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].DocID < out[j].DocID
		}
		return out[i].Score > out[j].Score
	})
	return out
}

// DocIDs 提取 ID 列表。
func DocIDs(hits []Hit) []uint64 {
	out := make([]uint64, len(hits))
	for i, h := range hits {
		out[i] = h.DocID
	}
	return out
}

// SortByDocID 按文档号排序。
func SortByDocID(hits []Hit) []Hit {
	out := make([]Hit, len(hits))
	copy(out, hits)
	sort.Slice(out, func(i, j int) bool { return out[i].DocID < out[j].DocID })
	return out
}

// FilterLive 仅保留活文档。
func FilterLive(hits []Hit, live map[uint64]struct{}) []Hit {
	if len(live) == 0 {
		return hits
	}
	out := make([]Hit, 0, len(hits))
	for _, h := range hits {
		if _, ok := live[h.DocID]; ok {
			out = append(out, h)
		}
	}
	return out
}
