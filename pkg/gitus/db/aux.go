package db

import (
	"path/filepath"
	"strings"
)

func ToSqlSearchPattern(s string) string {
	res := strings.ReplaceAll(s, "\\", "\\\\")
	res = strings.ReplaceAll(s, "%", "\\%")
	res = strings.ReplaceAll(s, "_", "\\_")
	res = "%" + res + "%"
	return res
}

func IsSubDir(r string, d string) bool {
	rr, err := filepath.Rel(r, d)
	if err != nil { return false }
	if strings.HasPrefix(rr, "..") { return false }
	return true
}

