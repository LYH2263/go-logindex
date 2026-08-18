package logindex_test

import (
	"testing"

	"github.com/LYH2263/go-logindex/internal/merge"
	"github.com/LYH2263/go-logindex/internal/segment"
)

func TestBug05_MergeDropsTombstones(t *testing.T) {
	b1 := segment.NewBuilder()
	b1.AddDoc(1, []string{"error"})
	b1.AddDoc(2, []string{"error"})
	s1 := b1.Build(1)
	if !s1.MarkDeleted(2) {
		t.Fatal("mark")
	}
	b2 := segment.NewBuilder()
	b2.AddDoc(3, []string{"error"})
	s2 := b2.Build(2)
	out, err := (merge.Merger{}).Merge(9, []*segment.Segment{s1, s2})
	if err != nil {
		t.Fatal(err)
	}
	if out.IsLive(2) {
		t.Fatal("merge must drop tombstoned docs from live set")
	}
	if out.Posting("error").Contains(2) {
		t.Fatalf("merge must drop dead postings, got %v", out.Posting("error").Docs())
	}
}
