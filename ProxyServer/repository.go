package ProxyServer

import (
	"Proxy/Repeater"
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

func (srv *Server) saveRequest(r *http.Request) (int, error) {
	var reqId int
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

func (srv *Server) saveResponse(reqId int, resp *http.Response) (*Repeater.Response, error) {
	var respId int
	var addTime time.Time
	respHeaders := headersToString(resp.Header)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))

	err = srv.DB.QueryRow(`INSERT INTO response (req_id, code, resp_message, header, body)
	values ($1, $2, $3, $4, $5) RETURNING id, add_time`,
		reqId,
		resp.StatusCode,
		resp.Status[4:],
		respHeaders,
		respBody).Scan(&respId, &addTime)

	response := &Repeater.Response{
		Id:          respId,
		ReqId:       reqId,
		Code:        resp.StatusCode,
		RespMessage: resp.Status[4:],
		Header:      Repeater.StrToHeader(respHeaders),
		Body:        string(respBody),
		AddTime:     addTime,
	}

	return response, err
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
