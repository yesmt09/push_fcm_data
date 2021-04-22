package main

import (
	"flag"
	"fmt"
	"gitlab.babeltime.com/packagist/blogger"
	"os/signal"
	"sync"
	"syscall"
	"upload_fcm_data/global"
	"upload_fcm_data/initialize"
	"upload_fcm_data/project"
)

var wgall []*sync.WaitGroup
var signalList []chan bool

func main() {
	flag.BoolVar(&global.DEBUG, "debug", false, "")
	flag.Parse()
	signal.Notify(global.Signal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	if global.DEBUG {
		global.ConfigFilePath = "config.debug.json"
		global.Config.Logger.Level = blogger.L_DEBUG
	} else {
		global.ConfigFilePath = "config.json"
		global.Config.Logger.Level = blogger.L_INFO
	}

	initialize.InitConfig()
	defer global.Rdb.Close()
	for _, v := range global.Config.Project {
		var wg sync.WaitGroup
		wgall = append(wgall, &wg)
		signalChan := make(chan bool, 1)
		signalList = append(signalList, signalChan)
		wg.Add(1)
		projectConfig := v
		pushRequest := project.NewPushRequest(projectConfig, &wg, signalChan)
		go pushRequest.PushAction()
	}
	select {
	case <-global.Signal:
		for _, v := range signalList {
			v <- true
		}
	}

	defer func() {
		fmt.Println("end")
	}()
	for _, v := range wgall {
		v.Wait()
	}
}
