package main

import (
	"context"
	"fmt"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ctx 根context
var ctx context.Context
var ctxCancel func()

func main() {
	// 根上下文和取消
	ctx, ctxCancel = context.WithCancel(context.Background())

	// 监听系统关闭信号
	var waiter = make(chan os.Signal, 1)
	signal.Notify(waiter, syscall.SIGTERM, syscall.SIGINT)

	// 创建一个errgroup
	eg := errgroup.Group{}
	// 创建服务器
	httpService := CreatHttpService(12001)
	httpService2 := CreatHttpService(12002)

	fmt.Printf("%+v", httpService)
	fmt.Printf("%+v", httpService2)

	// 关闭总服务，只需要执行一次
	var once sync.Once
	// 杀掉所有资源
	onceBody := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		// 杀掉所有依赖上下文的请求
		ctxCancel()
		// 关闭服务
		if err := httpService.Close(ctx); err != nil {
			fmt.Println("1" + err.Error())
		}
		if err := httpService2.Close(ctx); err != nil {
			fmt.Println("2" + err.Error())
		}
		fmt.Println("关闭所有服务监听")
	}

	// 赋值路由
	err := httpService.RegisterHandle("/test", HelloWorld)
	if err != nil {
		log.Fatalln(err)
	}
	err = httpService2.RegisterHandle("/test2", HelloWorld)
	if err != nil {
		log.Fatalln(err)
	}

	// 启动2个http监听
	eg.Go(func() error {
		fmt.Println(httpService.Open())
		return nil
	})
	eg.Go(func() error {
		fmt.Println(httpService2.Open())
		return nil
	})

	// 等待系统关闭信号
	eg.Go(func() error {
		select {
		case <-waiter:
			once.Do(onceBody)
		}
		return nil
	})

	// 一个服务退出，所有的服务都注销退出
	eg.Go(func() error {
		select {
		case <-httpService.IsRunning():
			fmt.Println("httpService 未知退出啦")
		case <-httpService2.IsRunning():
			fmt.Println("httpService2 未知退出啦")
		}
		once.Do(onceBody)
		return nil
	})

	err = eg.Wait()
	if err != nil {
		fmt.Printf("等待所有Grountine退出err:%v", err)
		return
	}
	fmt.Println("所有程序关闭退出")
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	eg, ctx := errgroup.WithContext(ctx)

	// 模拟数据操作
	var sum int64 = 0

	eg.Go(func() error {
		io.WriteString(w, "hello world1\n")
		return nil
	})

	// 模拟可能出现的多个耗时的Grountine
	eg.Go(func() error {
		for {
			atomic.AddInt64(&sum, 2)
			select {
			case <-ctx.Done():
				fmt.Println("HelloWorld方法1" + ctx.Err().Error())
				return nil
			default:
				time.Sleep(5 * time.Second)
			}
		}
	})
	eg.Go(func() error {
		for {
			atomic.AddInt64(&sum, -1)
			select {
			case <-ctx.Done():
				fmt.Println("HelloWorld方法2" + ctx.Err().Error())
				return nil
			default:
				time.Sleep(5 * time.Second)
			}
		}
	})

	err := eg.Wait()
	if err != nil {
		return
	}
	fmt.Println(atomic.LoadInt64(&sum))
	fmt.Println("HelloWorld方法退出")
}

// http的Service 服务
type HTTPService interface {
	Open() error
	Close(ctx context.Context) error
	RegisterHandle(path string, handle func(http.ResponseWriter, *http.Request)) error
	IsRunning() chan struct{}
}

type httpServer struct {
	Port       int
	Server     *http.Server
	Routerlist *http.ServeMux
	IsClose    chan struct{}
}

// Open 阻塞开启，应该使用goruntion来开启
func (h *httpServer) Open() error {
	err := fmt.Errorf("端口号:%d 服务：%w", h.Port, h.Server.ListenAndServe())
	close(h.IsClose)
	return err
}

// 关闭
func (h *httpServer) Close(ctx context.Context) error {
	err := h.Server.Shutdown(ctx)
	return err
}

// 添加路由
func (h *httpServer) RegisterHandle(path string, handle func(http.ResponseWriter, *http.Request)) error {
	h.Routerlist.HandleFunc(path, handle)
	return nil
}

// 判断是否关闭
func (h *httpServer) IsRunning() chan struct{} {
	return h.IsClose
}

//  创建一个HTTPService
func CreatHttpService(port int) HTTPService {
	server := &httpServer{
		port,
		&http.Server{Addr: ":" + cast.ToString(port)},
		http.NewServeMux(),
		make(chan struct{}),
	}
	server.Server.Handler = server.Routerlist
	return server
}
