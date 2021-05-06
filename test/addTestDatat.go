package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"math/rand"
	"sync"
	"time"
	"upload_fcm_data/global"
	"upload_fcm_data/initialize"
)

var gn string
var bizid string
var randNum int

func main() {
	flag.StringVar(&gn, "gn", "", "")
	flag.StringVar(&bizid, "bizid", "", "")
	flag.IntVar(&randNum, "randnum", 1000, "")
	flag.Parse()
	global.ConfigFilePath = "../config.debug.json"
	initialize.InitConfig()
	var wg sync.WaitGroup
	fmt.Println("start add data")
	defer global.Rdb.Close()
	p, _ := ants.NewPoolWithFunc(1000, func(i interface{}) {
		pushTestData()
		wg.Done()
	})
	defer p.Release()
	// Submit tasks one by one.
	for i := 0; i < randNum; i++ {
		wg.Add(1)
		_ = p.Invoke(int32(i))
	}
	wg.Wait()
	fmt.Println("done")
}

// md5
func getmd5(s string) string {
	m := md5.New()
	m.Write([]byte (s))
	return hex.EncodeToString(m.Sum(nil))
}

// 获取随机数
func getRand(start, end int) (r int) {
	t := end - start
	r = rand.Intn(t) + start
	return
}

func pushTestData() {
	listKey := global.GetRdbKey(gn, bizid)
	piArr := [...]string{
		"1fffbjzos82bs9cnyj1dna7d6d29zg4esnh99u",
		"1fffbkmd9ebtwi7u7f4oswm9li6twjydqs7qjv",
		"1fffblf892i0p1zh6wlec2quukxtw29v4yismp",
		"1fffbmr55j92gttv5wxspm0mgvw8x3p0n7cy0j",
		"1fffbjqfba5y6uwr55cdak6faokhm4s02qkyue",
		"1fffbkrwndszes1sngfx3v6pdqh87fi4zhz9ur",
		"1fffbl6st3fbp199i8zh5ggcp84fgo3rj7pn1y",
		"1fffbmzwmr1k3y8bri2linqbhnvmu510u5jj6z",
	}

	ctArr := [...]int{
		0,
		2,
	}
	pi := piArr[rand.Intn(8)]
	ct := ctArr[getRand(0, 2)]
	if 2 == ct {
		pi = ""
	}
	// 模拟数据
	/****
	 * si           游戏内部会话标识
	 * bt           用户行为类型 0 1  0：下线\登出；1：上线\登入},
	 * ot           行为发生时间
	 * ct           {上报类型，0：已认证通过用户；2：游客用户},
	 * di           设备标识
	 * pi           用户唯一标识
	 */
	newData := map[string]interface{}{}
	newData["si"] = getmd5(pi)
	newData["bt"] = getRand(0, 2)
	newData["ot"] = time.Now().Unix() - int64(getRand(0, 60))
	newData["ct"] = ct
	newData["di"] = getmd5(pi)
	newData["pi"] = pi
	newStr, _ := json.Marshal(newData)
	ctx := context.Background()
	_, err := global.Rdb.LPush(ctx, listKey, string(newStr)).Result()
	if err != nil {
		panic(err)
	}
}
