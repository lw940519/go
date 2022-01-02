package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	// 创建一个滑窗
	var slidingWindow SlidingWindowCounter
	slidingWindow = NewWindowCounterTime()
	// 初始化滑窗的时间长度以及桶数量
	err := slidingWindow.SetUp(2, 10)
	if err != nil {
		fmt.Println(err)
	}
	err = slidingWindow.Run()
	if err != nil {
		fmt.Println(err)
	}

	eg, ctx := errgroup.WithContext(context.Background())

	// 监听系统关闭信号
	var waiter = make(chan os.Signal, 1) // 按文档指示，至少设置1的缓冲
	signal.Notify(waiter, syscall.SIGTERM, syscall.SIGINT)

	// 模拟请求1
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("1退出了")
				return nil
			default:
				err := slidingWindow.Event()
				if err != nil {
					return err
				}
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			}
		}
	})

	// 模拟请求2
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("2退出了")
				return nil
			default:
				err := slidingWindow.Event()
				if err != nil {
					return err
				}
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			}
		}
	})

	// 模拟结果遍历
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("3退出了")
				return nil
			default:
				fmt.Println(slidingWindow.Output())
				time.Sleep(time.Second * 1)
			}
		}
	})

	// 等待程序关闭
	eg.Go(func() error {
		select {
		case <-waiter:
			err := slidingWindow.Stop()
			if err != nil {
				return err
			}
		}
		return errors.New("收到Signal信号")
	})

	if err := eg.Wait(); err != nil {
		fmt.Println("退出原因" + err.Error())
	}
	fmt.Println("程序被关闭")
}

var (
	NoRunError   = errors.New("no run error")
	RunIngError  = errors.New("running error")
	InvalidError = errors.New("invalid error")
)

// SlidingWindowCounter 滑窗计数器
type SlidingWindowCounter interface {
	SetUp(timeLength int64, bucketNum int64) error
	Event() error
	Output() (int64, float64, error)
	Run() error
	Stop() error
}

type WindowCounterTime struct {
	timeLength    int64         // 时间长度,单位秒
	bucketNum     int64         // 桶数量
	buckets       []int64       // 滑窗
	index         int64         // 待更新桶下标，循环移动
	num           int64         // 滑窗内事件数量
	bucketFilling int64         // 正在填充的桶
	isRun         int64         // 是否在运行 0未运行,1运行
	invalid       int64         // 输出是否有效，0 无效，1有效
	closeChan     chan struct{} // 运行关闭的通知通道
	lock          sync.Mutex    // 主要时对buckets加锁
}

// NewWindowCounterTime 实现滑动计数器
func NewWindowCounterTime() SlidingWindowCounter {
	return &WindowCounterTime{
		timeLength: 1,
		bucketNum:  10,
		closeChan:  make(chan struct{}, 1),
	}
}

// SetUp 设置滑窗时间长度，桶数量
// 如果滑窗在开始运动前没有执行此方法，滑窗将按照默认值 timeLength=1秒,bucketNum=10进行运动
// 正在允许的窗体不支持设置该值，报错误RunIngError
func (s *WindowCounterTime) SetUp(timeLength int64, bucketNum int64) error {
	if atomic.LoadInt64(&s.isRun) == 1 {
		return RunIngError
	} else {
		atomic.StoreInt64(&s.timeLength, timeLength)
		atomic.StoreInt64(&s.bucketNum, bucketNum)
		return nil
	}
}

// Event 往桶填充事件，每次调用加1，线程安全
// 不再允许的窗体会报NoRunError
func (s *WindowCounterTime) Event() error {
	if atomic.LoadInt64(&s.isRun) == 0 {
		return NoRunError
	} else {
		atomic.AddInt64(&s.bucketFilling, 1)
		return nil
	}
}

// Output 滑窗输出
// 1输出：滑窗事件数量
// 2输出：滑窗统计的QPS
// 3输出：如果滑窗从未运行过，则会报InvalidError
func (s *WindowCounterTime) Output() (int64, float64, error) {
	if atomic.LoadInt64(&s.invalid) == 0 {
		return 0, 0, InvalidError
	} else {
		ti := atomic.LoadInt64(&s.timeLength)
		num := atomic.LoadInt64(&s.num)
		return num, float64(num) / float64(ti), nil
	}
}

// Run 开始运动，运动时支持事件填充
func (s *WindowCounterTime) Run() error {
	if atomic.LoadInt64(&s.isRun) == 1 {
		return RunIngError
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.buckets = make([]int64, s.bucketNum)
	s.index = 0

	atomic.StoreInt64(&s.num, 0)
	atomic.StoreInt64(&s.bucketFilling, 0)
	atomic.StoreInt64(&s.isRun, 1)
	atomic.StoreInt64(&s.invalid, 1)
	timeLength := atomic.LoadInt64(&s.timeLength)
	bucketNum := atomic.LoadInt64(&s.bucketNum)
	ti := float64(timeLength) / float64(bucketNum) * 1000 // 毫秒
	go func() {
		// 做定时
		timeout := make(chan bool, 1)
		timeoutFunc := func() {
			time.Sleep(time.Duration(ti) * time.Millisecond) // 指定超时时长
			timeout <- true
		}
		go timeoutFunc()
		for {
			select {
			case <-timeout:
				s.lock.Lock()
				// 剔除原始数据并从总请求数中减去,插入新聚合桶数据。
				atomic.AddInt64(&s.num, -s.buckets[s.index])
				s.buckets[s.index] = atomic.LoadInt64(&s.bucketFilling)
				// 新桶清洗成0
				atomic.StoreInt64(&s.bucketFilling, 0)
				// 更新总请求数，加上新增的桶的数值
				atomic.AddInt64(&s.num, s.buckets[s.index])
				// 更新桶队列下角标
				count := atomic.LoadInt64(&s.bucketNum)
				s.index++
				if s.index >= count {
					s.index = 0
				}
				s.lock.Unlock()
				// 继续做超时
				go timeoutFunc()
			case <-s.closeChan:
				// 结束运行状态，并退出groutine
				atomic.StoreInt64(&s.isRun, 0)
				break
			}
		}
	}()
	return nil
}

// Stop 停止运动，停止应和run成对出现，任何时候run之后都要记得停止,不然相应资源不会被释放
func (s *WindowCounterTime) Stop() error {
	if atomic.LoadInt64(&s.isRun) == 1 {
		close(s.closeChan) //关闭运行通道
		return nil
	} else {
		return NoRunError
	}
}
