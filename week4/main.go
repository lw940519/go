package main

import (
	"ceshi/week4/app"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 监听系统关闭信号
	var waiter = make(chan os.Signal, 1) // 按文档指示，至少设置1的缓冲
	signal.Notify(waiter, syscall.SIGTERM, syscall.SIGINT)

	// 创建grpc
	ua, err := app.InitService()
	if err != nil {
		log.Fatalln(err)
	}

	// 创建一个errgroup
	eg := errgroup.Group{}

	// 运行grpc服务
	eg.Go(func() error {
		select {
		case <-waiter:
			err := ua.Stop()
			if err != nil {
				return err
			}
		}
		return errors.New("收到Signal信号")
	})
	eg.Go(func() error {
		return ua.Start()
	})

	if err := eg.Wait(); err != nil {
		fmt.Println("退出原因" + err.Error())
	}
	fmt.Println("程序被关闭")
}
