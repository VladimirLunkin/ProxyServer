package Repeater

import (
	"net/http"
	"strings"
	"time"
)

func StrToHeader(headers string) map[string]string {
	h := make(map[string]string)
	for _, header := range strings.Split(headers, "\n") {
		if len(header) < 2 {
			continue
		}
		str := strings.SplitN(header, " ", 2)
		h[str[0]] = str[1]
	}
	return h
}

type Request struct {
	Id      int               `json:"id"`
	Method  string            `json:"method"`
	Scheme  string            `json:"scheme"`
	Host    string            `json:"host"`
	Path    string            `json:"path"`
	Header  map[string]string `json:"header,omitempty"`
	Body    string            `json:"body,omitempty"`
	AddTime time.Time         `json:"add_time"`
}

type Response struct {
	Id          int               `json:"id"`
	ReqId       int               `json:"req_id"`
	Code        int               `json:"code"`
	RespMessage string            `json:"resp_message"`
	Header      map[string]string `json:"header"`
	Body        string            `json:"body"`
	AddTime     time.Time         `json:"add_time"`
}

type Repository interface {
	GetRequests() ([]Request, error)
	GetRequest(id int) (Request, error)
}

type ProxyServer interface {
	ProxyHTTP(r *http.Request) *Response
}
