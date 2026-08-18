package segment_test

import (
	"path/filepath"
	"testing"

	"github.com/LYH2263/go-logindex/internal/segment"
)

func TestPersistLoadAndTombstone(t *testing.T) {
	dir := t.TempDir()
	b := segment.NewBuilder()
	b.AddDoc(1, []string{"error", "timeout"})
	b.AddDoc(2, []string{"error", "disk"})
	b.AddDoc(3, []string{"info"})
	seg := b.Build(1)
	if err := seg.Persist(dir); err != nil {
		t.Fatal(err)
	}
	cat := &segment.Catalog{NextID: 2, Segments: []segment.ID{1}}
	if err := segment.SaveCatalog(dir, cat); err != nil {
		t.Fatal(err)
	}
	loaded, err := segment.Load(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Posting("error").Len() != 2 {
		t.Fatalf("error posting %v", loaded.Posting("error").Docs())
	}
	if !loaded.MarkDeleted(2) {
		t.Fatal("mark deleted")
	}
	if loaded.Posting("error").Len() != 1 {
		t.Fatalf("after tombstone %v", loaded.Posting("error").Docs())
	}
	if !segment.Exists(dir, 1) {
		t.Fatalf("missing %s", filepath.Join(dir, segment.DirName(1)))
	}
}
