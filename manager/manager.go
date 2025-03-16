package manager

import (
	"context"
	"sync"


type ManageOption func(*manager)

func NewManager(opts ...ManageOption) IManager {
	agent := &manager{
		jobResult: map[string]interface{}{},
		jobDependGraph: map[string][]string{},
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

type manager struct {
	// load&job的配置；用户定义，由manager回传到loader
	config *constant.LoadJobConfig
	// jobMgr 任务调度管理器
	*jobMgr
	// cache 缓存能力
	cache
	// request 请求参数
	request interface{}
	// jobResult 各个job执行结果
	jobResult map[string]interface{}
	// jobDependGraph job依赖图
	jobDependGraph map[string][]string
}

func (m *manager) Clear(req interface{}) {
	m.jobResult = map[string]interface{}{}
	m.jobDependGraph = map[string][]string{}
	m.jobMgr = nil
	m.request = req
}

func (m *manager) SetJobMgr(ctx context.Context, jobsChan map[string]chan struct{}, missJobListener chan *MissJob, loaderRun *int32) {
	if m.jobMgr == nil {
		m.jobMgr = &jobMgr{
			jobsChan:        jobsChan,
			missJobListener: missJobListener,
			missJobLock:     sync.Mutex{},
			loaderRun:       loaderRun,
		}
	} else {
		panic("禁止重复执行SetJobMgr")
	}
}

func (m *manager) CacheAble() bool {
	return m.cache != nil
}

func (m *manager) GetConfig() constant.LoadJobConfig {
	if m.config != nil {
		return *m.config
	}
	return constant.DefaultConfig()
}

var setResultLock sync.Mutex

// 1. 通过锁机制，控制了写并发；2. 通过cow机制，将读写分离；
func (m *manager) SetJobResult(jobName string, res interface{}) {
	setResultLock.Lock()
	newJobResult := make(map[string]interface{}, len(m.jobResult)+1)
	newJobResult[jobName] = res
	for name, res := range m.jobResult {
		newJobResult[name] = res
	}
	m.jobResult = newJobResult
	setResultLock.Unlock()
}

func (m *manager) GetRequest() interface{} {
	return m.request
}

func (m *manager) GetJobResult(ctx context.Context, jobName string) interface{} {
	m.waitResult(ctx, jobName, m)  // job.Do结束后才会close channel
	return m.jobResult[jobName]
}

/*
【插曲记录】《如何构建依赖图，出现了瓶颈》
load_job框架并不需要显式指明两个job之间的依赖。而是在某个job获取其他job的result时候自动生成依赖关系（通过channel进行阻塞）。
所以（原先的实现方案）"Job.Do函数入参有IManager"+"使用Manage->jobResult获取任务结果"的数据交互方案，就导致无法识别调用者job是谁，最终导致无法构建依赖关系。

有几种改进思路：
(1)在底层调用函数中，比如waitResult、GetJobResult，通过runtime捞出堆栈信息，然后在上一跳中使用正则匹配找出是否是调用者、调用者信息。
缺点：正则匹配是不准确的，过于依赖代码实现的约定；用户使用函数是不可控的，就会导致这个"找调用者"的逻辑很复杂；runtime还是会影响性能；
(2)把调用者作为参数传进来，函数变成这样：WhoGetJobResult(job job_def.Jober, ctx context.Context, jobName string)
缺点：同样的，用户使用函数是不可控的，如果一参（本job）被错误使用的话，就导致了错乱；这种函数定义和用法看起来就很奇怪；
(3)学习/temai/task_patterns/srv_flow，Jober的Do方法成员的二参dataAgent其实是对应了该job。
二参dataAgent是通过load内部生成和赋值，对用户无感。
更好的是，将job result的获取和manager隔离开，避免了manager暴露、有被乱用的风险。

采用方案(3)。额外的，再增加采样、开关的能力。便于进行离线分析、避免影响性能。

*/

func (m *manager) drawJobDepend(call, depended string) {
	// todo bug
	// fatal error: concurrent map writes
	// m.jobDependGraph[call] = append(m.jobDependGraph[call], depended)
}

func (m *manager) GetJobDependGraph() map[string][]string {
	return m.jobDependGraph
}
