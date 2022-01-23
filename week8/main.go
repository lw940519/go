package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/extra/redisotel"
	redis "github.com/go-redis/redis/v8"
)

func main() {
	redisConf, err := NewConfig("./week8/redisConfig.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := NewRdbData(redisConf)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer closerRdb()

	ctx := context.Background()

	setRdbValueMemory(db.Rdb, ctx, 10)   // 10字节
	setRdbValueMemory(db.Rdb, ctx, 20)   // 20字节
	setRdbValueMemory(db.Rdb, ctx, 50)   // 50字节
	setRdbValueMemory(db.Rdb, ctx, 100)  // 100字节
	setRdbValueMemory(db.Rdb, ctx, 200)  // 200字节
	setRdbValueMemory(db.Rdb, ctx, 1024) // 1024字节
	setRdbValueMemory(db.Rdb, ctx, 5120) // 5120字节
}

// setRdbValueMemory 开启一次测试
func setRdbValueMemory(rdb *redis.Client, ctx context.Context, size int) {
	rdb.FlushAll(ctx)
	startMemory, err := getUserMemory(rdb, ctx)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("1万个%d字节的value插入前内存：%d\n", size, startMemory)
	}
	buf := make([]byte, size)
	setValue(rdb, ctx, fmt.Sprintf("%d_", size), string(buf), 10000)
	total, avgmem, err := getAverageMemory(rdb, ctx, startMemory)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("1万个%d字节的value插入后单个key插入后内存：%d\n", size, total)
		fmt.Printf("1万个%d字节的value插入后单个key平均内存：%d\n", size, avgmem)
	}
}

// setValue 往redis填充测试数据
func setValue(rdb *redis.Client, ctx context.Context, key string, value string, count int) {
	errCount := 0
	for i := 0; i < count; i++ {
		err := rdb.Set(ctx, fmt.Sprintf("%s%d", key, i), value, 0).Err()
		if err != nil {
			errCount++
		}
	}
	fmt.Printf("%v 插入失败个数：%d\n", key, errCount)
}

// getUserMemory 得到总内存
func getUserMemory(rdb *redis.Client, ctx context.Context) (int, error) {
	var totalMemory int
	var err error

	// 获取内存
	result := rdb.Info(ctx, "Memory").Val()
	scanner := bufio.NewScanner(
		strings.NewReader(result),
	)
	var bufT []byte
	for scanner.Scan() {
		bufT = scanner.Bytes() // scanner.Bytes()
		if bufT != nil {
			bufS := string(bufT)
			if strings.Contains(bufS, "used_memory:") {
				totalMemory, err = strconv.Atoi(strings.Replace(bufS, "used_memory:", "", -1))
				if err != nil {
					return 0, err
				} else {
					return totalMemory, nil
				}
			}
		}
	}
	return 0, errors.New("得到总内存失败")
}

// getAverageMemory 得到内存总量以及单个key的内存
func getAverageMemory(rdb *redis.Client, ctx context.Context, startMemory int) (int, int, error) {
	var totalMemory, keys int
	var err error

	// 获取内存
	result := rdb.Info(ctx, "Memory").Val()
	scanner := bufio.NewScanner(
		strings.NewReader(result),
	)
	var bufT []byte
	for scanner.Scan() {
		bufT = scanner.Bytes() // scanner.Bytes()
		if bufT != nil {
			bufS := string(bufT)
			if strings.Contains(bufS, "used_memory:") {
				totalMemory, err = strconv.Atoi(strings.Replace(bufS, "used_memory:", "", -1))
				if err != nil {
					return 0, 0, err
				} else {
					break
				}
			}
		}
	}
	// 获取key的数量
	keysResult := rdb.Info(ctx, "Keyspace").Val()
	items := strings.Split(keysResult, ",")
	for _, val := range items {
		if strings.Contains(val, "# Keyspace\r\ndb0:keys") { // 默认只使用db0
			keys, err = strconv.Atoi(strings.Replace(val, "# Keyspace\r\ndb0:keys=", "", -1))
			if err != nil {
				return 0, 0, err
			} else {
				break
			}
		}
	}
	if keys != 0 {
		return totalMemory, (totalMemory - startMemory) / keys, nil
	} else {
		return 0, 0, errors.New("no key")
	}
}

type RdbData struct {
	Rdb *redis.Client
}

var closerRdb func() error

func NewRdbData(config *RedisConfig) (*RdbData, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.Db,
		DialTimeout:  config.DialTimeout,
		WriteTimeout: config.WriteTimeout,
		ReadTimeout:  config.ReadTimeout,
	})
	rdb.AddHook(redisotel.TracingHook{}) //
	r := &RdbData{
		Rdb: rdb,
	}
	closerRdb = func() error {
		if err := r.Rdb.Close(); err != nil {
			fmt.Printf("RdbData 关闭失败 错误原因:%v\n", err)
			return err
		}
		return nil
	}
	return r, nil
}

type RedisConfig struct {
	Network      string        `json:"network"`
	Addr         string        `json:"addr"`
	Password     string        `json:"password"`
	Db           int           `json:"db"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

func NewConfig(path string) (*RedisConfig, error) {
	confByte, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(confByte) <= 0 {
		return nil, errors.New("获取redis配置错误")
	}

	var cif RedisConfig
	err = json.Unmarshal(confByte, &cif)
	return &cif, err
}
