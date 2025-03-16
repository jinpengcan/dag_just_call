package biz_demo

import (
	"context"
	"fmt"
	"time"


// ***********************Job3************************//
func NewJob3(ctx context.Context) job_def.Jober {
	return &Job3{}
}

type Job3 struct{}

func (job *Job3) GetName() string {
	return Job3Name
}

func (job *Job3) Do(ctx context.Context, agent manager.IDataAgenter) (interface{}, error) {
	agent.GetJobResult()

	job1Result := GetJob1Result(ctx, agent)
	time.Sleep(time.Millisecond * 500)
	if needLog(ctx) {
		fmt.Println(time.Now(), "（job3 依赖 job1）（耗时0.5s）执行完成 job3 的内容-----", job1Result)
	}
	return "job3的result33333", nil
}

func (job *Job3) RequestKeys(ctx context.Context, agent manager.IDataAgenter) []string {
	req := agent.GetRequest().([]int64)
	var keys []string
	for _, id := range req {
		keys = append(keys, fmt.Sprintf("%v:hot_tag:%d", job.GetName(), id))
	}
	return keys
}

func (job *Job3) GetHotKeys(ctx context.Context) []string {
	res := job_def.UnwrapHotKeys(ctx)
	return res
}

func (job *Job3) HotKeyLimit() int64 {
	return 5
}

func (job *Job3) DealHotBefore(ctx context.Context) {
	fmt.Printf("job3 触发 DealHotBefore。热key有：%v", job.GetHotKeys(ctx))
}
