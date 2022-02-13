package ProxyServer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
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

func (srv *Server) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc,
	}
	return c
}

type conn struct {
	server *Server
	rwc    net.Conn
	//remoteAddr string
}

func (c *conn) serve() {
	defer c.rwc.Close()

	if d := c.server.ReadTimeout; d > 0 {
		c.rwc.SetReadDeadline(time.Now().Add(d))
	}
	if d := c.server.WriteTimeout; d > 0 {
		c.rwc.SetWriteDeadline(time.Now().Add(d))
	}

	buf := bufio.NewReader(c.rwc)

	req, err := http.ReadRequest(buf)
	if err != nil {
		log.Fatal(err)
	}

	req.RequestURI = ""
	req.Header.Del("Proxy-Connection")

	resp, err := proxyRequest(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	respByte, err := ReadResp(resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.rwc.Write(respByte)
}

func proxyRequest(req *http.Request) (*http.Response, error) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return resp, err
	}

	return resp, nil
}

func ReadResp(resp *http.Response) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(resp.Proto)
	b.WriteString(" ")
	b.WriteString(resp.Status)
	b.WriteString("\r\n")

	for h, v := range resp.Header {
		b.WriteString(fmt.Sprintf("%s:", h))
		for _, i := range v {
			b.WriteString(fmt.Sprintf(" %v", i))
		}
		b.WriteString("\r\n")
	}

	b.WriteString("\r\n")

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return b.Bytes(), err
	}

	b.Write(respBody)
	b.WriteString("\r\n")

	return b.Bytes(), err
}

func PrintReq(req *http.Request) {
	fmt.Println(req.Method, req.RequestURI, req.Proto)

	fmt.Println("Host:", req.Host)

	for h, v := range req.Header {
		fmt.Printf("%s:", h)
		for _, i := range v {
			fmt.Printf(" %v", i)
		}
		fmt.Println()
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(reqBody))
}

func PrintResp(resp *http.Response) {
	fmt.Println(resp.Proto, resp.Status)

	for h, v := range resp.Header {
		fmt.Printf("%s:", h)
		for _, i := range v {
			fmt.Printf(" %v", i)
		}
		fmt.Println()
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(respBody))
}
