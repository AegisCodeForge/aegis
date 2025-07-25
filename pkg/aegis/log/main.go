package log

import (
	"fmt"
	"time"
)

func INFO(a ...any) {
	fmt.Printf("%s [INFO] ", time.Now().Format(time.RFC3339))
	fmt.Print(a...)
}

func WARN(a ...any) {
	fmt.Printf("%s [WARN] ", time.Now().Format(time.RFC3339))
	fmt.Print(a...)
}

func ERR(a ...any) {
	r := fmt.Sprintf("$s [ERR] ", time.Now().Format(time.RFC3339))
	r2 := fmt.Sprint(a...)
	fmt.Print(r + r2)
}
