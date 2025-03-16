package biz_demo

import (
	"context"
	"fmt"
	"strconv"
	"time"


func NewJob2(ctx context.Context) job_def.Jober {
	return &Job2{}
}

type Job2 struct{}

func (job *Job2) Do(ctx context.Context, agent manager.IDataAgenter) (interface{}, error) {
	job1Result := GetJob1Result(ctx, agent)
	if needLog(ctx) {
		fmt.Println(time.Now(), "（job2 依赖 job1）执行完成 job 2 的内容-----", job1Result)
	}
	// fmt.Println("读取job2的miss key：", job.GetMissKeys(ctx, agent))
	return []int64{11,22,33,44}, nil
}

func (job *Job2) GetName() string {
	return Job2Name
}

func (job *Job2) IdsCacheKey(ctx context.Context, dataAgent manager.IDataAgenter) []string {
	req := dataAgent.GetRequest().([]int64)
	var keys []string
	for _, id := range req {
		keys = append(keys, fmt.Sprintf("Job2_IdCache_%d", id))
	}
	return keys
}

func (job *Job2) GetMissKeys(ctx context.Context) []string {
	res := job_def.UnwrapComplexIdsCacheMissKeys(ctx)
	return res
}

func (job *Job2) MergeIdsCacheAndJobResult(ctx context.Context, cache map[string]string, res interface{}) (interface{}, error) {
	r := res.([]int64)
	for _, v := range cache {
		i, err := strconv.Atoi(v)
		if err != nil {
			utils.LogError(ctx, "MergeIdsCacheAndJobResult strconv.Atoi error: %v", err)
			continue
		}
		r = append(r, int64(i))
	}
	return r, nil
}

func (job *Job2) JobResult2IdsCache(ctx context.Context, res interface{}) (map[string]string, error) {
	r := res.([]int64)
	m := map[string]string{}
	for _, v := range r {
		m[fmt.Sprintf("Job2_IdCache_%d", v)] = strconv.Itoa(int(v))
	}
	return m, nil
}

func (job *Job2) IdsCache2JobResult(ctx context.Context, cache map[string]string,) (interface{}, error) {
	var r []int64
	for _, v := range cache {
		i, err := strconv.Atoi(v)
		if err != nil {
			utils.LogError(ctx, "IdsCache2JobResult strconv.Atoi error: %v", err)
			continue
		}
		r = append(r, int64(i))
	}
	return r, nil
}
