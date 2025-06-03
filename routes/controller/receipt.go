package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bctnry/aegis/pkg/aegis/receipt"
	"github.com/bctnry/aegis/routes"
)

func bindReceiptController(ctx *routes.RouterContext) {
	//     /receipt?id={receipt id}
	http.HandleFunc("GET /receipt", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rid := r.URL.Query().Get("id")
		re, err := ctx.ReceiptSystem.RetrieveReceipt(rid)
		fmt.Println(re)
		if err != nil { routes.FoundAt(w, "/"); return }
		if (time.Now().Unix() - re.IssueTime) >= re.TimeoutMinute*60 {
			ctx.ReceiptSystem.CancelReceipt(rid)
			routes.FoundAt(w, "/")
			return
		}
		if len(re.Command) <= 0 {
			// invalid receipt command...
			routes.FoundAt(w, "/")
			return
		}
		fmt.Println(re)
		switch re.Command[0] {
		case receipt.CONFIRM_REGISTRATION:
			routes.FoundAt(w, fmt.Sprintf("/confirm-registration?id=%s", rid))
			return
		case receipt.RESET_PASSWORD:
			routes.FoundAt(w, fmt.Sprintf("/reset-password/update-password?id=%s", rid))
			return
		}
		ctx.ReportNormalError("Invalid receipt", w, r)
	}))
}

