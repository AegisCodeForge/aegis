package receipt

import (
	"errors"
	"math/rand/v2"
)

type Receipt struct {
	Id string
	Command []string
	IssueTime int64  // timestamp
	TimeoutMinute int64
}

type GitusReceiptSystemInterface interface {
	IsReceiptSystemUsable() (bool, error)
	Install() error
	RetrieveReceipt(rid string) (*Receipt, error)
	IssueReceipt(timeoutMinute int64, command []string) (string, error)
	CancelReceipt(rid string) error
}

const passchdict = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
func NewReceiptId() string {
	res := make([]byte, 0)
	for _ = range 48 {
		res = append(res, passchdict[rand.IntN(len(passchdict))])
	}
	return string(res)
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

