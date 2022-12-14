package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"
)

func TestServer(t *testing.T) {

	config.InitGlobal()
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	//创建一个server句柄
	s := NewWSServer(log)

	//注册链接hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	// 多路由
	s.AddRouter(1, &PingRouter{})
	s.AddRouter(2, &HelloRouter{})

	//2 开启服务
	s.Serve()
	//	客户端测试
	// go ClientTest(1)
	// go ClientTest(2)

	select {
	case <-time.After(time.Second * 120):
		return
	}
}

func ClientTest(i uint32) {
	log.Println("client starting")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client start err, exit!", err)
		return
	}
	j := 0
	for {

		dp := NewDataPack()
		msg, _ := dp.Pack(NewMessage(i, []byte("client test message+"+strconv.Itoa(j))))
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Println("client write err: ", err)
			return
		}

		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData)
		if err != nil {
			fmt.Println("client read head err: ", err)
			return
		}

		// 将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("client unpack head err: ", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("client unpack data err")
				return
			}

			fmt.Printf("==> Client receive Msg: ID = %d, len = %d , data = %s\n", msg.ID, msg.DataLen, msg.Data)
		}

		time.Sleep(1 * time.Millisecond)
		j++
	}
}

//ping test 自定义路由
type PingRouter struct {
	BaseRouter
}

//Test PreHandle
func (this *PingRouter) PreHandle(request iface.IRequest) {
	fmt.Println("Call pingRouter PreHandle")
	err := request.GetConnection().SendBuffMsg(1, []byte("before ping ....\n"))
	if err != nil {
		fmt.Println("preHandle SendMsg err: ", err)
	}
}

//Test Handle
func (this *PingRouter) Handle(request iface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgID=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendBuffMsg(1, []byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("Handle SendMsg err: ", err)
	}
}

//Test PostHandle
func (this *PingRouter) PostHandle(request iface.IRequest) {
	fmt.Println("Call Router PostHandle")
	err := request.GetConnection().SendBuffMsg(1, []byte("After ping .....\n"))
	if err != nil {
		fmt.Println("Post SendMsg err: ", err)
	}
}

type HelloRouter struct {
	BaseRouter
}

func (this *HelloRouter) Handle(request iface.IRequest) {
	fmt.Println("call helloRouter Handle")
	fmt.Printf("receive from client msgID=%d, data=%s\n", request.GetMsgID(), string(request.GetData()))

	err := request.GetConnection().SendBuffMsg(2, []byte("hello zix hello Router"))
	if err != nil {
		fmt.Println(err)
	}
}

func DoConnectionBegin(conn iface.IConnection) {
	fmt.Println("DoConnectionBegin is Called ... ")
	err := conn.SendMsg(2, []byte("DoConnection BEGIN..."))
	if err != nil {
		fmt.Println(err)
	}
}

//连接断开的时候执行
func DoConnectionLost(conn iface.IConnection) {
	fmt.Println("DoConnectionLost is Called ... ")
}

func TestClient(t *testing.T) {
	go ClientTest(1)
	go ClientTest(2)
	go ClientTest(1)
	go ClientTest(1)
	select {
	case <-time.After(time.Second * 60):
		return
	}
}
