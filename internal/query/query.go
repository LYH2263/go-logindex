package query

import (
	"fmt"
	"strings"

	"github.com/LYH2263/go-logindex/internal/analyze"
)

// Kind 查询节点类型。
type Kind int

const (
	KindTerm Kind = iota
	KindAnd
	KindOr
	KindNot
)

// Query 布尔查询树。
type Query struct {
	Kind  Kind
	Term  string
	Kids  []*Query
}

// Term 词项查询。
func Term(term string) *Query {
	return &Query{Kind: KindTerm, Term: analyze.Normalize(term)}
}

// And 合取。
func And(parts ...*Query) *Query {
	kids := flatten(KindAnd, parts)
	return &Query{Kind: KindAnd, Kids: kids}
}

// Or 析取。
func Or(parts ...*Query) *Query {
	kids := flatten(KindOr, parts)
	return &Query{Kind: KindOr, Kids: kids}
}

// Not 否定。
func Not(inner *Query) *Query {
	return &Query{Kind: KindNot, Kids: []*Query{inner}}
}

func flatten(kind Kind, parts []*Query) []*Query {
	out := make([]*Query, 0, len(parts))
	for _, p := range parts {
		if p == nil {
			continue
		}
		if p.Kind == kind {
			out = append(out, p.Kids...)
			continue
		}
		out = append(out, p)
	}
	return out
}

// String 调试串。
func (q *Query) String() string {
	if q == nil {
		return "<nil>"
	}
	switch q.Kind {
	case KindTerm:
		return fmt.Sprintf("%q", q.Term)
	case KindAnd:
		return "AND(" + joinKids(q.Kids) + ")"
	case KindOr:
		return "OR(" + joinKids(q.Kids) + ")"
	case KindNot:
		if len(q.Kids) == 1 {
			return "NOT(" + q.Kids[0].String() + ")"
		}
		return "NOT(?)"
	default:
		return "?"
	}
}

func joinKids(kids []*Query) string {
	parts := make([]string, 0, len(kids))
	for _, k := range kids {
		parts = append(parts, k.String())
	}
	return strings.Join(parts, ", ")
}

// IsEmpty 空查询。
func (q *Query) IsEmpty() bool {
	if q == nil {
		return true
	}
	switch q.Kind {
	case KindTerm:
		return q.Term == ""
	case KindAnd, KindOr:
		return len(q.Kids) == 0
	case KindNot:
		return len(q.Kids) == 0 || q.Kids[0] == nil
	default:
		return true
	}
}
