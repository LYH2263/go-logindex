package logindex_test

import (
	"errors"
	"testing"

	logindex "github.com/LYH2263/go-logindex"
	"github.com/LYH2263/go-logindex/internal/segment"
)

func TestBug09_FlushPersistErrorDoesNotDropSegment(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithWAL(true), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Add(1, "persist me please"); err != nil {
		t.Fatal(err)
	}
	segment.FailNextPersist = errors.New("segment: injected persist/fsync failure")
	_ = idx.Flush()
	if err := idx.Close(); err != nil {
		t.Fatal(err)
	}
	idx2, err := logindex.Open(dir, logindex.WithWAL(true), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	defer idx2.Close()
	ids, err := idx2.SearchDocIDs(logindex.Term("persist"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("reopen missing last segment, got %v", ids)
	}
}
