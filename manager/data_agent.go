package manager

import (
	"context"
)

func NewDataAgent(jobName string, manager IManager) IDataAgenter {
	return &dataAgent{jobName, manager}
}

type dataAgent struct {
	jobName string
	manager IManager
}

func (da *dataAgent) GetRequest() interface{} {
	return da.manager.GetRequest()
}

func (da *dataAgent) GetJobResult(ctx context.Context, jobName string) interface{} {
	da.manager.drawJobDepend(da.jobName, jobName)
	return da.manager.GetJobResult(ctx, jobName)
}

/*
	【建议】需要业务额外写转换函数（比如在biz/result目录下），将 interface{} 类型的JobResult转为业务的response struct。
func GetJobXyResult(agent *manager.IDataAgenter) XxYy {
	return agent.GetJobResult("name").(XxYy)
}
*/
