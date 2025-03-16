package biz_demo

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"runtime"


func GetJob1Result(ctx context.Context, agent manager.IDataAgenter) *bizDefineResult1 {
	res := agent.GetJobResult(ctx, Job1Name)
	if res != nil {
		return res.(*bizDefineResult1)
	}
	return nil
}

func GetJob2Result(ctx context.Context, agent manager.IDataAgenter) []int64 {
	res := agent.GetJobResult(ctx, Job2Name)
	if res != nil {
		return res.([]int64)
	}
	return []int64{}
}

func GetJob3Result(ctx context.Context, agent manager.IDataAgenter) string {
	res := agent.GetJobResult(ctx, Job3Name)
	if res != nil {
		return res.(string)
	}
	return ""
}

func getCallerInfo(skip int) (info string) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		info = "runtime.Caller() failed"
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	fileName := path.Base(file) // Base函数返回路径的最后一个元素

	// pc这个指针怎么转成有用的？// 学习反射知识 todo
	typeFunc := reflect.TypeOf(pc)
	fmt.Printf("kind=%v\nis function type: %t\n", typeFunc.Kind(), typeFunc.Kind() == reflect.Func)
	//argInNum := typeFunc.NumIn()   //输入参数的个数
	//argOutNum := typeFunc.NumOut() //输出参数的个数
	//for i := 0; i < argInNum; i++ {
	//	argTyp := typeFunc.In(i)
	//	fmt.Printf("第%d个输入参数的类型%s\n", i, argTyp)
	//}
	//for i := 0; i < argOutNum; i++ {
	//	argTyp := typeFunc.Out(i)
	//	fmt.Printf("第%d个输出参数的类型%s\n", i, argTyp)
	//}

	return fmt.Sprintf("FuncName:%s, file:%s, line:%d", funcName, fileName, lineNo)
}
