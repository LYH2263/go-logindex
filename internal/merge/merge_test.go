package merge_test

import (
	"testing"

	"github.com/LYH2263/go-logindex/internal/merge"
	"github.com/LYH2263/go-logindex/internal/segment"
)

func TestMergeNoLossNoDup(t *testing.T) {
	b1 := segment.NewBuilder()
	b1.AddDoc(1, []string{"a", "x"})
	b1.AddDoc(2, []string{"b"})
	s1 := b1.Build(1)

	b2 := segment.NewBuilder()
	b2.AddDoc(3, []string{"a", "y"})
	b2.AddDoc(2, []string{"b"}) // 同 doc 出现在另一段（覆盖场景简化）
	s2 := b2.Build(2)

	out, err := (merge.Merger{}).Merge(9, []*segment.Segment{s1, s2})
	if err != nil {
		t.Fatal(err)
	}
	if err := merge.VerifyMerge([]*segment.Segment{s1, s2}, out); err != nil {
		t.Fatal(err)
	}
	if out.Posting("a").Len() != 2 {
		t.Fatalf("a=%v", out.Posting("a").Docs())
	}
}
