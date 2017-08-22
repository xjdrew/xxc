package xxc

import (
	"crypto/md5"
	"encoding/hex"
)

func hashPassword(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
