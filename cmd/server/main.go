package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"github.com/hashicorp/golang-lru/v2/expirable"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/zorotocol/zoro/pkg/auth"
	"github.com/zorotocol/zoro/pkg/db"
	"github.com/zorotocol/zoro/pkg/gun"
	"github.com/zorotocol/zoro/pkg/misc"
	"github.com/zorotocol/zoro/pkg/multirun"
	"github.com/zorotocol/zoro/pkg/selfcert"
	"github.com/zorotocol/zoro/pkg/trojan"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"time"
)

func main() {
	ln := misc.Must(net.Listen("tcp", os.Getenv("TROJAN")))
	defer ln.Close()
	tlsServer := misc.Must(tls.Listen("tcp", os.Getenv("GRPC"), &tls.Config{
		Certificates: []tls.Certificate{*selfcert.New()},
	}))
	defer tlsServer.Close()
	sqlDB := misc.Must(sql.Open("postgres", os.Getenv("DB")))
	defer sqlDB.Close()
	authenticator := auth.Authenticator{
		DB: &db.DB{
			PG: sqlDB,
		},
		Cache: expirable.NewLRU[string, time.Time](10000, nil, time.Minute),
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
			conn := misc.Must(ln.Accept())
			go trojanServer.ServeConn(conn)

		}
	}, func(context.Context) error {
		grpcServer := grpc.NewServer()
		gun.RegisterGunService(grpcServer, &gun.Gun{
			Handler: func(stream io.ReadWriteCloser) {
				_ = trojanServer.ServeConn(stream)
			},
			ServiceName: "Gun",
		})
		return grpcServer.Serve(tlsServer)
	})
	log.Fatalln(err)
}
