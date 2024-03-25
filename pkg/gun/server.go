package gun

//go:generate protoc --go_out=. --go-grpc_out=. ./grpc.proto

import (
	"context"
	"github.com/zorotocol/zoro/pkg/proto"
	"google.golang.org/grpc"
	"io"
)

type impl struct {
	gun *Gun
	proto.UnimplementedGunServer
}

func (i impl) Tun(conn proto.Gun_TunServer) error {
	c := newConn(conn)
	go i.gun.Handler(c)
	<-c.ctx.Done()
	return context.Cause(c.ctx)
}

type Gun struct {
	Handler     func(closer io.ReadWriteCloser)
	ServiceName string
}

func RegisterGunService(s grpc.ServiceRegistrar, gunService *Gun) {
	desc := proto.Gun_ServiceDesc
	desc.ServiceName = gunService.ServiceName
	s.RegisterService(&desc, impl{gun: gunService})
}
