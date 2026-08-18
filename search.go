package logindex

import (
	"sort"

	"github.com/LYH2263/go-logindex/internal/analyze"
	"github.com/LYH2263/go-logindex/internal/collect"
	"github.com/LYH2263/go-logindex/internal/posting"
	"github.com/LYH2263/go-logindex/internal/query"
)

// Search 执行布尔查询；结果按分数降序、DocID 升序。
func (idx *Index) Search(q Query) ([]Hit, error) {
	if q == nil || q.IsEmpty() {
		return nil, ErrEmptyQuery
	}
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	if idx.closed {
		return nil, ErrClosed
	}
	q = query.Optimize(q)
	if q == nil || q.IsEmpty() {
		return nil, ErrEmptyQuery
	}
	src := idx.readSource()
	list, err := (query.Planner{Src: src}).Eval(q)
	if err != nil {
		return nil, err
	}
	// 仅保留活文档。
	live := src.Universe()
	list = posting.Intersect(list, live)

	var scorer collect.Scorer = collect.ConstScorer(1)
	if idx.tf != nil {
		terms := collectQueryTerms(q)
		per := make([]map[uint64]int, 0, len(terms))
		for _, term := range terms {
			m := make(map[uint64]int)
			for _, id := range list.Docs() {
				m[id] = idx.tf.DocTF(id, term)
			}
			per = append(per, m)
		}
		scorer = collect.MultiTermScorer{PerTerm: per}
	}
	col := collect.Collector{Scorer: scorer, Limit: idx.opts.MaxSearchHits}
	raw := col.Collect(list)
	out := make([]Hit, len(raw))
	for i, h := range raw {
		out[i] = Hit{DocID: h.DocID, Score: h.Score}
	}
	return out, nil
}

func collectQueryTerms(q *query.Query) []string {
	if q == nil {
		return nil
	}
	var out []string
	var walk func(*query.Query)
	walk = func(n *query.Query) {
		if n == nil {
			return
		}
		if n.Kind == query.KindTerm && n.Term != "" {
			out = append(out, n.Term)
		}
		for _, k := range n.Kids {
			walk(k)
		}
	}
	walk(q)
	return out
}

// SearchDocIDs 仅返回排序后的 DocID。
func (idx *Index) SearchDocIDs(q Query) ([]uint64, error) {
	hits, err := idx.Search(q)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, len(hits))
	for i, h := range hits {
		ids[i] = h.DocID
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

// BruteForceSearch 对内存中已知原文做暴力扫描（测试对照）。
func (idx *Index) BruteForceSearch(q Query) ([]uint64, error) {
	if q == nil || q.IsEmpty() {
		return nil, ErrEmptyQuery
	}
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	if idx.closed {
		return nil, ErrClosed
	}
	q = query.Optimize(q)
	var out []uint64
	for id, text := range idx.texts {
		terms := analyze.UniqueTerms(idx.opts.Analyzer, text)
		set := make(map[string]struct{}, len(terms))
		for _, t := range terms {
			set[t] = struct{}{}
		}
		if evalBrute(q, set) {
			out = append(out, id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

func evalBrute(q *query.Query, terms map[string]struct{}) bool {
	if q == nil {
		return false
	}
	switch q.Kind {
	case query.KindTerm:
		_, ok := terms[q.Term]
		return ok
	case query.KindAnd:
		for _, k := range q.Kids {
			if !evalBrute(k, terms) {
				return false
			}
		}
		return len(q.Kids) > 0
	case query.KindOr:
		for _, k := range q.Kids {
			if evalBrute(k, terms) {
				return true
			}
		}
		return false
	case query.KindNot:
		if len(q.Kids) == 0 {
			return false
		}
		return !evalBrute(q.Kids[0], terms)
	default:
		return false
	}
}
