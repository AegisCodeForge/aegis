package in_memory

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GitusCodeForge/Gitus/pkg/gitus"
	"github.com/GitusCodeForge/Gitus/pkg/gitus/session"
	"github.com/GitusCodeForge/Gitus/pkg/tcache"
)

type GitusInMemorySessionStore struct {
	config *gitus.GitusConfig
	cache *tcache.TCache
}

func NewGitusInMemorySessionStore(cfg *gitus.GitusConfig) (*GitusInMemorySessionStore, error) {
	c := tcache.NewTCache(24 * time.Hour)
	return &GitusInMemorySessionStore{
		config: cfg,
		cache: c,
	}, nil
}

func (ssif *GitusInMemorySessionStore) Install() error {
	return nil
}

func (ssif *GitusInMemorySessionStore) IsSessionStoreUsable() (bool, error) {
	return true, nil
}

func (ssif *GitusInMemorySessionStore) Dispose() error {
	return nil
}

func insertSet(set string, s string) string {
	if inSet(set, s) {
		return set
	} else {
		return fmt.Sprintf("%s{%s}", set, s)
	}
}
func inSet(set string, s string) bool {
	return strings.Contains(set, fmt.Sprintf("{%s}", s))
}
func removeFromSet(set string, s string) string {
	ss := fmt.Sprintf("{%s}", s)
	i := strings.Index(set, ss)
	if i <= -1 { return set }
	return set[0:i] + set[i+len(ss):]
}

func (ssif *GitusInMemorySessionStore) RegisterSession(name string, session string) error {
	key := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, name)
	i, _ := ssif.cache.Get(key)
	ssif.cache.Register(key, insertSet(i, session), 24*time.Hour)
	key2 := fmt.Sprintf("%s:%s:session:%s", ssif.config.Session.TablePrefix, name, session)
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
	ssif.cache.Register(key2, timestampStr, 24*time.Hour)
	return nil
}

func (ssif *GitusInMemorySessionStore) RetrieveSession(name string) ([]*session.GitusSession, error) {
	groupKey := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, name)
	i, _ := ssif.cache.Get(groupKey)
	res := make([]*session.GitusSession, 0)
	newset := ""
	for k := range strings.SplitSeq(string(i[1:len(i)]), "}{") {
		kk := fmt.Sprintf("%s:%s", groupKey, k)
		v, _ := ssif.cache.Get(kk)
		if v != "" {
			timestamp, _ := strconv.ParseInt(v, 10, 64)
			res = append(res, &session.GitusSession{
				Username: name,
				Id: k,
				Timestamp: timestamp,
			})
			newset = fmt.Sprintln("%s{%s}", newset, k)
		}
	}
	if newset == "" {
		ssif.cache.Delete(groupKey)
	} else {
		ssif.cache.Register(groupKey, newset, 24*time.Hour)
	}
	return res, nil
}

func (ssif *GitusInMemorySessionStore) RetrieveSessionByKey(username string, sessionid string) (*session.GitusSession, error) {
	key := fmt.Sprintf("%s:%s:session:%s", ssif.config.Session.TablePrefix, username, sessionid)
	i, _ := ssif.cache.Get(key)
	timestamp, _ := strconv.ParseInt(i, 10, 64)
	return &session.GitusSession{
		Username: username,
		Id: sessionid,
		Timestamp: timestamp,
	}, nil
}

func (ssif *GitusInMemorySessionStore) VerifySession(name string, target string) (bool, error) {
	key := fmt.Sprintf("%s:%s:session:%s", ssif.config.Session.TablePrefix, name, target)
	i, _ := ssif.cache.Get(key)
	if i == "" { return false, nil }
	timestamp, _ := strconv.ParseInt(i, 10, 64)
	if time.Now().Unix() > timestamp {
		ssif.cache.Delete(key)
		return false, nil
	}
	return true, nil
}

func (ssif *GitusInMemorySessionStore) RevokeSession(username string, target string) error {
	// NOTE: we don't have transaction semantics here, which could be
	// a problem down the line.
	// TODO: attempt to fix this.
	key := fmt.Sprintf("%s:%s:session:%s", ssif.config.Session.TablePrefix, username, target)
	ssif.cache.Delete(key)
	sessionSetKey := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, username)
	i, ok := ssif.cache.Get(sessionSetKey)
	if !ok { return nil }
	i = removeFromSet(i, target)
	ssif.cache.Register(sessionSetKey, i, 24*time.Hour)
	return nil
}

