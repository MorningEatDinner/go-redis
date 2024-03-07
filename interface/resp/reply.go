package resp

type Reply interface {
	ToBytes() []byte // 把回复的消息转换为字节
}
