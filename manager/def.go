package manager

import (
	"context"



type IManager interface {
	GetRequest() interface{}
	GetJobResult(ctx context.Context, jobName string) interface{}
	Clear(req interface{})

	/* job manager */
	SetJobMgr(ctx context.Context, jobsChan map[string]chan struct{}, missJobListener chan *MissJob, loaderRun *int32)
	SetJobResult(jobName string, res interface{})
	GetConfig() constant.LoadJobConfig

	/* manager */
	// job依赖关系
	drawJobDepend(call, depended string)
	GetJobDependGraph() map[string][]string
	// 简单的单key缓存能力
	CacheAble() bool
	SetCache(context.Context, string, string)
	GetCache(context.Context, string) string
	// 多id的复杂缓存能力
	MSetCache(context.Context, map[string]string)
	MGetCache(context.Context, []string) map[string]string
}

type IDataAgenter interface {
	GetRequest() interface{}
	GetJobResult(ctx context.Context, jobName string) interface{}
}
