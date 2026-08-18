package query

import (
	"context"

	"github.com/LYH2263/go-logindex/internal/posting"
)

func evalOr(ctx context.Context, p Planner, kids []*Query) (*posting.List, error) {
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
	return posting.UnionMany(lists), nil
}
