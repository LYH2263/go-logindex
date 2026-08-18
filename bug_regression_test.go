package logindex_test

import (
	"testing"

	logindex "github.com/LYH2263/go-logindex"
)

func TestBug07_SearchIntersectsLiveUniverse(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithFlushThreshold(10000), logindex.WithWAL(false))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err := idx.Add(1, "error timeout connecting"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Delete(1); err != nil {
		t.Fatal(err)
	}
	ids, err := idx.SearchDocIDs(logindex.Term("error"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("Search must Intersect live Universe; deleted still %v", ids)
	}
}
