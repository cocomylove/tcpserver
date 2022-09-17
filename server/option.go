package server

import "github.com/cocomylove/tcpserver/iface"

type Option func(s *Server)

func WithPacket(pack iface.IDataPack) Option {
	return func(s *Server) {
		s.packet = pack
	}
}
