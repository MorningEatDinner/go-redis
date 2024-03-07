package tcp

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/xiaorui/go-redis/interface/tcp"
	"github.com/xiaorui/go-redis/lib/logger"
)

type Config struct {
	Address string
}

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	l, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("start listen...")
	ListenAndServe(l, handler, closeChan)

	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	go func() {
		<-closeChan
		logger.Info("shutting down...")
		listener.Close()
		handler.Close()
	}()

	defer func() {
		listener.Close()
		handler.Close()
	}()
	ctx := context.Background()
	// 等待所有的客户端退出之后才结束程序， 否则你下面break了，可能程序就结束了
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accepted link.. ")
		// 有一个新的请求过来了， 新建一个协程
		// 在每次要建立新的连接时候
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handler(ctx, conn)
		}()
	}
	waitDone.Wait()
}
