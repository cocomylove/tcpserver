package net

import (
	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"
)

type MessageHandler struct {
	Apis       map[uint32]iface.IRouter
	WorkerPool uint32
	TaskQueue  []chan iface.IRequest
	logger     *zap.Logger
}

func NewMessageHandler(logger *zap.Logger) *MessageHandler {
	mh := &MessageHandler{
		Apis:       make(map[uint32]iface.IRouter),
		WorkerPool: config.GlobalObj.WorkerPoolSize,
		TaskQueue:  make([]chan iface.IRequest, config.GlobalObj.WorkerPoolSize, config.GlobalObj.WorkerPoolSize),
		logger:     logger,
	}
	return mh
}

func (m *MessageHandler) DoMsgHandler(request iface.IRequest) {
	handler, ok := m.Apis[request.GetMsgID()]
	if !ok {
		m.logger.Warn("api is not FOUND ", zap.Uint32("msgID", request.GetMsgID()))
		return
	}
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}
func (m *MessageHandler) AddRouter(msgID uint32, router iface.IRouter) {
	m.Apis[msgID] = router
}
func (m *MessageHandler) StartWorkerPool() {
	for i := 0; i < int(m.WorkerPool); i++ {
		m.TaskQueue[i] = make(chan iface.IRequest, config.GlobalObj.MaxWorkerTaskLen)
		go m.StartOneWorker(i, m.TaskQueue[i])
	}
}
func (m *MessageHandler) SendMsgToTaskQueue(request iface.IRequest) {
	id := request.GetConnection().GetConnID() % m.WorkerPool

	m.TaskQueue[id] <- request
}

func (m *MessageHandler) StartOneWorker(workerID int, taskQueue chan iface.IRequest) {
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			m.DoMsgHandler(request)
		}
	}
}
