package loader

import (
	"context"


func runJob(ctx context.Context, job job_def.Jober, mgr manager.IManager) error {
	res, err := job.Do(ctx, manager.NewDataAgent(job.GetName(), mgr))
	mgr.SetJobResult(job.GetName(), res)
	return err
}

func runSimpleCache(ctx context.Context, simpleCache job_def.SimpleCache, job job_def.Jober, mgr manager.IManager) error {
	if simpleCache.SimpleCacheKey(ctx, mgr) == "" || !mgr.CacheAble() {
		return runJob(ctx, job, mgr)
	}

	data := mgr.GetCache(ctx, simpleCache.SimpleCacheKey(ctx, mgr))
	if len(data) > 0 {
		res, err := simpleCache.Cache2JobResult(ctx, data)
		if err == nil {
			mgr.SetJobResult(job.GetName(), res)
			return nil
		}
	}

	res, err := job.Do(ctx, manager.NewDataAgent(job.GetName(), mgr))
	if err != nil {
		return err
	}
	mgr.SetJobResult(job.GetName(), res)
	mgr.SetCache(ctx, simpleCache.SimpleCacheKey(ctx, mgr), utils.MarshalWithNoErr(res))
	return nil
}

func runComplexIdsCache(ctx context.Context, complexIdsCache job_def.ComplexIdsCache, job job_def.Jober, mgr manager.IManager) error {
	if len(complexIdsCache.IdsCacheKey(ctx, mgr)) == 0 || !mgr.CacheAble() {
		return runJob(ctx, job, mgr)
	}

	// 获取缓存数据
	allKeys := complexIdsCache.IdsCacheKey(ctx, mgr)
	cacheRes := mgr.MGetCache(ctx, allKeys)

	var hitKeys []string
	for key, _ := range cacheRes {
		hitKeys = append(hitKeys, key)
	}

	missKeys := utils.StringArrayDiff(allKeys, hitKeys)
	// 不缺少数据则直接返回
	if len(missKeys) == 0 {
		res, err := complexIdsCache.IdsCache2JobResult(ctx, cacheRes)
		if err != nil {
			utils.LogError(ctx, "[runComplexIdsCache]IdsCache2JobResult error: %v", err)
			return err
		}
		mgr.SetJobResult(job.GetName(), res)
		return nil
	}

	// 缺少数据，需要继续执行job.Do
	ctx = job_def.WrapComplexIdsCacheMissKeys(ctx, missKeys)
	doRes, err := job.Do(ctx, manager.NewDataAgent(job.GetName(), mgr))
	if err != nil {
		return err
	}

	// merge&set
	res, err := complexIdsCache.MergeIdsCacheAndJobResult(ctx, cacheRes, doRes)
	if err != nil {
		return err
	}
	mgr.SetJobResult(job.GetName(), res)

	// 缓存miss数据
	newCache, err := complexIdsCache.JobResult2IdsCache(ctx, doRes)
	if err == nil {
		mgr.MSetCache(ctx, newCache)
	} else {
		utils.LogError(ctx, "JobResult2IdsCache error: %v", err)
	}
	return nil
}
