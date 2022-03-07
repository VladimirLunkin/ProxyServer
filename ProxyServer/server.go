package ProxyServer

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"strconv"
	"time"
)

type Server struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (srv *Server) Serve(l net.Listener) error {
	for {
		rwc, err := l.Accept()
		if err != nil {
			return err
		}

		conn := srv.newConn(rwc)
		go conn.serve()
	}
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

type conn struct {
	server *Server
	rwc    net.Conn
}

func (srv *Server) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc,
	}
	return c
}

func (c *conn) serve() {
	defer c.rwc.Close()

	if d := c.server.ReadTimeout; d > 0 {
		c.rwc.SetReadDeadline(time.Now().Add(d))
	}
	if d := c.server.WriteTimeout; d > 0 {
		c.rwc.SetWriteDeadline(time.Now().Add(d))
	}

	reader := bufio.NewReader(c.rwc)
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println(err)
		return
	}

	proxyConn, err := Dial(c.rwc, req)
	if err != nil {
		log.Println(err)
		return
	}
	defer proxyConn.Close()

	resp, err := HandleProxy(proxyConn, req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	respByte, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = c.rwc.Write(respByte)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Request:", req.Method, req.URL, "Response:", resp.Status)
}

func Dial(clientConn net.Conn, r *http.Request) (net.Conn, error) {
	if r.Method == http.MethodConnect {
		_, err := clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			return nil, err
		}

		tlsConfig, err := generateTLSConfig(r.Host, r.URL.Scheme)
		if err != nil {
			return nil, err
		}

		tlsLocalConn := tls.Server(clientConn, &tlsConfig)
		err = tlsLocalConn.Handshake()
		if err != nil {
			return nil, err
		}

		remoteConn, err := tls.Dial("tcp", r.URL.Host, &tlsConfig)
		if err != nil {
			return nil, err
		}

		return remoteConn, nil
	}

	remoteConn, err := net.Dial("tcp", r.URL.Host + ":80")
	if err != nil {
		return nil, err
	}
	return remoteConn, nil
}

func generateTLSConfig(host, URL string) (tls.Config, error) {
	cmd := exec.Command("/bin/sh", "./scripts/gen_cert.sh", host, strconv.Itoa(rand.Intn(math.MaxInt32)))

	err := cmd.Start()
	if err != nil {
		return tls.Config{}, errors.New(fmt.Sprintf("Start create cert file script error: %v\n", err))
	}

	err = cmd.Wait()
	if err != nil {
		return tls.Config{}, errors.New(fmt.Sprintf("Wait create cert file script error: %v\n", err))
	}

	tlsCert, err := tls.LoadX509KeyPair("certs/"+host+".crt", "cert.key")
	if err != nil {
		log.Println("error loading pair", err)
		return tls.Config{}, err
	}

	return tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ServerName:   URL,
	}, nil
}

func HandleProxy(c net.Conn, req *http.Request) (*http.Response, error) {
	err := req.WriteProxy(c)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(c)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
