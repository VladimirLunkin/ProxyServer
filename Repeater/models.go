package Repeater

import "time"

type Request struct {
	Id      int       `json:"id"`
	Method  string    `json:"method"`
	Scheme  string    `json:"scheme"`
	Host    string    `json:"host"`
	Path    string    `json:"path"`
	Header  string    `json:"header,omitempty"`
	Body    string    `json:"body,omitempty"`
	AddTime time.Time `json:"add_time"`
}

type Response struct {
	Id          int       `json:"id"`
	ReqId       int       `json:"req_id"`
	Code        int       `json:"code"`
	RespMessage string    `json:"resp_message"`
	Header      string    `json:"header"`
	Body        string    `json:"body"`
	AddTime     time.Time `json:"add_time"`
}

type Repository interface {
	GetRequests() ([]Request, error)
	GetRequest(id int) (Request, error)
}
