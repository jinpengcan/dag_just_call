package biz_demo

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

/*
TestLoad
集成测试：不带缓存能力；
| job1耗时1s
	   | job2耗时0s
	   | job3耗时0.5s
					  | job4耗时1s
*/
func TestLoad(t *testing.T) {
	manager.InitDefaultCache(32)
	//manager.InitCustomLog(logs.CtxDebug, logs.CtxInfo, logs.CtxWarn, logs.CtxError)
	//manager.InitDefaultLogPrefix(func(ctx context.Context) string {return "ppppccc"})

	ctx := context.WithValue(context.Background(), "log", "1")
	conf := &constant.LoadJobConfig{
		Timeout:  5 * 1000,
		LogLevel: 0,
		JobConfigs: map[string]constant.JobConfig{
			Job2Name: constant.JobConfig{
				Timeout:   0,
				Downgrade: false,
			},
		},
	}
	l := loader.NewLoader(ctx,
		[]job_def.JobConstructor{NewJob1, NewJob2, NewJob3, NewJob4},
		manager.WithCache(nil),
		manager.WithConfig(conf))
	fmt.Println(time.Now(), "开始load")
	if err := l.Load(ctx, []int64{11, 22, 33, 44}); err != nil {
		fmt.Println(time.Now(), "完成 load error is", err)
	} else {
		fmt.Println(time.Now(), "完成")
	}
	time.Sleep(time.Second)

	//
	//// 开始第二次执行
	//fmt.Println(time.Now(), "\n\n\n\n开始第二次load")
	//if err := l.Load(ctx, dataAgent); err != nil {}
	//time.Sleep(time.Second)
}

/*
TestParallelLoad
并发测试：
*/
func TestParallelLoad(t *testing.T) {
	ctx := context.Background()
	conf := &constant.LoadJobConfig{
		Timeout:  5 * 1000,
		LogLevel: 0,
		JobConfigs: map[string]constant.JobConfig{
			Job2Name: constant.JobConfig{
				Timeout:   0,
				Downgrade: false,
			},
		},
	}

	var count int64
	for {
		if atomic.LoadInt64(&count) >= 10 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		//go func() {
		atomic.AddInt64(&count, 1)
		l := loader.NewLoader(ctx,
			[]job_def.JobConstructor{NewJob1, NewJob2, NewJob3, NewJob4},
			manager.WithConfig(conf))
		l.Load(ctx, []int64{100, 200, 300, 400})
		fmt.Printf("TestParallelLoad goroutine count: %v \n\n\n", atomic.LoadInt64(&count))
		//atomic.AddInt64(&count, -1)
		//}()

	}
}

func needLog(ctx context.Context) bool {
	r, ok := ctx.Value("log").(string)
	if !ok || r != "1" {
		return false
	}
	return true
}
