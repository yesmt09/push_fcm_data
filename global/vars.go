package global

import (
	"github.com/go-redis/redis/v8"
	"gitlab.babeltime.com/packagist/blogger"
	"os"
	"upload_fcm_data/types"
)

var (
	ConfigFilePath string
	Config         types.ConfigType
	Logger         blogger.BLogger
	Rdb            redis.UniversalClient
	rdbLinkKey     = "-"
	DEBUG          bool
	Signal         = make(chan os.Signal, 1)
)

func GetRdbKey(gn string, bizid string) string {
	return gn + rdbLinkKey + "addiction-fail-subscribe-machao2"
}
