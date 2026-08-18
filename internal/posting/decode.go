package posting

import "fmt"

type corruptError struct {
	what string
}

func (e corruptError) Error() string {
	return fmt.Sprintf("posting: corrupt encoding (%s)", e.what)
}

func errCorrupt(what string) error {
	return corruptError{what: what}
}

// IsCorrupt 判断是否为编码损坏错误。
func IsCorrupt(err error) bool {
	_, ok := err.(corruptError)
	return ok
}
