package pool

import "time"

// HTTP 请求配置常量
const (
	DefaultMaxRetries     = 3              // 默认最大重试次数
	DefaultRetryDelayMS   = 100            // 默认重试延迟（毫秒）
	DefaultTimeoutSeconds = 10             // 默认请求超时时间（秒）
	DefaultTimeout        = 10 * time.Second
)

// 缓存配置常量
const (
	DefaultCacheTTLMin     = 5              // 默认缓存TTL（分钟）
	DefaultCacheTTL        = 5 * time.Minute
	CacheFileMode          = 0644           // 缓存文件权限
)

// 缓存文件名常量
const (
	CoinPoolCacheFile = "latest.json"       // 币种池缓存文件名
	OITopCacheFile    = "oi_top_latest.json" // OI Top缓存文件名
)

// 默认缓存目录
const (
	DefaultCoinPoolCacheDir = "data/cache/coin_pool"
	DefaultOITopCacheDir    = "data/cache/oi_top"
)

// API 响应限制
const (
	MaxCoinsPerResponse = 500 // 单次响应最大币种数量
	MaxOITopItems       = 100 // OI Top最大项目数
)

// 币种过滤规则
const (
	MinSymbolLength = 3     // 最小币种符号长度
	RequiredSuffix  = "USDT" // 必需的后缀
)

// 数据源优先级（用于去重时选择）
const (
	SourcePriorityAI500  = 1 // AI500币种池优先级
	SourcePriorityOITop  = 2 // OI Top优先级
	SourcePriorityDefault = 3 // 默认币种列表优先级
)

// 错误重试策略
const (
	RetryStatusCode5xx = true  // 5xx错误时重试
	RetryStatusCode429 = true  // 429（限流）时重试
	RetryStatusCode408 = true  // 408（超时）时重试
)

// HTTP 状态码
const (
	StatusCodeOK            = 200
	StatusCodeTooManyReqs   = 429
	StatusCodeTimeout       = 408
	StatusCodeInternalError = 500
)

// 日志配置
const (
	LogPrefixCoinPool = "[CoinPool]"
	LogPrefixOITop    = "[OITop]"
	LogPrefixCache    = "[Cache]"
)
