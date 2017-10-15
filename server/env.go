package server

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/Sovianum/acquaintance-server/config"
	"github.com/Sovianum/acquaintance-server/dao"
	"github.com/Sovianum/acquaintance-server/mylog"
	"github.com/patrickmn/go-cache"
	"time"
)

type tokenKeyGetterType func() string

func NewEnv(db *sql.DB, conf config.Conf, logger *mylog.Logger) *Env {
	var env = &Env{
		userDAO:        dao.NewDBUserDAO(db),
		positionDAO:    dao.NewDBPositionDAO(db),
		meetRequestDAO: dao.NewMeetDAO(db),
		conf:           conf,
		meetRequestCache: cache.New(
			time.Second*time.Duration(conf.Logic.RequestExpiration),
			time.Second*time.Duration(conf.Logic.CleanupInterval),
		),
		hashFunc: func(password []byte) ([]byte, error) {
			var h = sha256.New()
			h.Write(password)
			return h.Sum(nil), nil
		},
		hashValidator: func(password []byte, hash []byte) error {
			var h = sha256.New()
			h.Write(password)
			var passHash = h.Sum(nil)

			if string(passHash) != string(hash) {
				return fmt.Errorf("hashes %s, %s do not match", string(passHash), string(hash))
			}
			return nil
		},
		logger: logger,
	}

	env.RunDaemons()
	return env
}

type Env struct {
	userDAO          dao.UserDAO
	positionDAO      dao.PositionDAO
	meetRequestDAO   dao.MeetRequestDAO
	conf             config.Conf
	hashFunc         func(password []byte) ([]byte, error)
	hashValidator    func(password []byte, hash []byte) error
	meetRequestCache *cache.Cache
	logger           *mylog.Logger
}
