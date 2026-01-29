package confirm_code

import (
	"errors"
	"time"

	"github.com/GitusCodeForge/Gitus/pkg/gitus"
	"github.com/GitusCodeForge/Gitus/pkg/tcache"
)

// temporary cache.
// used to store kv pairs that expires after a set amount of time.

type GitusConfirmCodeManager interface {
	Register(key string, value string, d time.Duration)
	Get(key string) (string, bool)
}

var ErrNotSupported = errors.New("Not supported")

func InitializeConfirmCodeManager(cfg *gitus.GitusConfig) (GitusConfirmCodeManager, error) {
	switch cfg.ConfirmCodeManager.Type {
	case "in-memory":
		return tcache.NewTCache(time.Duration(cfg.ConfirmCodeManager.DefaultTimeoutMinute * int(time.Minute))), nil
	}
	return nil, ErrNotSupported
}

