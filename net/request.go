package net

import "github.com/cocomylove/tcpserver/iface"

type Requset struct {
	conn iface.IConnection //已经和客户端建立好的 链接
	msg  iface.IMessage    //客户端请求的数据
}

func (r *Requset) GetConnection() iface.IConnection {
	return r.conn
}
func (r *Requset) GetData() []byte {
	return r.msg.GetData()
}
func (r *Requset) GetMsgID() uint32 {
	return r.msg.GetMsgID()
}
