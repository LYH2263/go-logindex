package logindex_test

import (
	"testing"

	logindex "github.com/LYH2263/go-logindex"
)

func TestBug01_AddSearchAfterCloseReturnsErrClosed(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithWAL(false), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Add(1, "error timeout"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Close(); err != nil {
		t.Fatal(err)
	}
	err = idx.Add(2, "error disk")
	if err != logindex.ErrClosed {
		t.Fatalf("Add after Close: want ErrClosed, got %v", err)
	}
	_, err = idx.SearchDocIDs(logindex.Term("error"))
	if err != logindex.ErrClosed {
		t.Fatalf("Search after Close: want ErrClosed, got %v", err)
	}
}
