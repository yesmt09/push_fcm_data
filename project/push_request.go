package project

import (
	"context"
	"fmt"
	"gitlab.babeltime.com/packagist/blogger"
	"sync"
	"time"
	"upload_fcm_data/fcm"
	"upload_fcm_data/global"
	"upload_fcm_data/types"
)

var ctx = context.Background()

type pushRequest struct {
	logger       *blogger.BLogger
	project      types.ProjectType
	mu           sync.Mutex
	wg           *sync.WaitGroup
	behaviorList *[]fcm.Behavior
	fcm          *fcm.Fcm
	signalChan   chan bool
}

/**
新建类对象
*/
func NewPushRequest(project types.ProjectType, wg *sync.WaitGroup, signalChan chan bool) pushRequest {
	log := blogger.NewBlogger(global.Config.Logger.Filepath, global.Config.Logger.Level)
	return pushRequest{
		logger:       &log,
		project:      project,
		wg:           wg,
		mu:           sync.Mutex{},
		behaviorList: &[]fcm.Behavior{},
		fcm:          fcm.NewFcm(project.Appid, project.Bizid, project.SecretKey),
		signalChan:   signalChan,
	}
}

/**
上传数据
*/
func (p *pushRequest) PushAction() {
	p.logger.AddBase("gn", p.project.Game)
	p.logger.Info(fmt.Sprintf("start Project :%v", p.project.Game))
	p.logger.Flush()
	defer p.wg.Done()
	for {
		select {
		case <-p.signalChan:
			fmt.Println("exit")
			return
		default:
			break
		}
		p.mu.Lock()
		p.behaviorList = p.getFcmBehaviorList()
		p.mu.Unlock()
		if len(*p.behaviorList) == 0 {
			p.logger.Info("list empty")
			time.Sleep(time.Second * 5)
			p.logger.Flush()
			continue
		}
		p.logger.Info(fmt.Sprintf("push list %v", p.behaviorList))

		var result fcm.Result
		var err error
		if global.DEBUG {
			result, err = p.fcm.TestLoginOrOut(p.behaviorList, "")
		} else {
			result, err = p.fcm.LoginOrOut(p.behaviorList)
		}
		if err != nil {
			fmt.Println(result)
			continue
		}
		p.logger.Info(result.Data)
		p.logger.Flush()
	}
}
