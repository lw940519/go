package redis

import (
	"context"
	redis "github.com/go-redis/redis/v8"
	"time"
)

type RedisClient = redis.Client

func NewRedisClient(cif *RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		//连接信息
		Network:  cif.Network,  //网络类型，tcp or unix，默认tcp
		Addr:     cif.Addr,     //主机名+冒号+端口，默认localhost:6379
		Password: cif.Password, //密码
		DB:       cif.DB,       // redis数据库index

		//连接池容量及闲置连接数量
		PoolSize:     cif.PoolSize,     // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: cif.MinIdleConns, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

		//超时
		DialTimeout:  cif.DialTimeout * time.Second,  //连接建立超时时间，默认5秒。
		ReadTimeout:  cif.ReadTimeout * time.Second,  //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: cif.WriteTimeout * time.Second, //写超时，默认等于读超时
		PoolTimeout:  cif.PoolTimeout * time.Second,  //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

		//闲置连接检查包括IdleTimeout，MaxConnAge
		IdleCheckFrequency: cif.IdleCheckFrequency * time.Second, //闲置连接检查的周期，默认为1分钟，-1表示不做周期性检查，只在客户端获取连接时对闲置连接进行处理。
		IdleTimeout:        cif.IdleTimeout * time.Minute,        //闲置超时，默认5分钟，-1表示取消闲置超时检查
		MaxConnAge:         cif.MaxConnAge * time.Second,         //连接存活时长，从创建开始计时，超过指定时长则关闭连接，默认为0，即不关闭存活时长较长的连接

		//命令执行失败时的重试策略
		MaxRetries:      cif.MaxRetries,                         // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: cif.MinRetryBackoff * time.Millisecond, //每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: cif.MaxRetryBackoff * time.Millisecond, //每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

		//可自定义连接函数

		//钩子函数
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	} else {
		return rdb, nil
	}
}
