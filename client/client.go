package client

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type Config struct {
	Host    string
	Timeout time.Duration
	// 重连次数
	RetryTimes int
	// 消息类型 1 text 2 Binary
	MessageType int
}

type WSClient struct {
	config Config
	conn   *websocket.Conn
	ctx    context.Context
	retry  int
}

func (c *WSClient) Connect() error {
	ws, resp, err := websocket.DefaultDialer.Dial(c.config.Host, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return errors.New("response code is not 101")
	}
	c.conn = ws
	_ = c.conn.SetWriteDeadline(time.Now().Add(c.config.Timeout))
	return nil
}

func (c *WSClient) Send(data []byte) error {
	return c.conn.WriteMessage(c.config.MessageType, data)
}

func (c *WSClient) ReadMessage() (<-chan []byte, error) {
	message := make(chan []byte, 1)

	return message, nil
}

func (c *WSClient) reader(message chan []byte) {
Loop:
	for {
		t, data, err := c.conn.ReadMessage()
		if t != websocket.BinaryMessage {
			continue
		}
		if err != nil {
			if c.retry < c.config.RetryTimes {
				goto Loop
			}
			return
		}
		message <- data
	}

}
