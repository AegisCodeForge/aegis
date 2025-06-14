package redis_like

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/session"
	"github.com/redis/go-redis/v9"
)

type AegisRedisLikeSessionStore struct {
	config *aegis.AegisConfig
	connection *redis.Client
}

func NewAegisRedisLikeSessionStore(cfg *aegis.AegisConfig) (*AegisRedisLikeSessionStore, error) {
	c := redis.NewClient(&redis.Options{
		Addr: cfg.Session.Host,
		Username: cfg.Session.UserName,
		Password: cfg.Session.Password,
		DB: cfg.Session.DatabaseNumber,
	})
	return &AegisRedisLikeSessionStore{
		config: cfg,
		connection: c,
	}, nil
}

func (ssif *AegisRedisLikeSessionStore) Install() error {
	return nil
}

func (ssif *AegisRedisLikeSessionStore) IsSessionStoreUsable() (bool, error) {
	return true, nil
}

func (ssif *AegisRedisLikeSessionStore) RegisterSession(name string, session string) error {
	key := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, name)
	ctx := context.TODO()
	// timeoutStr := fmt.Sprintf("%d", 24 * 60 * 60)
	timestampStr := fmt.Sprintf("%d", time.Now().UnixMilli())
	r1 := ssif.connection.HSet(ctx, key, session, timestampStr)
	if r1.Err() != nil { return r1.Err() }
	return nil
}

func (ssif *AegisRedisLikeSessionStore) RetrieveSession(name string) ([]*session.AegisSession, error) {
	key := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, name)
	// TODO: session being a set is reserved for the support of multiple
	// sessions. we currently don't have related logic yet; this is here
	// just so we don't have to change how things are stored in redis in the
	// future; it's almost always more hassle if we were forced to change
	// things beyond the code.
	// NOTE (for people who're unfamiliar with redis): *SCAN commands (as
	// told by Redis's documentations) does not guarantee the number of
	// values and you "should not consider the iteration complete as long
	// as the returned cursor is not zero".
	res := make([]*session.AegisSession, 0)
	lastCursor := uint64(0)
	for {
		cmd := ssif.connection.HScan(context.TODO(), key, uint64(lastCursor), "*", 0)
		keys, cursor, err := cmd.Result()
		if err != nil { return nil, err }
		i := 0
		lenk := len(keys)
		for i < lenk {
			timestamp, _ := strconv.ParseInt(keys[2*i+1], 10, 64)
			res = append(res, &session.AegisSession{
				Username: name,
				Id: keys[2*i],
				Timestamp: timestamp,
			})
			i += 2
		}
		if cursor == 0 { break}
		lastCursor = cursor
	}
	return res, nil
}

func (ssif *AegisRedisLikeSessionStore) RetrieveSessionByKey(username string, sessionid string) (*session.AegisSession, error) {
	key := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, username)
	cmd := ssif.connection.HGet(context.TODO(), key, sessionid)
	if cmd.Err() == redis.Nil { return nil, nil }
	if cmd.Err() != nil { return nil, cmd.Err() }
	r, err := cmd.Result()
	if err != nil { return nil, err }
	if len(r) <= 0 { return nil, nil }
	timestamp, _ := strconv.ParseInt(r, 10, 64)
	return &session.AegisSession{
		Username: username,
		Id: sessionid,
		Timestamp: timestamp,
	}, nil
}

func (ssif *AegisRedisLikeSessionStore) VerifySession(name string, target string) (bool, error) {
	key := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, name)
	cmd := ssif.connection.HGet(context.TODO(), key, target)
	if cmd.Err() == redis.Nil { return false, nil }
	if cmd.Err() != nil { return false, cmd.Err() }
	r, err := cmd.Result()
	if err != nil { return false, err }
	if len(r) <= 0 { return false, nil }
	return true, nil
}

func (ssif *AegisRedisLikeSessionStore) RevokeSession(username string, target string) error {
	key := fmt.Sprintf("%s:%s:session", ssif.config.Session.TablePrefix, username)
	cmd := ssif.connection.HDel(context.TODO(), key, target)
	if cmd.Err() != nil { return cmd.Err() }
	return nil
}

