package walix_test

import (
	"path/filepath"
	"testing"

	"github.com/LYH2263/go-logindex/internal/walix"
)

func TestWALAppendReplay(t *testing.T) {
	dir := t.TempDir()
	w, err := walix.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Append(walix.Record{Op: walix.OpAdd, DocID: 1, Text: "hello world"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Append(walix.Record{Op: walix.OpDelete, DocID: 1}); err != nil {
		t.Fatal(err)
	}
	path := w.Path()
	_ = w.Close()

	recs, err := walix.Replay(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Fatalf("recs=%d path=%s", len(recs), filepath.Base(path))
	}
	if recs[0].Op != walix.OpAdd || recs[0].Text != "hello world" {
		t.Fatalf("%+v", recs[0])
	}
}
