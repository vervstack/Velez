// Command socksproxy is a minimal SOCKS5 server (RFC 1928: no authentication,
// CONNECT only) the e2e harness runs inside a DinD container, so the test
// process reaches DinD-internal addresses through one published port.
package main

import (
	"context"
	"encoding/binary"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"slices"
	"strconv"
	"time"

	"go.redsock.ru/rerrors"
)

const (
	socksVersion  = 5
	methodNoAuth  = 0
	methodNoMatch = 0xff
	cmdConnect    = 1

	atypIpv4   = 1
	atypDomain = 3
	atypIpv6   = 4

	replySucceeded       = 0
	replyGeneralFailure  = 1
	replyHostUnreachable = 4
	replyCmdNotSupported = 7
	replyAtypNotSupport  = 8

	handshakeTimeout = 10 * time.Second
	dialTimeout      = 10 * time.Second
	ipv4Len          = 4
	ipv6Len          = 16
	replyHeaderLen   = 4
	greetingLen      = 2
	portLen          = 2
	pipeDirections   = 2
)

var (
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errBadVersion = rerrors.New("unsupported socks version")
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errNoAuthMethod = rerrors.New("no acceptable auth method")
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errBadCommand = rerrors.New("unsupported command")
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errBadAddressType = rerrors.New("unsupported address type")
)

func main() {
	listen := flag.String("listen", ":1080", "address to listen on")

	flag.Parse()

	listenConfig := &net.ListenConfig{}

	listener, err := listenConfig.Listen(context.Background(), "tcp", *listen)
	if err != nil {
		log.Printf("error listening on %s: %v", *listen, err)
		os.Exit(1)
	}

	for {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			log.Printf("error accepting connection: %v", acceptErr)

			continue
		}

		go serve(conn)
	}
}

func serve(client net.Conn) {
	defer func() { _ = client.Close() }()

	_ = client.SetDeadline(time.Now().Add(handshakeTimeout))

	err := negotiate(client)
	if err != nil {
		return
	}

	target, replyCode, err := readConnectRequest(client)
	if err != nil {
		_ = writeReply(client, replyCode, nil)

		return
	}

	dialer := &net.Dialer{Timeout: dialTimeout}

	upstream, err := dialer.DialContext(context.Background(), "tcp", target)
	if err != nil {
		_ = writeReply(client, replyHostUnreachable, nil)

		return
	}

	defer func() { _ = upstream.Close() }()

	err = writeReply(client, replySucceeded, upstream.LocalAddr())
	if err != nil {
		return
	}

	_ = client.SetDeadline(time.Time{})

	pipe(client, upstream)
}

func negotiate(client net.Conn) error {
	header := make([]byte, greetingLen)

	_, err := io.ReadFull(client, header)
	if err != nil {
		return rerrors.Wrap(err, "read greeting")
	}

	if header[0] != socksVersion {
		return rerrors.Wrap(errBadVersion, strconv.Itoa(int(header[0])))
	}

	methods := make([]byte, header[1])

	_, err = io.ReadFull(client, methods)
	if err != nil {
		return rerrors.Wrap(err, "read methods")
	}

	if slices.Contains(methods, methodNoAuth) {
		_, err = client.Write([]byte{socksVersion, methodNoAuth})
		if err != nil {
			return rerrors.Wrap(err, "write method choice")
		}

		return nil
	}

	_, _ = client.Write([]byte{socksVersion, methodNoMatch})

	return errNoAuthMethod
}

func readConnectRequest(client net.Conn) (string, byte, error) {
	header := make([]byte, replyHeaderLen)

	_, err := io.ReadFull(client, header)
	if err != nil {
		return "", replyGeneralFailure, rerrors.Wrap(err, "read request")
	}

	if header[0] != socksVersion {
		return "", replyGeneralFailure, rerrors.Wrap(errBadVersion, strconv.Itoa(int(header[0])))
	}

	if header[1] != cmdConnect {
		return "", replyCmdNotSupported, rerrors.Wrap(errBadCommand, strconv.Itoa(int(header[1])))
	}

	host, replyCode, err := readHost(client, header[3])
	if err != nil {
		return "", replyCode, err
	}

	portBytes := make([]byte, portLen)

	_, err = io.ReadFull(client, portBytes)
	if err != nil {
		return "", replyGeneralFailure, rerrors.Wrap(err, "read port")
	}

	port := strconv.Itoa(int(binary.BigEndian.Uint16(portBytes)))

	return net.JoinHostPort(host, port), replySucceeded, nil
}

func readHost(client net.Conn, atyp byte) (string, byte, error) {
	switch atyp {
	case atypIpv4:
		raw := make([]byte, ipv4Len)

		_, err := io.ReadFull(client, raw)
		if err != nil {
			return "", replyGeneralFailure, rerrors.Wrap(err, "read ipv4")
		}

		return net.IP(raw).String(), replySucceeded, nil
	case atypIpv6:
		raw := make([]byte, ipv6Len)

		_, err := io.ReadFull(client, raw)
		if err != nil {
			return "", replyGeneralFailure, rerrors.Wrap(err, "read ipv6")
		}

		return net.IP(raw).String(), replySucceeded, nil
	case atypDomain:
		length := make([]byte, 1)

		_, err := io.ReadFull(client, length)
		if err != nil {
			return "", replyGeneralFailure, rerrors.Wrap(err, "read domain length")
		}

		raw := make([]byte, length[0])

		_, err = io.ReadFull(client, raw)
		if err != nil {
			return "", replyGeneralFailure, rerrors.Wrap(err, "read domain")
		}

		return string(raw), replySucceeded, nil
	default:
		return "", replyAtypNotSupport, rerrors.Wrap(errBadAddressType, strconv.Itoa(int(atyp)))
	}
}

func writeReply(client net.Conn, code byte, bound net.Addr) error {
	reply := []byte{socksVersion, code, 0, atypIpv4, 0, 0, 0, 0, 0, 0}

	tcpAddr, ok := bound.(*net.TCPAddr)
	if ok && tcpAddr.IP.To4() != nil {
		copy(reply[4:8], tcpAddr.IP.To4())
		binary.BigEndian.PutUint16(reply[8:10], uint16(tcpAddr.Port))
	}

	_, err := client.Write(reply)
	if err != nil {
		return rerrors.Wrap(err, "write reply")
	}

	return nil
}

func pipe(client, upstream net.Conn) {
	done := make(chan struct{}, pipeDirections)

	go func() {
		_, _ = io.Copy(upstream, client)
		closeWrite(upstream)

		done <- struct{}{}
	}()

	go func() {
		_, _ = io.Copy(client, upstream)
		closeWrite(client)

		done <- struct{}{}
	}()

	<-done
	<-done
}

func closeWrite(conn net.Conn) {
	tcpConn, ok := conn.(*net.TCPConn)
	if ok {
		_ = tcpConn.CloseWrite()
	}
}
