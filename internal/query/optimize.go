package query

import (
	"sort"
)

// Optimize 做简单重写：去掉空孩子、单孩子提升、常量折叠。
func Optimize(q *Query) *Query {
	if q == nil {
		return nil
	}
	switch q.Kind {
	case KindTerm:
		if q.Term == "" {
			return nil
		}
		return Term(q.Term)
	case KindNot:
		if len(q.Kids) == 0 {
			return nil
		}
		inner := Optimize(q.Kids[0])
		if inner == nil {
			return nil
		}
		// NOT(NOT(x)) => x
		if inner.Kind == KindNot && len(inner.Kids) == 1 {
			return Optimize(inner.Kids[0])
		}
		return Not(inner)
	case KindAnd, KindOr:
		kids := make([]*Query, 0, len(q.Kids))
		for _, k := range q.Kids {
			ok := Optimize(k)
			if ok == nil {
				continue
			}
			kids = append(kids, ok)
		}
		if len(kids) == 0 {
			return nil
		}
		if len(kids) == 1 {
			return kids[0]
		}
		if q.Kind == KindAnd {
			return And(kids...)
		}
		return Or(kids...)
	default:
		return q
	}
}

// CostEstimate 粗估执行代价（文档触达数近似）。
func CostEstimate(q *Query, src PostingSource) int {
	if q == nil {
		return 0
	}
	switch q.Kind {
	case KindTerm:
		return src.Posting(q.Term).Len()
	case KindAnd:
		if len(q.Kids) == 0 {
			return 0
		}
		costs := make([]int, len(q.Kids))
		for i, k := range q.Kids {
			costs[i] = CostEstimate(k, src)
		}
		sort.Ints(costs)
		return costs[0]
	case KindOr:
		sum := 0
		for _, k := range q.Kids {
			sum += CostEstimate(k, src)
		}
		return sum
	case KindNot:
		u := 0
		if src != nil && src.Universe() != nil {
			u = src.Universe().Len()
		}
		inner := 0
		if len(q.Kids) > 0 {
			inner = CostEstimate(q.Kids[0], src)
		}
		if u > inner {
			return u - inner
		}
		return u
	default:
		return 0
	}
}
