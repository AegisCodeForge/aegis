package memcached

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/session"
	"github.com/bradfitz/gomemcache/memcache"
)

type AegisMemcachedSessionStore struct {
	config *aegis.AegisConfig
	connection *memcache.Client
}

func NewAegisMemcachedSessionStore(cfg *aegis.AegisConfig) (*AegisMemcachedSessionStore, error) {
	c := memcache.New(cfg.Session.Host)
	return &AegisMemcachedSessionStore{
		config: cfg,
		connection: c,
	}, nil
}

func (ssif *AegisMemcachedSessionStore) Install() error {
	return nil
}

func (ssif *AegisMemcachedSessionStore) IsSessionStoreUsable() (bool, error) {
	return true, nil
}

func insertSet(set []byte, s string) []byte {
	if inSet(set, s) {
		return set
	} else {
		if len(set) <= 0 { return []byte(s) }
		return fmt.Appendf(set, ",%s", s)
	}
}
func inSet(set []byte, s string) bool {
	if len(set) <= 0 { return false }
	for item := range strings.SplitSeq(string(set), ",") {
		if item == s { return true }
	}
	return false
}
func removeFromSet(set []byte, s string) []byte {
	ss := strings.Split(string(set), ",")
	targetI := -1
	for i, k := range ss {
		if k == s { targetI = i; break }
	}
	if targetI == -1 { return set }
	ress := slices.Delete(ss, targetI, targetI+1)
	return []byte(strings.Join(ress, ","))
}

func (ssif *AegisMemcachedSessionStore) RegisterSession(name string, session string) error {
	sessionSetKey := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, name)
	i, err := ssif.connection.Get(sessionSetKey)
	if err != nil {
		// cache miss is memcached's way of saying the key not found...
		if err != memcache.ErrCacheMiss { return err }
		i = &memcache.Item{
			Key: sessionSetKey,
			Value: []byte(session),
			Flags: 0,
			Expiration: 0,
		}
	} else {
		i.Value = insertSet(i.Value, session)
	}
	err = ssif.connection.Set(i)
	if err != nil { return err }
	key := fmt.Sprintf("%s:%s:session:%s", ssif.config.Session.TablePrefix, name, session)
	timestampStr := fmt.Sprintf("%d", time.Now().UnixMilli())
	// this is possibly the only case where a sessionid register twice
	// makes sense:
	// + we plan to support multiple session, which we may make into
	//   not having an expiration datetime a la github;
	// + memcached does not support sets like redis, so the way is
	//   to store all of them as a long string, which would subject us
	//   to the size limit of values, which is 1MB, which considering
	//   the length of each session key and how many sessions there
	//   *typically* will be, should be plenty enough.
	// + we still want easy check for each session key instead of
	//   deserializing the long string every time.
	err = ssif.connection.Set(&memcache.Item{
		Key: key,
		Value: []byte(timestampStr),
		Flags: 0,
		Expiration: 0,
	})
	if err != nil { return err }
	return nil
}

func (ssif *AegisMemcachedSessionStore) RetrieveSession(name string) ([]*session.AegisSession, error) {
	key := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, name)
	i, err := ssif.connection.Get(key)
	res := make([]*session.AegisSession, 0)
	if err == memcache.ErrCacheMiss { return res, nil }
	if err != nil { return nil, err }
	for k := range strings.SplitSeq(string(i.Value), ",") {
		kk := fmt.Sprintf("%s:%s", key, k)
		v, err := ssif.connection.Get(kk)
		var val string
		if err != nil {
			val = "0"
		} else {
			val = string(v.Value)
		}
		timestamp, _ := strconv.ParseInt(val, 10, 64)
		res = append(res, &session.AegisSession{
			Username: name,
			Id: k,
			Timestamp: timestamp,
		})
	}
	return res, nil
}

func (ssif *AegisMemcachedSessionStore) RetrieveSessionByKey(username string, sessionid string) (*session.AegisSession, error) {
	key := fmt.Sprintf("%s:%s:session:%s", ssif.config.Session.TablePrefix, username, sessionid)
	i, err := ssif.connection.Get(key)
	if err != nil { return nil, err }
	timestamp, _ := strconv.ParseInt(string(i.Value), 10, 64)
	return &session.AegisSession{
		Username: username,
		Id: sessionid,
		Timestamp: timestamp,
	}, nil
}

func (ssif *AegisMemcachedSessionStore) VerifySession(name string, target string) (bool, error) {
	key := fmt.Sprintf("%s:%s:session:%s", ssif.config.Session.TablePrefix, name, target)
	i, err := ssif.connection.Get(key)
	if err == memcache.ErrCacheMiss { return false, nil }
	if err != nil { return false, err }
	if len(i.Value) <= 0 { return false, nil }
	return true, nil
}

func (ssif *AegisMemcachedSessionStore) RevokeSession(username string, target string) error {
	// NOTE: we don't have transaction semantics here, which could be
	// a problem down the line.
	// TODO: attempt to fix this.
	key := fmt.Sprintf("%s:%s:session:%s", ssif.config.Session.TablePrefix, username, target)
	err := ssif.connection.Delete(key)
	if err != nil && err != memcache.ErrCacheMiss { return err }
	sessionSetKey := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, username)
	i, err := ssif.connection.Get(sessionSetKey)
	if err == memcache.ErrCacheMiss { return nil }
	if err != nil { return err }
	i.Value = removeFromSet(i.Value, target)
	err = ssif.connection.Set(i)
	if err != nil { return err }
	return nil
}

