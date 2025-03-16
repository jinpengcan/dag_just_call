package biz_demo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"



// ***********************Job1************************//
func NewJob1(ctx context.Context) job_def.Jober {
	return &Job1{}
}

type Job1 struct{}

type bizDefineResult1 struct {
	S string
	I int64
}

func (job *Job1) GetName() string {
	return Job1Name
}

func (job *Job1) Do(ctx context.Context, agent manager.IDataAgenter) (interface{}, error) {
	time.Sleep(time.Second)
	if needLog(ctx) {
		fmt.Println(time.Now(), "（耗时1s）执行完成 job 1 的内容")
	}
	return &bizDefineResult1{"Job1 的result", 1}, nil
}

func (job *Job1) SimpleCacheKey(ctx context.Context, agent manager.IDataAgenter) string {
	// 将request转为cache key
	// req := agent.GetRequest().(some request)
	// id := req.Id
	// return fmt.Sprintf("key_%d", id)
	return "job1_result"
}

func (job *Job1) Cache2JobResult(ctx context.Context, data string) (interface{}, error) {
	var res *bizDefineResult1
	err := json.Unmarshal([]byte(data), &res)
	return res, err
}
