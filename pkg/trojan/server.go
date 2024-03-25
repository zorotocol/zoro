package trojan

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/netip"
	"time"
)

type Dialer = func(addr netip.AddrPort) (net.Conn, error)
type Associate = func(addr netip.AddrPort) (net.PacketConn, error)
type Resolver = func(domain string) (netip.Addr, error)
type Authenticator = func(token string) (deadline time.Time, err error)
type Server struct {
	Dialer            Dialer
	Associate         Associate
	Resolver          Resolver
	Authenticator     Authenticator
	ReadHeaderTimeout time.Duration
}

func (s *Server) ServeConn(conn io.ReadWriteCloser) error {
	hdr, err := func() (*header, error) {
		defer time.AfterFunc(s.ReadHeaderTimeout, func() { _ = conn.Close() }).Stop()
		return readHeader(conn, s.Resolver, s.Authenticator)
	}()
	if err != nil {
		return err
	}
	defer time.AfterFunc(time.Until(hdr.Deadline), func() { _ = conn.Close() }).Stop()
	if hdr.IsUDP {
		packetConn, err := s.Associate(hdr.Addr)
		if err != nil {
			return err
		}
		defer packetConn.Close()
		proxyPacket(conn, packetConn)
	} else {
		streamConn, err := s.Dialer(hdr.Addr)
		if err != nil {
			return err
		}
		defer streamConn.Close()
		proxyStream(conn, streamConn)
	}
	return nil
}

func proxyPacket(conn io.ReadWriteCloser, packetConn net.PacketConn) {
	//TODO
}

func proxyStream(conn io.ReadWriteCloser, streamConn net.Conn) {
	go func() {
		defer streamConn.Close()
		pipeConn(streamConn, conn)
	}()
	pipeConn(conn, streamConn)
}
func pipeConn(dst io.Writer, src io.Reader) {
	defer func() {
		if tcpConn, ok := dst.(interface {
			CloseWrite() error
		}); ok {
			_ = tcpConn.CloseWrite()
		}
	}()
	_, _ = io.Copy(dst, src)
}

type header struct {
	Token    string
	IsUDP    bool
	Addr     netip.AddrPort
	Deadline time.Time
}

func readHeader(conn io.Reader, resolver Resolver, authenticator Authenticator) (_ *header, err error) {
	var hdr header
	buff := [264]byte{}
	if _, err = io.ReadFull(conn, buff[:56]); err != nil {
		return nil, err
	}
	hdr.Token = string(buff[:56])
	hdr.Deadline, err = authenticator(hdr.Token)
	if err != nil {
		return nil, err
	}
	if _, err = io.ReadFull(conn, buff[:4]); err != nil {
		return nil, err
	}
	hdr.IsUDP = buff[2] == 3
	switch buff[3] {
	case 1:
		if _, err = io.ReadFull(conn, buff[:4+2+2]); err != nil {
			return nil, err
		}
		hdr.Addr = netip.AddrPortFrom(netip.AddrFrom4([4]byte(buff[:4])), binary.BigEndian.Uint16(buff[4:]))
	case 4:
		if _, err = io.ReadFull(conn, buff[:16+2+2]); err != nil {
			return nil, err
		}
		hdr.Addr = netip.AddrPortFrom(netip.AddrFrom16([16]byte(buff[:16])), binary.BigEndian.Uint16(buff[16:]))
	default:
		if _, err = io.ReadFull(conn, buff[:1]); err != nil {
			return nil, err
		}
		if _, err = io.ReadFull(conn, buff[1:1+buff[0]+2+2]); err != nil {
			return nil, err
		}
		port := binary.BigEndian.Uint16(buff[1+buff[0]:])
		if port == 0 {
			return nil, errors.New("invalid header ports")
		}
		addr, err := resolver(string(buff[1 : 1+buff[0]]))
		if err != nil {
			return nil, err
		}
		hdr.Addr = netip.AddrPortFrom(addr, port)
	}
	return &hdr, nil
}
