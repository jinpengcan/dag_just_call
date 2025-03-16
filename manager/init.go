package manager

import (
	"context"


/* 《思考》
怎么去初始化'load&manager'的全局对象/工具比较好呢？（全局=进程生命周期）
现在是拆成多个Init函数，供使用者自由选择。但是Init函数太多了，成本也大。
*/

// 兜底的缓存能力；defaultCache为localCache
func InitDefaultCache(useLocalSize int64) {
	// 小于32M不设置兜底的localCache
	if useLocalSize < 32 {
		return
	}
	defaultCache = initBigCache(useLocalSize)
}

// 打点能力
func InitMetric(metricEmitCounter, metricEmitLatencyer func(context.Context, string, int64, map[string]string))  {
	utils.SetMetric(metricEmitCounter, metricEmitLatencyer)
}

// 自定义日志能力；会覆盖兜底的日志能力；要求四个入参均!=nil；\
// 注意的是，日志自定义后再执行InitDefault会覆盖回defaul日志实现。
func InitCustomLog(logDebug, logInfo, logWarn, logError func(context.Context, string, ...interface{})) {
	utils.SetCustomLog(logDebug, logInfo, logWarn, logError)
}

// 兜底的日志能力设置日志输出文件
func InitDefaultLogFile(path string) {
	utils.SetDefaultLogFile(path)
}

// 兜底的日志能力设置日志前缀
func InitDefaultLogPrefix(f func(ctx context.Context) string) {
	utils.SetDefaultLogPrefix(f)
}

// 兜底的日志能力设置日志输出文件、前缀
func InitDefaultLogFilePrefix(path string, f func(ctx context.Context) string) {
	utils.SetDefaultLogPrefix(f)
	utils.SetDefaultLogFile(path)
}
