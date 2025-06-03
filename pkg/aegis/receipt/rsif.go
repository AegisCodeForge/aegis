package receipt

import (
	"errors"
	"math/rand/v2"
	"strings"
	"time"
)

type Receipt struct {
	Id string
	Command []string
	IssueTime int64  // timestamp
	TimeoutMinute int64
}

type AegisReceiptSystemInterface interface {
	IsReceiptSystemUsable() (bool, error)
	Install() error
	RetrieveReceipt(rid string) (*Receipt, error)
	IssueReceipt(timeoutMinute int64, command []string) (string, error)
	CancelReceipt(rid string) error
	GetAllReceipt(pageNum int, pageSize int) ([]*Receipt, error)
	SearchReceipt(q string, pageNum int, pageSize int) ([]*Receipt, error)
	EditReceipt(id string, robj *Receipt) error
}

const passchdict = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
func NewReceiptId() string {
	res := make([]byte, 0)
	for _ = range 48 {
		res = append(res, passchdict[rand.IntN(len(passchdict))])
	}
	return string(res)
}

func SerializeReceiptCommand(s []string) string {
	es := make([]string, len(s))
	for i, k := range s {
		es[i] = cmdArgEscape(k)
	}
	return strings.Join(es, ",")
}

func cmdArgEscape(s string) string {
	// escape for receipt command.  receipt command is separated by
	// commas, thus special care needs to be taken if the args
	// themselves contain commas.
	quoted := false
	if strings.ContainsRune(s, ',') { quoted = true }
	r := new(strings.Builder)
	if quoted { r.WriteRune('"') }
	for _, ch := range s {
		if quoted && ch == '"' {
			r.WriteString("\\\"")
		} else {
			r.WriteRune(ch)
		}
	}
	if quoted { r.WriteRune('"') }
	return r.String()
}
func NewReceiptCommand(s... string) string {
	res := make([]string, 0)
	for _, item := range s {
		res = append(res, cmdArgEscape(item))
	}
	return strings.Join(res, ",")
}
func ParseReceiptCommand(s string) []string {
	res := make([]string, 0)
	b := new(strings.Builder)
	inQuote := false
	escaped := false
	for _, item := range s {
		if escaped {
			b.WriteRune(item)
			escaped = false
		} else if inQuote {
			if item == '\\' {
				escaped = true
			} else {
				b.WriteRune(item)
			}
		} else {
			if item == ',' {
				res = append(res, b.String())
				b = new(strings.Builder)
			} else {
				b.WriteRune(item)
			}
		}
	}
	if b.Len() > 0 { res = append(res, b.String()) }
	return res
}

func (r *Receipt) Expired() bool {
	return (time.Now().Unix() - r.IssueTime) >= r.TimeoutMinute * 60
}

const (
	// confirm-registration,{username},{email},{passwordHash}
	CONFIRM_REGISTRATION = "confirm-registration"
	// confirm-email-change,{username},{email}
	CONFIRM_EMAIL_CHANGE = "confirm-email-change"
	// reset-password,{username}
	RESET_PASSWORD = "reset-password"
)

var ErrUnsupportedSystemType = errors.New("Unsupported receipt system type")

