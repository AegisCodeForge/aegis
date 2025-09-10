package db

import (
	"strings"
)

func ToSqlSearchPattern(s string) string {
	res := strings.ReplaceAll(s, "\\", "\\\\")
	res = strings.ReplaceAll(s, "%", "\\%")
	res = strings.ReplaceAll(s, "_", "\\_")
	res = "%" + res + "%"
	return res
}

