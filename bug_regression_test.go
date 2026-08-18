package logindex_test

import (
	"testing"

	logindex "github.com/LYH2263/go-logindex"
)

func TestBug10_DeleteWALSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithWAL(true), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Add(1, "alpha beta gamma"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := idx.Delete(1); err != nil {
		t.Fatal(err)
	}
	ids, err := idx.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("in-memory delete still hit %v", ids)
	}
	if err := idx.Close(); err != nil {
		t.Fatal(err)
	}
	idx2, err := logindex.Open(dir, logindex.WithWAL(true), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	defer idx2.Close()
	ids, err = idx2.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("recover after reopen brought deleted doc back: %v", ids)
	}
}
