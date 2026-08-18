package collect

import (
	"sort"

	"github.com/LYH2263/go-logindex/internal/posting"
)

// Hit 命中结果。
type Hit struct {
	DocID uint64
	Score float64
}

// Collector 收集并可选评分。
type Collector struct {
	Scorer Scorer
	Limit  int
}

// Collect 从 posting 列表生成命中。
func (c Collector) Collect(list *posting.List) []Hit {
	if list == nil || list.Len() == 0 {
		return nil
	}
	hits := make([]Hit, 0, list.Len())
	scorer := c.Scorer
	if scorer == nil {
		scorer = ConstScorer(1)
	}
	// 先对全集评分，再排序，最后才截断 Limit。
	for _, id := range list.Docs() {
		hits = append(hits, Hit{DocID: id, Score: scorer.Score(id)})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].DocID < hits[j].DocID
		}
		return hits[i].Score > hits[j].Score
	})
	if c.Limit > 0 && len(hits) > c.Limit {
		hits = hits[:c.Limit]
	}
	return hits
}
