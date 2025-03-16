package constant

func DefaultConfig() LoadJobConfig {
	return LoadJobConfig {
		Timeout:    2000,
		LogLevel:   1,
	}
}

type LoadJobConfig struct {
	Timeout      int32    `json:"timeout"`   // 毫秒
	LogLevel     int32    `json:"log_level"`
	JobConfigs   map[string]JobConfig `json:"job_configs"`
}

type JobConfig struct {
	IgnoreErr	     bool       `json:"ignore_err"`
	Timeout          int32      `json:"timeout"`   // 毫秒
	Downgrade        bool	    `json:"downgrade"`
	Retry  	         Retry      `json:"retry"`
	CircuitBreakRate int32	    `json:"circuit_break_rate"`  // 熔断门槛：error百分占比*100

	//更多sre等等??....
}

type Retry struct {
	TryCount  int32
	IgnoreCountLimit bool   // 代码中写死的最多重试n次，取消该限制
	TryRatioRule   TryRatioRule // 按比例限制重试 // 必须
}

type TryRatioRule int32
const (
	NotMoreThan10 TryRatioRule = 10  // 默认 // 重试比例不超过10%
	NotMoreThan20 TryRatioRule = 20  // 重试比例不超过20%
	NotMoreThan30 TryRatioRule = 30  // 重试比例不超过30%
)
