package tcp

import (
	"bufio"
	"context"
	"github.com/xiaorui/go-redis/lib/logger"
	"io"
	"net"
	"sync"
	"time"

	"github.com/xiaorui/go-redis/lib/sync/atomic"
	"github.com/xiaorui/go-redis/lib/sync/wait"
)

// 记录客户端的信息
type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

// Close：关闭客户端连接
func (client *EchoClient) Close() error {
	// 1. 等待客户端任务完成
	client.Waiting.WaitWithTimeout(10 * time.Second) // 等待超过10秒钟之后就关闭
	_ = client.Conn.Close()
	return nil
}

type EchoHandler struct {
	activeConn sync.Map       // 记录有多少个客户端连接
	closing    atomic.Boolean // 记录是否现在正在关闭
}

func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

// Handler: TCP服务器服务客户端连接
func (e *EchoHandler) Handler(ctx context.Context, conn net.Conn) {
	// conn 就是一个客户端连接
	// 1. 如果当下正在关闭， 那么就将所有的客户端进行关闭
	if e.closing.Get() {
		_ = conn.Close()
	}
	// 如果正在运行中，则将当下这个链接转换为一个Client
	client := &EchoClient{
		Conn: conn,
	}
	// 记录当下正在连接着的所有客户
	e.activeConn.Store(client, struct{}{})

	// 处理用户请求
	reader := bufio.NewReader(conn)
	for {
		// 循环接受用户发送过来的消息， 以换行符结束
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("Connection close.")
				e.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}
func (e *EchoHandler) Close() error {
	logger.Info("handler shutting down")
	e.closing.Set(true)

	// 把当下服务器记录的所有客户端服务都干掉
	e.activeConn.Range(func(key, value any) bool {
		client := key.(*EchoClient)
		_ = client.Conn.Close() // 关闭客户端连接
		return true             // 继续对于下一个key执行当下这个操作
	})
	return nil
}
