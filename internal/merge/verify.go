package merge

import (
	"fmt"

	"github.com/LYH2263/go-logindex/internal/posting"
	"github.com/LYH2263/go-logindex/internal/segment"
)

// VerifyMerge 校验合并结果相对源段：活 doc 不丢；同 term 同 doc 不重。
func VerifyMerge(src []*segment.Segment, out *segment.Segment) error {
	if out == nil {
		return fmt.Errorf("merge verify: nil out")
	}
	if err := segment.ValidateNoDup(out); err != nil {
		return err
	}
	srcLive := make(map[uint64]struct{})
	for _, s := range src {
		for d := range s.LiveDocs() {
			srcLive[d] = struct{}{}
			if !out.IsLive(d) {
				return fmt.Errorf("merge verify: missing live doc %d", d)
			}
		}
	}
	for d := range out.LiveDocs() {
		if _, ok := srcLive[d]; !ok {
			return fmt.Errorf("merge verify: unexpected doc %d", d)
		}
	}
	// 对每个源词项，合并后 posting 应等于各源并集。
	termSet := make(map[string]struct{})
	for _, s := range src {
		for _, t := range s.Terms() {
			termSet[t] = struct{}{}
		}
	}
	for term := range termSet {
		var lists []*posting.List
		for _, s := range src {
			lists = append(lists, s.Posting(term))
		}
		want := posting.UnionMany(lists)
		got := out.Posting(term)
		if !posting.Equal(want, got) {
			return fmt.Errorf("merge verify: term %q posting mismatch want=%v got=%v",
				term, want.Docs(), got.Docs())
		}
	}
	return nil
}

// DryRun 估算合并后元数据。
func DryRun(src []*segment.Segment) segment.Meta {
	docs := EstimateMergedDocs(src)
	terms := make(map[string]struct{})
	var minDoc, maxDoc uint64
	first := true
	for _, s := range src {
		for _, t := range s.Terms() {
			terms[t] = struct{}{}
		}
		m := s.Meta()
		if m.DocCount == 0 {
			continue
		}
		if first || m.MinDoc < minDoc {
			minDoc = m.MinDoc
		}
		if first || m.MaxDoc > maxDoc {
			maxDoc = m.MaxDoc
		}
		first = false
	}
	return segment.Meta{
		DocCount:  docs,
		MinDoc:    minDoc,
		MaxDoc:    maxDoc,
		TermCount: len(terms),
	}
}
