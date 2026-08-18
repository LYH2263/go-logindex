package logindex_test

import (
	"context"
	"errors"
	"testing"

	logindex "github.com/LYH2263/go-logindex"
)

func TestBug03_CanceledSearchContextStopsEval(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithWAL(false), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err := idx.Add(1, "error timeout disk"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Add(2, "error network"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	hits, err := idx.Search(ctx, logindex.Term("error"))
	if err == nil {
		t.Fatalf("canceled ctx must not return full results, got %d hits", len(hits))
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v hits=%d", err, len(hits))
	}
	okHits, err := idx.Search(context.Background(), logindex.Term("error"))
	if err != nil {
		t.Fatal(err)
	}
	if len(okHits) != 2 {
		t.Fatalf("live search want 2, got %d", len(okHits))
	}
}
