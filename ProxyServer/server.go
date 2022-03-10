package ProxyServer

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DB           *pgx.ConnPool
}

func (srv *Server) ListenAndServe() error {
	server := http.Server{
		Addr: srv.Addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("======")
			fmt.Println(r)
			fmt.Println("======")
			srv.proxy(w, r)
		}),
	}

	return server.ListenAndServe()
}

func (srv *Server) proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		srv.ProxyHTTPS(w, r)
	} else {
		srv.ProxyHTTP(w, r)
	}
}

func (srv *Server) ProxyHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Proxy-Connection")

	reqId, err := srv.saveRequest(r)
	if err != nil {
		log.Printf("fail save to db: %v", err)
	}

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	err = srv.saveResponse(reqId, resp)
	if err != nil {
		log.Printf("fail save to db: %v", err)
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (srv *Server) ProxyHTTPS(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	localConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("hijacking error: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	_, err = localConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		log.Printf("handshaking failed: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		localConn.Close()
		return
	}
	defer localConn.Close()

	host := strings.Split(r.Host, ":")[0]

	tlsConfig, err := generateTLSConfig(host, r.URL.Scheme)
	if err != nil {
		log.Printf("error getting cert: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tlsLocalConn := tls.Server(localConn, &tlsConfig)
	err = tlsLocalConn.Handshake()
	if err != nil {
		tlsLocalConn.Close()
		log.Printf("handshaking failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tlsLocalConn.Close()

	remoteConn, err := tls.Dial("tcp", r.URL.Host, &tlsConfig)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer remoteConn.Close()

	reader := bufio.NewReader(tlsLocalConn)
	request, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("error getting request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	requestByte, err := httputil.DumpRequest(request, true)
	if err != nil {
		log.Printf("failed to dump request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = remoteConn.Write(requestByte)
	if err != nil {
		log.Printf("failed to write request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	serverReader := bufio.NewReader(remoteConn)
	response, err := http.ReadResponse(serverReader, request)
	if err != nil {
		log.Printf("failed to read response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rawResponse, err := httputil.DumpResponse(response, true)
	if err != nil {
		log.Printf("failed to dump response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tlsLocalConn.Write(rawResponse)
	if err != nil {
		log.Printf("fail to write response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request.URL.Scheme = "https"
	hostAndPort := strings.Split(r.URL.Host, ":")
	request.URL.Host = hostAndPort[0]

	reqId, err := srv.saveRequest(request)
	if err != nil {
		log.Printf("fail save to db: %v", err)
	}

	err = srv.saveResponse(reqId, response)
	if err != nil {
		log.Printf("fail save to db: %v", err)
	}
}

func (srv *Server) saveRequest(r *http.Request) (int32, error) {
	var reqId int32
	reqHeaders := headersToString(r.Header)
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	err = srv.DB.QueryRow(`INSERT INTO request (method, scheme, host, path, header, body)
			values ($1, $2, $3, $4, $5, $6) RETURNING id`,
		r.Method,
		r.URL.Scheme,
		r.URL.Host,
		r.URL.Path,
		reqHeaders,
		string(reqBody)).Scan(&reqId)

	return reqId, err
}

func (srv *Server) saveResponse(reqId int32, resp *http.Response) error {
	var respId int32
	respHeaders := headersToString(resp.Header)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))

	return srv.DB.QueryRow(`INSERT INTO response (req_id, code, resp_message, header, body)
	values ($1, $2, $3, $4, $5) RETURNING id`,
		reqId,
		resp.StatusCode,
		resp.Status[4:],
		respHeaders,
		respBody).Scan(&respId)
}

func headersToString(headers http.Header) string {
	var stringHeaders string
	for key, values := range headers {
		for _, value := range values {
			stringHeaders += key + " " + value + "\n"
		}
	}
	return stringHeaders
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
