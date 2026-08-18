package posting

import "encoding/binary"

// Encode 将有序列表编码为字节（delta + varint）。
func Encode(l *List) []byte {
	if l == nil || l.Len() == 0 {
		return []byte{0}
	}
	buf := make([]byte, 0, 1+l.Len()*5)
	tmp := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(tmp, uint64(l.Len()))
	buf = append(buf, tmp[:n]...)
	var prev DocID
	for _, id := range l.docs {
		delta := id - prev
		n = binary.PutUvarint(tmp, uint64(delta))
		buf = append(buf, tmp[:n]...)
		prev = id
	}
	return buf
}

// Decode 解码 Encode 产物。
func Decode(buf []byte) (*List, error) {
	if len(buf) == 0 {
		return New(), nil
	}
	n, err := binary.Uvarint(buf)
	if err <= 0 {
		return nil, errCorrupt("length")
	}
	buf = buf[err:]
	ids := make([]DocID, 0, n)
	var prev DocID
	for i := uint64(0); i < n; i++ {
		delta, m := binary.Uvarint(buf)
		if m <= 0 {
			return nil, errCorrupt("delta")
		}
		buf = buf[m:]
		id := prev + DocID(delta)
		ids = append(ids, id)
		prev = id
	}
	return FromSorted(ids), nil
}

// EncodeRaw 不压缩，仅写定长 u64（调试用）。
func EncodeRaw(l *List) []byte {
	if l == nil {
		return nil
	}
	buf := make([]byte, 8*l.Len())
	for i, id := range l.docs {
		binary.LittleEndian.PutUint64(buf[i*8:], uint64(id))
	}
	return buf
}

// DecodeRaw 解码定长 u64。
func DecodeRaw(buf []byte) *List {
	if len(buf)%8 != 0 {
		return New()
	}
	n := len(buf) / 8
	ids := make([]DocID, n)
	for i := 0; i < n; i++ {
		ids[i] = binary.LittleEndian.Uint64(buf[i*8:])
	}
	return FromSorted(ids)
}
