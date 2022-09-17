package server

import (
	"bytes"
	"encoding/binary"

	"github.com/cocomylove/tcpserver/iface"
)

var defaultHeaderLen uint32 = 8

// 使用小端
type DataPack struct {
}

func NewDataPack() *DataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 {
	return defaultHeaderLen
}
func (dp *DataPack) Pack(msg iface.IMessage) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	if err := binary.Write(buf, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return []byte{}, err
	}
	if err := binary.Write(buf, binary.LittleEndian, msg.GetMsgID()); err != nil {
		return []byte{}, err
	}
	if err := binary.Write(buf, binary.LittleEndian, msg.GetData()); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}
func (dp *DataPack) Unpack(data []byte) (iface.IMessage, error) {
	dataBuf := bytes.NewReader(data)

	msg := msgPool.Get().(*Message)

	if err := binary.Read(dataBuf, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}
	if err := binary.Read(dataBuf, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}

	if err := binary.Read(dataBuf, binary.LittleEndian, &msg.Data); err != nil {
		return nil, err
	}

	return msg, nil
}
