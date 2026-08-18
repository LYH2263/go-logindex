package walix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

// Replay 读取全部合法记录；遇损坏停止并返回已读记录 + 错误。
func Replay(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var out []Record
	hdr := make([]byte, 8)
	for {
		if _, err := io.ReadFull(f, hdr); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return out, err
		}
		n := binary.LittleEndian.Uint32(hdr[0:4])
		sum := binary.LittleEndian.Uint32(hdr[4:8])
		if n > 16<<20 {
			return out, fmt.Errorf("walix: record too large")
		}
		payload := make([]byte, n)
		if _, err := io.ReadFull(f, payload); err != nil {
			return out, fmt.Errorf("walix: truncated payload: %w", err)
		}
		if crc32.ChecksumIEEE(payload) != sum {
			return out, fmt.Errorf("walix: checksum mismatch")
		}
		var rec Record
		if err := json.Unmarshal(payload, &rec); err != nil {
			return out, err
		}
		out = append(out, rec)
	}
	return out, nil
}

// RecoverDir 打开目录下 walix 并回放。
func RecoverDir(dir string) ([]Record, error) {
	w, err := Open(dir)
	if err != nil {
		return nil, err
	}
	path := w.Path()
	_ = w.Close()
	return Replay(path)
}
