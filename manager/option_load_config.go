package manager

import (
	"fmt"


/*
	load、job所需要的配置，不需要外部直传config结构体了；
	也提供多种方式，供使用者方便使用；也减少结构体对外的暴露。
 */

func WithConfig(conf *constant.LoadJobConfig) ManageOption {
	return func(agent *manager) {
		agent.config = conf
	}
}

func WithConfigByStr(conf string) ManageOption {
	return func(agent *manager) {
		if len(conf) == 0 {
			return
		}
		err := utils.Unmarshal([]byte(conf), agent.config)
		if err != nil {
			panic(fmt.Errorf("<WithConfigByStr> Unmarshal error: %v", err))
		}
	}
}

func WithConfigByFile(path string) ManageOption {
	return func(agent *manager) {

	}
}
