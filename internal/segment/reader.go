package segment

import (
	"fmt"
	"os"
	"path/filepath"
)

// Reader 批量加载根目录下全部活跃段。
type Reader struct {
	Root string
}

// LoadAll 按 catalog 加载。
func (r Reader) LoadAll() ([]*Segment, *Catalog, error) {
	cat, err := LoadCatalog(r.Root)
	if err != nil {
		return nil, nil, err
	}
	segs := make([]*Segment, 0, len(cat.Segments))
	for _, id := range cat.Segments {
		seg, err := Load(r.Root, id)
		if err != nil {
			return nil, nil, fmt.Errorf("load segment %d: %w", id, err)
		}
		segs = append(segs, seg)
	}
	return segs, cat, nil
}

// Exists 检查段目录是否存在。
func Exists(root string, id ID) bool {
	_, err := os.Stat(filepath.Join(root, DirName(id)))
	return err == nil
}

// RemoveDir 删除段目录（磁盘）。
func RemoveDir(root string, id ID) error {
	return os.RemoveAll(filepath.Join(root, DirName(id)))
}
