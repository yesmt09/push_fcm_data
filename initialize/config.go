package initialize

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gitlab.babeltime.com/packagist/blogger"
	"os"
	"upload_fcm_data/global"
)

func InitConfig() {
	if _, err := os.Stat(global.ConfigFilePath); os.IsNotExist(err) {
		panic(fmt.Sprintf("types file not Exist: %v", global.ConfigFilePath))
	}
	configFile, _ := os.Open(global.ConfigFilePath)
	defer configFile.Close()
	json.NewDecoder(configFile).Decode(&global.Config)
	if global.Config.Redis.IsCluster {
		global.Rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    global.Config.Redis.HostList,
			Password: global.Config.Redis.Password,
		})
	} else {
		global.Rdb = redis.NewClient(&redis.Options{
			Addr:     global.Config.Redis.HostList[0],
			Password: global.Config.Redis.Password,
		})
	}
	global.Logger = blogger.NewBlogger(global.Config.Logger.Filepath, global.Config.Logger.Level)
	ctx := context.Background()
	global.Rdb.Ping(ctx)
}
