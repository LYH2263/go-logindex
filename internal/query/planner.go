package query

import (
	"context"

	"github.com/LYH2263/go-logindex/internal/posting"
)

// PostingSource 为规划器提供词项倒排与全集。
type PostingSource interface {
	Posting(term string) *posting.List
	Universe() *posting.List
}

// Planner 将查询树求值为文档列表。
type Planner struct {
	Src PostingSource
}

// Eval 执行查询。
func (p Planner) Eval(q *Query) (*posting.List, error) {
	if q == nil || q.IsEmpty() {
		return posting.New(), nil
	}
	return p.eval(q)
}

// EvalContext 带取消的求值。
func (p Planner) EvalContext(ctx context.Context, q *Query) (*posting.List, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	return p.Eval(q)
}

func (p Planner) eval(q *Query) (*posting.List, error) {
	switch q.Kind {
	case KindTerm:
		return p.Src.Posting(q.Term), nil
	case KindAnd:
		return evalAnd(p, q.Kids)
	case KindOr:
		return evalOr(p, q.Kids)
	case KindNot:
		return evalNot(p, q.Kids)
	default:
		return posting.New(), nil
	}
}
