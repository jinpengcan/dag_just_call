package biz_demo

import (
	"context"
	"fmt"
	"time"


// ***********************Job4************************//
func NewJob4(ctx context.Context) job_def.Jober {
	return &Job4{}
}

type Job4 struct{}

func (job *Job4) GetName() string {
	return Job4Name
}

func (job *Job4) Do(ctx context.Context, agent manager.IDataAgenter) (interface{}, error) {
	GetJob2Result(ctx, agent)
	job3Result := GetJob3Result(ctx, agent)
	time.Sleep(time.Second)
	if needLog(ctx) {
		fmt.Println(time.Now(), "（job4 依赖 job3+job2）（耗时1s）执行完成 job4 的内容-----", job3Result)
	}
	return 0.12, fmt.Errorf("job4 error test")
}
