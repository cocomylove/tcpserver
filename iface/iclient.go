package iface

// IClient 客户接入
type IClient interface {
	Connect() error
	Send(data []byte) error
	ReadMessage() (<-chan []byte,error)
}