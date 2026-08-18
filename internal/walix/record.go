package walix

import "fmt"

// ValidateRecord 检查记录字段合法性。
func ValidateRecord(r Record) error {
	switch r.Op {
	case OpAdd:
		if r.Text == "" && r.DocID == 0 {
			// 允许空文本，但至少要有 doc。
		}
		return nil
	case OpDelete:
		return nil
	case OpFlush:
		return nil
	default:
		return fmt.Errorf("walix: unknown op %d", r.Op)
	}
}

// OpName 人类可读。
func OpName(op OpType) string {
	switch op {
	case OpAdd:
		return "add"
	case OpDelete:
		return "delete"
	case OpFlush:
		return "flush"
	default:
		return fmt.Sprintf("op(%d)", op)
	}
}
