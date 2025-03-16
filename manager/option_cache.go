package manager

import (
	"context"
	"errors"
	"github.com/allegro/bigcache"
	"time"

type cache interface {
	// 因为认为是弱依赖，所以不阻塞流程、不对外暴露error，在本方处理error
	// todo 但是需要打点统计：miss、hit率
	SetCache(context.Context, string, string)
	GetCache(context.Context, string) string
	MSetCache(context.Context, map[string]string)
	MGetCache(context.Context, []string) map[string]string
}

var defaultCache cache = nil

/*
WithCache: 给data agent设置cache能力
 - cache: 业务自定义实现的cache层
 - 若cache=nil，则使用defaultLocalCache对象
*/
func WithCache(cache cache) ManageOption {
	return func(agent *manager) {
		// 使用传入的自定义cache
		if cache != nil {
			agent.cache = cache
			return
		}
		// 使用默认的LocalCache
		if defaultCache != nil {
			agent.cache = defaultCache
		}
	}
}

func initBigCache(useLocalSize int64) cache {
	cache, err := bigcache.NewBigCache(bigcache.Config{
		// 设置分区的数量，必须是2的整倍数
		Shards: 1024,
		// LifeWindow后，缓存对象被认为不活跃，但并不会删除对象
		LifeWindow: 4 * time.Second,
		// CleanWindow后，会删除被认为不活跃的对象，0代表不操作
		CleanWindow: 2 * time.Second,
		// 设置最大存储对象数量，仅在初始化时可以设置
		MaxEntriesInWindow: 100000,
		// 缓存对象的最大字节数，仅在初始化时可以设置
		MaxEntrySize: 0,
		// 是否打印内存分配信息
		//Verbose: true,
		// 设置缓存最大值(单位为MB)，0表示无限制
		HardMaxCacheSize: int(useLocalSize),
		// 在缓存过期或者被删除时,可设置回调函数，参数是(key、val)，默认是nil不设置
		//OnRemove: callBack,
		// 在缓存过期或者被删除时,可设置回调函数，参数是(key、val,reason)默认是nil不设置
		//OnRemoveWithReason: nil,
	})
	if err != nil {
		panic("manager bigcache.NewBigCache fail")
	}
	return &defaultBigCache{
		BigCache: cache,
	}
}

// "github.com/allegro/bigcache"
type defaultBigCache struct {
	*bigcache.BigCache
}

func (cache *defaultBigCache) SetCache(ctx context.Context, key string, res string) {
	if err := cache.Set(key, []byte(res)); err != nil {
		utils.LogWarn(ctx, "[defaultBigCache.SetCache fail] key: %v; val: %v; error: %v", key, res, err)
	}
}

func (cache *defaultBigCache) GetCache(ctx context.Context, key string) string {
	res, err := cache.Get(key)
	if err != nil && (!errors.Is(err, bigcache.ErrEntryNotFound) || !env.IsProduct()) {
		utils.LogWarn(ctx, "[defaultBigCache.GetCache fail] key: %v; error: %v", key, err)
	}
	return string(res)
}

func (cache *defaultBigCache) MSetCache(ctx context.Context, data map[string]string) {
	for key, val := range data {
		if err := cache.Set(key, []byte(val)); err != nil {
			utils.LogWarn(ctx, "[defaultBigCache.MSetCache fail] key: %v; val: %v; error: %v", key, val, err)
		}
		//{
		//	// debug专用，只随机set一个
		//	logs.CtxDebug(ctx, "[jpc debug]MSetCache %v", key)
		//	break
		//}
	}
}

func (cache *defaultBigCache) MGetCache(ctx context.Context, keys []string) map[string]string {
	res := make(map[string]string, len(keys))
	for _, key := range keys {
		val, err := cache.Get(key)
		if err != nil && (!errors.Is(err, bigcache.ErrEntryNotFound) || !env.IsProduct()) {
			utils.LogWarn(ctx, "[defaultBigCache.MGetCache fail] key: %v; error: %v", key, err)
			continue
		}
		res[key] = string(val)
	}
	return res
}

type defaultBytesCache struct {
}
