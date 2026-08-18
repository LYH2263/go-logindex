package logindex_test

import (
	"path/filepath"
	"reflect"
	"testing"

	logindex "github.com/LYH2263/go-logindex"
	"github.com/LYH2263/go-logindex/internal/merge"
)

func TestSearchMatchesBruteForceSmallCorpus(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithFlushThreshold(1000), logindex.WithWAL(false))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	corpus := map[uint64]string{
		1: "error timeout connecting to database server",
		2: "info user login success from gateway",
		3: "error disk full on volume data",
		4: "warn timeout waiting for lock",
		5: "debug cache hit ratio high",
		6: "error timeout disk io wait",
		7: "info scheduled backup completed",
		8: "fatal database connection refused",
	}
	for id, text := range corpus {
		if err := idx.Add(id, text); err != nil {
			t.Fatal(err)
		}
	}
	if err := idx.Flush(); err != nil {
		t.Fatal(err)
	}

	queries := []logindex.Query{
		logindex.Term("error"),
		logindex.Term("timeout"),
		logindex.And(logindex.Term("error"), logindex.Term("timeout")),
		logindex.Or(logindex.Term("database"), logindex.Term("cache")),
		logindex.And(logindex.Term("error"), logindex.Not(logindex.Term("disk"))),
		logindex.Or(logindex.Term("info"), logindex.And(logindex.Term("error"), logindex.Term("timeout"))),
		logindex.Not(logindex.Term("error")),
	}

	for _, q := range queries {
		got, err := idx.SearchDocIDs(q)
		if err != nil {
			t.Fatalf("search %s: %v", q, err)
		}
		want, err := idx.BruteForceSearch(q)
		if err != nil {
			t.Fatalf("brute %s: %v", q, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("query %s\n got %v\nwant %v", q, got, want)
		}
	}
}

func TestFlushVisibilityAndDelete(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithFlushThreshold(10_000), logindex.WithWAL(true))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	if err := idx.Add(10, "alpha beta gamma"); err != nil {
		t.Fatal(err)
	}
	ids, err := idx.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []uint64{10}) {
		t.Fatalf("mem visibility: %v", ids)
	}
	if err := idx.Flush(); err != nil {
		t.Fatal(err)
	}
	if idx.MemDocCount() != 0 {
		t.Fatalf("mem not empty after flush")
	}
	ids, err = idx.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []uint64{10}) {
		t.Fatalf("post flush: %v", ids)
	}
	if err := idx.Delete(10); err != nil {
		t.Fatal(err)
	}
	ids, err = idx.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("deleted still hit: %v", ids)
	}
}

func TestMergePreservesPostings(t *testing.T) {
	dir := t.TempDir()
	policy := merge.Policy{MaxSegments: 2, MaxDocsPerSeg: 10_000, MinMergeSegs: 2}
	idx, err := logindex.Open(dir,
		logindex.WithFlushThreshold(2),
		logindex.WithWAL(false),
		logindex.WithMergePolicy(policy),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	docs := []struct {
		id   uint64
		text string
	}{
		{1, "red apple fruit"},
		{2, "green apple tree"},
		{3, "red berry fruit"},
		{4, "blue berry juice"},
		{5, "green tea leaf"},
		{6, "red tea pot"},
	}
	for _, d := range docs {
		if err := idx.Add(d.id, d.text); err != nil {
			t.Fatal(err)
		}
	}
	if err := idx.Flush(); err != nil {
		t.Fatal(err)
	}
	_ = idx.Compact()

	q := logindex.And(logindex.Term("red"), logindex.Term("fruit"))
	got, err := idx.SearchDocIDs(q)
	if err != nil {
		t.Fatal(err)
	}
	want, err := idx.BruteForceSearch(q)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("after merge got %v want %v segs=%d", got, want, idx.SegmentCount())
	}
}

func TestReopenLoadsSegments(t *testing.T) {
	dir := t.TempDir()
	{
		idx, err := logindex.Open(dir, logindex.WithFlushThreshold(100), logindex.WithWAL(false))
		if err != nil {
			t.Fatal(err)
		}
		_ = idx.Add(1, "persist me please")
		_ = idx.Flush()
		_ = idx.Close()
	}
	idx, err := logindex.Open(dir, logindex.WithWAL(false))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if idx.SegmentCount() < 1 {
		t.Fatalf("expected segments on disk in %s", filepath.Base(dir))
	}
}
