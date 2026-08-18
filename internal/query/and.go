package query

import (
	"context"
	"sort"

	"github.com/LYH2263/go-logindex/internal/posting"
)

func evalAnd(ctx context.Context, p Planner, kids []*Query) (*posting.List, error) {
	if len(kids) == 0 {
		return posting.New(), nil
	}
	lists := make([]*posting.List, 0, len(kids))
	for _, k := range kids {
		l, err := p.eval(ctx, k)
		if err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	// 短列表优先，加速求交。
	sort.Slice(lists, func(i, j int) bool { return lists[i].Len() < lists[j].Len() })
	return posting.IntersectMany(lists), nil
}
