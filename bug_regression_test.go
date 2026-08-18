package logindex_test

import (
	"testing"

	logindex "github.com/LYH2263/go-logindex"
)

func TestBug06_SearchDocsSliceDoesNotAliasIndex(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithWAL(false), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err := idx.Add(1, "alpha beta gamma"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Add(3, "alpha other"); err != nil {
		t.Fatal(err)
	}
	ids, err := idx.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) < 1 {
		t.Fatal("expected hits")
	}
	ids[0] = 0xdeadbeef
	ids = append(ids, 42)
	ids2, err := idx.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range ids2 {
		if id == 0xdeadbeef || id == 42 {
			t.Fatalf("caller mutate/append of search slice corrupted index posting: %v", ids2)
		}
	}
	if len(ids2) != 2 || ids2[0] != 1 || ids2[1] != 3 {
		t.Fatalf("want [1 3], got %v", ids2)
	}
}
