package memindex_test

import (
	"testing"

	"github.com/LYH2263/go-logindex/internal/analyze"
	"github.com/LYH2263/go-logindex/internal/memindex"
)

func TestMemIndexFlush(t *testing.T) {
	m := memindex.New(analyze.NewDefault())
	m.Add(1, "error timeout")
	m.Add(2, "error disk")
	m.Delete(2)
	if m.Posting("error").Len() != 1 {
		t.Fatalf("%v", m.Posting("error").Docs())
	}
	seg, pending := m.FlushToSegment(1)
	if seg.Meta().DocCount != 1 {
		t.Fatalf("docs %d", seg.Meta().DocCount)
	}
	if _, ok := pending[2]; !ok {
		t.Fatalf("pending %v", pending)
	}
	if m.DocCount() != 0 {
		t.Fatal("not cleared")
	}
}
