package logindex_test

import (
	"testing"

	logindex "github.com/LYH2263/go-logindex"
)

func TestBug04_FlushPendingDeletesMarkOldSegments(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithFlushThreshold(10000), logindex.WithWAL(false))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err := idx.Add(1, "alpha beta gamma"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := idx.Delete(1); err != nil {
		t.Fatal(err)
	}
	if err := idx.Flush(); err != nil {
		t.Fatal(err)
	}
	ids, err := idx.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("flush pending deletes must MarkDeleted on old segments, still %v", ids)
	}
}
