package project

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
	"upload_fcm_data/fcm"
	"upload_fcm_data/global"
)

func (p pushRequest) getFcmBehaviorList() []fcm.Behavior {
	cacheKey := global.GetRdbKey(p.project.Game, p.project.Bizid)
	var behaviorList = make([]fcm.Behavior, 0)
	for i := 1; i <= global.Config.RequestMaxNum; i++ {
		var data string
		var err error
		data, err = global.Rdb.RPop(ctx, cacheKey).Result()
		if redis.Nil == err {
			p.logger.Info("data is empty")
			break
		} else if nil != err {
			p.logger.Warning("err")
			panic(err)
		}
		// 整理数据
		var rdbData map[string]interface{}
		err = json.Unmarshal([]byte(data), &rdbData)

		if err != nil {
			p.logger.Warning("json err")
			continue
		}
		var ot int64
		if 180 < (time.Now().Unix() - int64(rdbData["ot"].(float64))) {
			ot = time.Now().Unix()
		} else {
			ot = int64(rdbData["ot"].(float64))
		}
		behaviorList = append(behaviorList, fcm.Behavior{
			No: i,
			Si: rdbData["si"].(string),
			Bt: int(rdbData["bt"].(float64)),
			Ot: ot,
			Ct: int(rdbData["ct"].(float64)),
			Di: rdbData["di"].(string),
			Pi: rdbData["pi"].(string),
		})
	}
	return behaviorList
}
