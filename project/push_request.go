package project

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.babeltime.com/packagist/blogger"
	"sync"
	"time"
	"upload_fcm_data/fcm"
	"upload_fcm_data/global"
	"upload_fcm_data/helper"
	"upload_fcm_data/types"
)

var ctx = context.Background()

type pushRequest struct {
	Logger     *blogger.BLogger
	project    types.ProjectType
	mu         sync.Mutex
	wg         *sync.WaitGroup
	fcm        fcm.Fcm
	signalChan chan bool
}

/**
新建类对象
*/
func NewPushRequest(project types.ProjectType, wg *sync.WaitGroup, signalChan chan bool) pushRequest {

	helper.CheckParameter(project.Game)
	helper.CheckParameter(project.Bizid)
	helper.CheckParameter(project.Appid)
	helper.CheckParameter(project.SecretKey)

	log := blogger.NewBlogger(global.Config.Logger.Filepath, global.Config.Logger.Level)
	return pushRequest{
		Logger:     &log,
		project:    project,
		wg:         wg,
		mu:         sync.Mutex{},
		fcm:        fcm.NewFcm(project.Appid, project.Bizid, project.SecretKey),
		signalChan: signalChan,
	}
}

/**
上传数据
*/
func (p *pushRequest) PushAction(taskCode string) {
	p.Logger.Info(fmt.Sprintf("start Project :%v", p.project.Game))
	defer p.wg.Done()
	for {
		p.Logger.Flush()
		select {
		case <-p.signalChan:
			fmt.Println("exit")
			return
		default:
			break
		}
		p.mu.Lock()
		behaviorList := p.getFcmBehaviorList()
		p.mu.Unlock()
		if len(behaviorList) == 0 {
			time.Sleep(time.Second * 5)
			continue
		}
		var result fcm.Result
		var err error
		if global.DEBUG {
			result, err = p.fcm.TestLoginOrOut(behaviorList, taskCode)
		} else {
			result, err = p.fcm.LoginOrOut(behaviorList)
		}
		if err != nil {
			jsonFailBehaviorList, _ := json.Marshal(behaviorList)
			global.Rdb.LPush(context.Background(), global.GetFailRdbKey(p.project.Game, p.project.Bizid), jsonFailBehaviorList)
			p.Logger.Fatal(err)
			continue
		}
		p.Logger.Info(result)
	}
}

func (p *pushRequest) PushFailList(taskCode string)  {
	p.Logger.Info(fmt.Sprintf("start Project :%v", p.project.Game))
	defer p.wg.Done()
	for {
		p.Logger.Flush()
		select {
		case <-p.signalChan:
			fmt.Println("exit")
			return
		default:
			break
		}
		p.mu.Lock()
		behaviorFailList := p.getFcmFailBehaviorList()
		p.mu.Unlock()
		if len(behaviorFailList) == 0 {
			time.Sleep(time.Second * 5)
			continue
		}
		var result fcm.Result
		var err error
		if global.DEBUG {
			result, err = p.fcm.TestLoginOrOut(behaviorFailList, taskCode)
		} else {
			result, err = p.fcm.LoginOrOut(behaviorFailList)
		}
		if err != nil {
			jsonFailBehaviorList, _ := json.Marshal(behaviorFailList)
			global.Rdb.LPush(context.Background(), global.GetFailRdbKey(p.project.Game, p.project.Bizid), jsonFailBehaviorList)
			p.Logger.Fatal(err)
			continue
		}
		p.Logger.Info(result)
	}
}
