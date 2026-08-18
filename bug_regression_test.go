package logindex_test

import (
	"reflect"
	"testing"

	logindex "github.com/LYH2263/go-logindex"
)

func TestBug08_WALRecoverMarksOverlayDeleted(t *testing.T) {
	dir := t.TempDir()
	{
		idx, err := logindex.Open(dir, logindex.WithFlushThreshold(10000), logindex.WithWAL(true))
		if err != nil {
			t.Fatal(err)
		}
		if err := idx.Add(1, "old alpha unique"); err != nil {
			t.Fatal(err)
		}
		if err := idx.Flush(); err != nil {
			t.Fatal(err)
		}
		if err := idx.Add(1, "new beta unique"); err != nil {
			t.Fatal(err)
		}
		if err := idx.Close(); err != nil {
			t.Fatal(err)
		}
	}
	idx, err := logindex.Open(dir, logindex.WithWAL(true))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	alpha, err := idx.SearchDocIDs(logindex.Term("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if len(alpha) != 0 {
		t.Fatalf("WAL recover overlay must MarkDeleted old segment, alpha still %v", alpha)
	}
	beta, err := idx.SearchDocIDs(logindex.Term("beta"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beta, []uint64{1}) {
		t.Fatalf("overlay text missing, beta %v", beta)
	}
}
