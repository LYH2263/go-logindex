package logindex_test

import (
	"errors"
	"testing"

	logindex "github.com/LYH2263/go-logindex"
	"github.com/LYH2263/go-logindex/internal/walix"
)

func TestBug02_AddDoesNotPublishIfWALFails(t *testing.T) {
	dir := t.TempDir()
	idx, err := logindex.Open(dir, logindex.WithWAL(true), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Add(1, "alpha keep me"); err != nil {
		t.Fatal(err)
	}
	walix.FailNextAppend = errors.New("walix: injected write failure")
	addErr := idx.Add(2, "beta should not persist")
	idsMem, err := idx.SearchDocIDs(logindex.Term("beta"))
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Close(); err != nil {
		t.Fatal(err)
	}
	idx2, err := logindex.Open(dir, logindex.WithWAL(true), logindex.WithFlushThreshold(10_000))
	if err != nil {
		t.Fatal(err)
	}
	defer idx2.Close()
	idsRec, err := idx2.SearchDocIDs(logindex.Term("beta"))
	if err != nil {
		t.Fatal(err)
	}
	if len(idsMem) != len(idsRec) {
		t.Fatalf("Add published in-memory %v but recover got %v (addErr=%v); WAL failure must not publish a doc", idsMem, idsRec, addErr)
	}
	keep, err := idx2.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if len(keep) != 1 || keep[0] != 1 {
		t.Fatalf("doc 1 must survive, got %v", keep)
	}
}
