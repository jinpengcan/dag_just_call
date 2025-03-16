package loader

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"


var (
	//jobStat = stat.NewJobStat()
	keyStat = stat.NewKeyStat()
)

type Loader struct {
	// jobs 需要执行的任务列表
	jobs []job_def.Jober
	// mgr 内置的load控制器；和job交互的控制器是IManager的子类IDataAgenter
	mgr manager.IManager

	// timeout Load维度的超时时间
	timeout time.Duration
	// logLevel 根据日志等级是否记录日志
	logLevel int32
	// jobDowngrade job维度的降级控制
	jobDowngrade map[string]struct{}
	// jobTimeout job维度的超时时间
	jobTimeout map[string]time.Duration
	// ignoreJobErr 需要忽略错误的job
	ignoreJobErr map[string]bool
}

func NewLoader(ctx context.Context, constructors []job_def.JobConstructor, opts ...manager.ManageOption) *Loader {
	loader := &Loader{
		mgr:		  manager.NewManager(opts...),
		jobDowngrade: map[string]struct{}{},
		jobTimeout:   map[string]time.Duration{},
		ignoreJobErr: map[string]bool{},
	}
	// todo bug，在这里初始化，会导致job变量不重置
	for _, fn := range constructors {
		loader.jobs = append(loader.jobs, fn(ctx))
	}

	// 配置中心：降级、日志等级、load超时、....
	config := loader.mgr.GetConfig()
	loader.timeout = time.Millisecond * time.Duration(config.Timeout)
	loader.logLevel = config.LogLevel
	for jobName, jobConf := range config.JobConfigs {
		if jobConf.Downgrade {
			loader.jobDowngrade[jobName] = struct{}{}
		}
		if jobConf.Timeout > 0 {
			loader.jobTimeout[jobName] = time.Millisecond * time.Duration(jobConf.Timeout)
		}
		if jobConf.IgnoreErr {
			loader.ignoreJobErr[jobName] = true
		}
	}
	if loader.timeout <= 0 {
		loader.timeout = time.Second // 没配置超时的话，则兜底1s
	}

	return loader
}

// Load 并发执行Job列表
func (l *Loader) Load(ctx context.Context, req interface{}) error {
	l.mgr.Clear(req)
	startTime := time.Now()
	jobLen := int32(len(l.jobs))
	var loaderRun int32 = constant.LoaderStateRunning

	// job 相关的
	jobsChan := make(map[string]chan struct{}, len(l.jobs))
	jobsState := make(map[string]*job_def.JobState, len(l.jobs))
	for _, job := range l.jobs {
		jobsState[job.GetName()] = &job_def.JobState{}
		jobsChan[job.GetName()] = make(chan struct{})
	}

	// 设置JobMgr：在数据交互的时候进行任务阻塞管理
	l.mgr.SetJobMgr(ctx, jobsChan, missJobListener, &loaderRun)

	// 开始并发执行 jobs
	completedJobs := make(chan job_def.Jober, jobLen) // jobs will send to this chan after finish running
	for _, job := range l.jobs {
		// job 不可重复执行
		if !atomic.CompareAndSwapInt32(&jobsState[job.GetName()].State, job_def.JOB_STATE_UPCOMING, job_def.JOB_STATE_UNFINISH) {
			utils.LogError(ctx, "[%v] 重复注入，不再重复执行", job.GetName())
			jobLen--
			continue
		}

		// 降级的job；需要避免被误以为missJob；
		if _, ok := l.jobDowngrade[job.GetName()]; ok {
			jobsState[job.GetName()].SetState(job_def.JOB_STATE_DOWNGRADE, startTime)
			completedJobs <- job
			continue
		}

		// 熔断逻辑在这里 todo 用 sentinel（学习sentinel用法）
		// todo

		go func(ctx context.Context, obj interface{}) {
			job, ok := obj.(job_def.Jober)
			if !ok || job == nil {
				panic("[run job fatal] obj is not a job")
			}

			defer func() {
				if e := recover(); e != nil {
					utils.LogError(ctx, "[run job] panic: %s: %s %s", job.GetName(), e, string(debug.Stack()))
					jobsState[job.GetName()].SetState(job_def.JOB_STATE_PANIC, startTime)
				}
				completedJobs <- job
			}()

			run := func(done chan struct{}, _ctx context.Context) error {
				defer func() {
					if done != nil {
						closeChan(done)
					}
				}()

				//// 热key处理
				//if v, ok := job_def.ExtendHotSpot(obj); ok {
				//	// 统计x秒内的频次；即便是触发了热key处理也需要统计，因为需要去判断是否一直"热"；
				//	var keys []string = v.RequestKeys(_ctx, l.mgr)
				//	go keyStat.MetricKeys(_ctx, keys)
				//
				//	// 发现热key后：处理机制、设置热key
				//	if hotKeys := keyStat.HotKeys(_ctx, keys, v.HotKeyLimit()); len(hotKeys) > 0 {
				//		ctx = job_def.WrapHotKeys(ctx, hotKeys) // 使用_ctx会set不进去Valu的bug，todo 要修一下
				//		v.DealHotBefore(ctx)
				//	}
				//}
				if v1, ok := job_def.ExtendSimpleCache(obj); ok {
					return runSimpleCache(_ctx, v1, job, l.mgr)
				} else if v2, ok := job_def.ExtendComplexIdsCache(obj); ok {
					return runComplexIdsCache(_ctx, v2, job, l.mgr)
				} else {
					return runJob(_ctx, job, l.mgr)
				}
			}

			// 重试逻辑在这里，runWithTimeout/run的外层
			// for {}
			// todo

			var err error
			if t, ok := l.jobTimeout[job.GetName()]; ok && t < l.timeout {
				err = runWithTimeout(ctx, t, run)
			} else {
				err = run(nil, ctx)
			}
			if err == job_def.ErrorTimeout {
				jobsState[job.GetName()].SetState(job_def.JOB_STATE_JOB_TIMEOUT, startTime)
			} else if err != nil {
				jobsState[job.GetName()].SetState(job_def.JOB_STATE_ERROR, startTime)
			}
		}(ctx, job)  // todo bug: 这里把context作为参数传入，但是在函数中使用context.Value进行赋值了，这里无法传递。
	}

	// 等待jobs完成，并设置超时
	timeout := time.NewTimer(l.timeout)
	defer timeout.Stop()
	var completedCount int32
WAIT:
	for completedCount < jobLen {
		select {
		case job := <-completedJobs:
			closeChan(jobsChan[job.GetName()])
			jobsState[job.GetName()].SetState(job_def.JOB_STATE_SUCCESS, startTime)
			completedCount += 1

		case <-timeout.C:
			for _, job := range l.jobs {
				closeChan(jobsChan[job.GetName()])
				jobsState[job.GetName()].SetState(job_def.JOB_STATE_LOAD_TIMEOUT, startTime)
			}
			utils.LogWarn(ctx, "execute job count: %v; completed job count: %v; timeout: %+v", jobLen, completedCount, l.timeout)
			atomic.SwapInt32(&loaderRun, constant.LoaderStateTimeout)
			break WAIT
		}
	}
	atomic.CompareAndSwapInt32(&loaderRun, constant.LoaderStateRunning, constant.LoaderStateDone)

	record(ctx, jobsState)
	if l.logLevel < logs.LevelError {
		// todo 日志等级有啥可做的？？？
		// 自定义一个日志生产器？通过日志等级来控制是否打印日志？
		// 可以在这里控制是否离线走链路分析、画图？？？
		utils.LogDebug(ctx, "[jpc debug] job依赖有向图：%v", l.mgr.GetJobDependGraph())
	}
	return wrapLoadErr(ctx, jobsState, l.ignoreJobErr)
}

func wrapLoadErr(ctx context.Context, jobsState map[string]*job_def.JobState, ignoreJobErr map[string]bool) error {
	var err string
	for jobName, state := range jobsState {
		if ignoreJobErr[jobName] {
			continue
		}
		if state.State != job_def.JOB_STATE_SUCCESS && state.State != job_def.JOB_STATE_DOWNGRADE {
			err += fmt.Sprintf("%v | ", jobName)
		}
	}
	if len(err) == 0 {
		return nil
	}
	return fmt.Errorf("run error jobs:[ %v ]", err)
}

func closeChan(ch chan struct{}) bool {
	select {
	case _, _ = <-ch:
		return false
	default:
		close(ch)
		return true
	}
}

func record(ctx context.Context, jobsState map[string]*job_def.JobState) {
	// 日志
	var log strings.Builder
	for jobName, state := range jobsState {
		_, err := fmt.Fprintf(&log, "[%v] %v \n", jobName, state)
		if err != nil {
			utils.LogError(ctx, "job log write err=%+v", err)
			return
		}
	}
	utils.LogInfo(ctx, "jobs load end, show job state:\n%v", log.String())

	// 打点 // todo 改改？打点影响性能，可以用log等级做成开关
	for jobName, jobState := range jobsState {
		utils.MetricEmitLatency(ctx, fmt.Sprintf("job_%s.latency", jobName), int64(jobState.Consume), nil)
		utils.MetricEmitCount(ctx, fmt.Sprintf("job_%s", jobName), 1, nil)
	}
}

func runWithTimeout(ctx context.Context, timeout time.Duration, run func(chan struct{}, context.Context) error) error {
	if timeout <= 0 {
		return run(nil, ctx)
	}

	done := make(chan struct{})
	var err error
	go func() {
		err = run(done, ctx)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return err
	case <-timer.C:
		utils.LogWarn(ctx, "job run timeout")
		return job_def.ErrorTimeout
	}
}
