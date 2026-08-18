package segment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// Catalog 记录根目录下的活跃段列表。
type Catalog struct {
	NextID   ID     `json:"next_id"`
	Segments []ID   `json:"segments"`
}

// LoadCatalog 读取 catalog.json。
func LoadCatalog(root string) (*Catalog, error) {
	path := filepath.Join(root, "catalog.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Catalog{NextID: 1}, nil
		}
		return nil, err
	}
	var c Catalog
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if c.NextID == 0 {
		c.NextID = 1
	}
	sort.Slice(c.Segments, func(i, j int) bool { return c.Segments[i] < c.Segments[j] })
	return &c, nil
}

// SaveCatalog 写回 catalog。
func SaveCatalog(root string, c *Catalog) error {
	if c == nil {
		c = &Catalog{NextID: 1}
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "catalog.json"), b, 0o644)
}

// AllocID 分配新段 ID。
func (c *Catalog) AllocID() ID {
	if c == nil {
		return 1
	}
	id := c.NextID
	if id == 0 {
		id = 1
	}
	c.NextID = id + 1
	return id
}

// Add 登记段。
func (c *Catalog) Add(id ID) {
	if c == nil {
		return
	}
	for _, existing := range c.Segments {
		if existing == id {
			return
		}
	}
	c.Segments = append(c.Segments, id)
	sort.Slice(c.Segments, func(i, j int) bool { return c.Segments[i] < c.Segments[j] })
}

// Remove 移除段登记。
func (c *Catalog) Remove(id ID) {
	if c == nil {
		return
	}
	out := c.Segments[:0]
	for _, s := range c.Segments {
		if s == id {
			continue
		}
		out = append(out, s)
	}
	c.Segments = out
}
