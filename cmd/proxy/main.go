package main

import (
	"context"
	"errors"
	"github.com/hashicorp/golang-lru/v2/expirable"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/zorotocol/zoro/pkg/auth"
	"github.com/zorotocol/zoro/pkg/gun"
	"github.com/zorotocol/zoro/pkg/misc"
	"github.com/zorotocol/zoro/pkg/multirun"
	"github.com/zorotocol/zoro/pkg/selfcert"
	"github.com/zorotocol/zoro/pkg/trojan"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"time"
)

func main() {
	httpServer := misc.Must(net.Listen("tcp", os.Getenv("TROJAN")))
	defer httpServer.Close()
	httpsServer := misc.Must(net.Listen("tcp", os.Getenv("GRPC")))
	defer httpsServer.Close()
	authenticator := auth.Client{
		HTTPClient: http.DefaultClient,
		URL:        os.Getenv("ORACLE"),
		Cache:      expirable.NewLRU[string, time.Time](misc.Must(strconv.Atoi(os.Getenv("CACHE"))), nil, time.Minute),
	}
	trojanServer := trojan.Server{
		Dialer: func(addr netip.AddrPort) (net.Conn, error) {
			if addr.Addr().IsLoopback() || addr.Addr().IsPrivate() {
				return nil, errors.New("invalid addr")
			}
			return net.DialTimeout("tcp", addr.String(), time.Second*3)
		},
		Associate: func(addr netip.AddrPort) (net.PacketConn, error) {
			return nil, errors.New("no udp support")
		},
		Resolver: func(domain string) (netip.Addr, error) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", domain)
			if err != nil {
				return netip.Addr{}, err
			}
			if len(addrs) == 0 {
				return netip.Addr{}, errors.New("no such host")
			}
			return addrs[0], err
		},
		Authenticator: func(token string) (deadline time.Time, err error) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			return authenticator.Authenticate(ctx, token)
		},
		ReadHeaderTimeout: time.Second * 3,
	}
	err := multirun.Run(context.Background(), func(context.Context) error {
		for {
			conn, err := httpServer.Accept()
			if err != nil {
				return err
			}
			go trojanServer.ServeConn(conn)
		}
	}, func(context.Context) error {
		grpcServer := grpc.NewServer(grpc.Creds(credentials.NewServerTLSFromCert(selfcert.New())))
		gun.RegisterGunService(grpcServer, &gun.Gun{
			Handler: func(stream io.ReadWriteCloser) {
				_ = trojanServer.ServeConn(stream)
			},
			ServiceName: "G",
		})
		return grpcServer.Serve(httpsServer)
	})
	log.Fatalln(err)
}
