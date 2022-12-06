package iface

// ServerMessage 来自服务器的消息
type ServerMessage interface {
	GetData() []byte
	GetType() uint32
}
