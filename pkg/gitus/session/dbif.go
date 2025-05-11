package session

import (
	"math/rand"
)

type GitusSessionStore interface {
	Install() error
	IsSessionStoreUsable() (bool, error)
	RegisterSession(name string, session string) error
	RetrieveSession(name string) (string, error)
	VerifySession(name string, target string) (bool, error)
	RevokeSession(target string) error
}

const passchdict = "abcdefghijklmnopqrstuvwxyz0123456789"
func NewSessionString() string {
	res := make([]byte, 0)
	for _ = range 48 {
		res = append(res, passchdict[rand.Intn(len(passchdict))])
	}
	return string(res)
}


