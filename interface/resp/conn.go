package resp

// Connection: Redis协议层的一个连接
type Connection interface {
	Write([]byte) error //  给客户端回复消息
	GetDBIndex() int
	SelectDB(int) // 切换DB
}
