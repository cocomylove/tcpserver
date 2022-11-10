package client

import (
)


type ServerMessage interface {
     GetData() []byte
     GetType() uint32
}

type message struct {
    Data []byte
    Type uint32
}

func (m *message)GetData() []byte  {
    return m.Data
}

func (m *message)GetType() uint32  {
    return m.Type
}
