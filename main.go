package main

import (
	"flag"
	"fmt"
	"github.com/pkg/profile"
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
	var taskCode string
	flag.BoolVar(&global.DEBUG, "debug", false, "debug mode")
	flag.StringVar(&taskCode, "testcode", "", "test code")
	flag.Parse()
	signal.Notify(global.Signal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	if global.DEBUG {
		defer profile.Start().Stop()
		if taskCode == "" {
			panic("debug mode must need testcode flag")
		}
		global.ConfigFilePath = "config.debug.json"
		global.Config.Logger.Level = blogger.L_DEBUG
		fmt.Println("debug mode runing...")
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
		wg.Add(2)
		projectConfig := v
		pushRequest := project.NewPushRequest(projectConfig, &wg, signalChan)
		pushRequest.Logger.AddBase("gn", projectConfig.Game)

		go pushRequest.PushAction(taskCode)
		go pushRequest.PushFailList(taskCode)
	}
	select {
	case <-global.Signal:
		for _, v := range signalList {
			v <- true
		}
	}

	defer func() {
		fmt.Println("end")
		global.Logger.Info("end")
	}()
	for _, v := range wgall {
		v.Wait()
	}
}
