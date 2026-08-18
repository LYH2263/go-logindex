package walix

import (
	"hash/crc32"
)

// Checksum 计算 payload CRC。
func Checksum(payload []byte) uint32 {
	return crc32.ChecksumIEEE(payload)
}

// Verify 校验。
func Verify(payload []byte, want uint32) bool {
	return Checksum(payload) == want
}

// EncodeRecordFrame 仅用于测试辅助：编码一帧。
func EncodeRecordFrame(payload []byte) []byte {
	sum := Checksum(payload)
	out := make([]byte, 8+len(payload))
	out[0] = byte(len(payload))
	out[1] = byte(len(payload) >> 8)
	out[2] = byte(len(payload) >> 16)
	out[3] = byte(len(payload) >> 24)
	out[4] = byte(sum)
	out[5] = byte(sum >> 8)
	out[6] = byte(sum >> 16)
	out[7] = byte(sum >> 24)
	copy(out[8:], payload)
	return out
}
