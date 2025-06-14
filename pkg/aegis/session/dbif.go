package session

import (
	"math/rand"
)

type AegisSession struct {
	Username string
	Id string
	Timestamp int64
}

type AegisSessionStore interface {
	Install() error
	IsSessionStoreUsable() (bool, error)
	RegisterSession(username string, session string) error
	RetrieveSession(username string) ([]*AegisSession, error)
	RetrieveSessionByKey(username string, session string) (*AegisSession, error)
	VerifySession(username string, target string) (bool, error)
	RevokeSession(username string, target string) error
}

const passchdict = "abcdefghijklmnopqrstuvwxyz0123456789"
func NewSessionString() string {
	res := make([]byte, 0)
	for range 48 {
		res = append(res, passchdict[rand.Intn(len(passchdict))])
	}
	return string(res)
}


