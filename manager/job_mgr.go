package manager

import (
	"context"
	"sync"
	"sync/atomic"


type jobMgr struct {
	jobsChan        map[string]chan struct{}
	missJobLock     sync.Mutex
	loaderRun       *int32
	missJobListener chan *MissJob
}

func (j *jobMgr) getJobChannel(jobName string) (chan struct{}, bool) {
	ch, ok := j.jobsChan[jobName]
	return ch, ok
}

func (j *jobMgr) waitResult(ctx context.Context, jobName string, agent IManager) {
	if atomic.LoadInt32(j.loaderRun) != constant.LoaderStateRunning {
		return
	}

	ch, ok := j.getJobChannel(jobName)
	if !ok { // mapper 后未注册到 loader，说明 jobs 间的依赖没写全，进行报警
	// todo 如何更好的报警？因为懒加载是一种可插拔能力
		utils.LogError(ctx,  "Loader()的jobs参数不全，需要额外对%v进行missJob懒加载", jobName)
		ch = j.loadMissJob(ctx, jobName, agent)
	}
	<-ch
}

type MissJob struct {
	JobName string
	Ch      chan struct{}
	Agent   IManager
	Ctx     context.Context
}

func (j *jobMgr) loadMissJob(ctx context.Context, missJobName string, agent IManager) chan struct{} {
	j.missJobLock.Lock()
	if ch, ok := j.jobsChan[missJobName]; ok {
		j.missJobLock.Unlock()
		return ch
	}

	// copy on write
	newJobsChan := make(map[string]chan struct{}, len(j.jobsChan)+1)
	newJobsChan[missJobName] = make(chan struct{})
	for name, ch := range j.jobsChan {
		newJobsChan[name] = ch
	}
	j.jobsChan = newJobsChan  // map是指针，才可以原子性赋值
	j.missJobLock.Unlock()

	// 设置全局的监听器
	// 为什么不实现为请求纬度的监听器？在这里往channel写数据，但是load结束时关掉通信channel的话会panic、关掉监听器的瞬间还会遗漏missJob且导致channel阻塞
	j.missJobListener <- &MissJob{
		JobName: missJobName,
		Ch:      j.jobsChan[missJobName],
		Agent:   agent,
		Ctx:     ctx,
	}
	return j.jobsChan[missJobName]
}
