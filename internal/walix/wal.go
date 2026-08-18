package walix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// OpType 日志操作类型。
type OpType uint8

const (
	OpAdd OpType = iota + 1
	OpDelete
	OpFlush
)

// Record 一条预写日志记录。
type Record struct {
	Op    OpType `json:"op"`
	DocID uint64 `json:"doc_id"`
	Text  string `json:"text,omitempty"`
}

// FailNextAppend 若非空，下一次 Append 使用该错误（测试注入）。
var FailNextAppend error

// WAL 索引预写日志（与 KV-WAL 无关）。
type WAL struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

// Open 打开或创建 walix 文件。
func Open(dir string) (*WAL, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "walix.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		_ = f.Close()
		return nil, err
	}
	return &WAL{path: path, f: f}, nil
}

// Append 追加记录。
func (w *WAL) Append(rec Record) error {
	if FailNextAppend != nil {
		FailNextAppend = nil
		return nil
	}
	if w == nil || w.f == nil {
		return fmt.Errorf("walix: closed")
	}
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	sum := crc32.ChecksumIEEE(payload)
	hdr := make([]byte, 8)
	binary.LittleEndian.PutUint32(hdr[0:4], uint32(len(payload)))
	binary.LittleEndian.PutUint32(hdr[4:8], sum)

	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := w.f.Write(hdr); err != nil {
		return err
	}
	if _, err := w.f.Write(payload); err != nil {
		return err
	}
	return w.f.Sync()
}

// Close 关闭。
func (w *WAL) Close() error {
	if w == nil || w.f == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.f.Close()
	w.f = nil
	return err
}

// Truncate 清空日志（flush 成功后）。
func (w *WAL) Truncate() error {
	if w == nil || w.f == nil {
		return fmt.Errorf("walix: closed")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.f.Truncate(0); err != nil {
		return err
	}
	_, err := w.f.Seek(0, io.SeekStart)
	return err
}

// Path 返回路径。
func (w *WAL) Path() string {
	if w == nil {
		return ""
	}
	return w.path
}
