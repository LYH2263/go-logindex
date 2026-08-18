package query_test

import (
	"reflect"
	"testing"

	"github.com/LYH2263/go-logindex/internal/posting"
	"github.com/LYH2263/go-logindex/internal/query"
)

type mapSrc struct {
	m map[string]*posting.List
	u *posting.List
}

func (s mapSrc) Posting(term string) *posting.List {
	if l, ok := s.m[term]; ok {
		return l.Clone()
	}
	return posting.New()
}
func (s mapSrc) Universe() *posting.List { return s.u.Clone() }

func TestPlannerAndOrNot(t *testing.T) {
	src := mapSrc{
		m: map[string]*posting.List{
			"a": posting.New(1, 2, 3),
			"b": posting.New(2, 3, 4),
			"c": posting.New(1, 5),
		},
		u: posting.New(1, 2, 3, 4, 5),
	}
	p := query.Planner{Src: src}

	got, err := p.Eval(query.And(query.Term("a"), query.Term("b")))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Docs(), []uint64{2, 3}) {
		t.Fatalf("and %v", got.Docs())
	}
	got, _ = p.Eval(query.Or(query.Term("a"), query.Term("c")))
	if !reflect.DeepEqual(got.Docs(), []uint64{1, 2, 3, 5}) {
		t.Fatalf("or %v", got.Docs())
	}
	got, _ = p.Eval(query.Not(query.Term("a")))
	if !reflect.DeepEqual(got.Docs(), []uint64{4, 5}) {
		t.Fatalf("not %v", got.Docs())
	}
}

func TestParseSimple(t *testing.T) {
	q, err := query.ParseSimple("error AND (timeout OR disk)")
	if err != nil {
		t.Fatal(err)
	}
	if q.Kind != query.KindAnd {
		t.Fatalf("kind %v", q.Kind)
	}
}
