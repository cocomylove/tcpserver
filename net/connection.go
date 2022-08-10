package net

import (
	"context"
	"net"
	"sync"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/utils/config"
)

type Connection struct {
	//当前Conn属于哪个Server
	TCPServer iface.IServer
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//消息管理MsgID和对应处理方法的消息管理模块
	MsgHandler iface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	sync.RWMutex
	//链接属性
	property map[string]interface{}
	////保护当前property的锁
	propertyLock sync.Mutex
	//当前连接的关闭状态
	isClosed bool
}

func NewConnection(server iface.IServer, conn *net.TCPConn, connID uint32, msgHandler iface.IMsgHandle) *Connection {
	c := &Connection{
		TCPServer:   server,
		Conn:        conn,
		ConnID:      connID,
		MsgHandler:  msgHandler,
		msgBuffChan: make(chan []byte, config.GlobalObj.MaxMsgChanLen),
	}
	c.TCPServer.GetConnMgr().Add(c)
	return c
}

func (c *Connection) Start() {

}
func (c *Connection) Stop() {

}
func (c *Connection) Context() context.Context {
	return c.ctx
}
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
func (c *Connection) SendMsg(msgID uint32, data []byte) error {
	return nil
}
func (c *Connection) SendBuffMsg(msgID uint32, data []byte) error {
	return nil
}
func (c *Connection) SetProperty(key string, value interface{}) {

}
func (c *Connection) GetProperty(key string) (interface{}, error) {

	return nil, nil
}
func (c *Connection) RemoveProperty(key string) {

}
