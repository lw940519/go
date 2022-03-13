package gorm

type GormConfig struct {
	DBType          string // 数据库类型
	DSN             string // 数据源连接字符串
	MaxOpen         int    // 连接池最大连接数
	MaxIdle         int    // 连接池最大空闲数
	ConnMaxLifetime int    // 连接最大存活时间
	LogMode         bool   // 日志模式，true：详细，false:无日志，default：只有错误
}
