package Repeater

import (
	"Proxy/Repeater/delivery"
	"bytes"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

type Handlers struct {
	Repo  Repository
	Proxy ProxyServer
}

func SetRepeaterRouting(r *mux.Router, h *Handlers) {
	r.HandleFunc("/requests", h.GetRequests).Methods("GET")
	r.HandleFunc("/requests/{id:[0-9]+}", h.GetRequest).Methods("GET")
	r.HandleFunc("/repeat/{id:[0-9]+}", h.RepeatRequest).Methods("GET")
	r.HandleFunc("/scan/{id:[0-9]+}", h.VulnerabilityScan).Methods("GET")
}

func (h *Handlers) GetRequests(w http.ResponseWriter, r *http.Request) {
	req, err := h.Repo.GetRequests()
	if err != nil {
		delivery.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	delivery.Send(w, http.StatusOK, req)
}

func (h *Handlers) GetRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		delivery.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	req, err := h.Repo.GetRequest(id)
	if err != nil {
		delivery.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	delivery.Send(w, http.StatusOK, req)
}

func (h *Handlers) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		delivery.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	reqDB, err := h.Repo.GetRequest(id)
	if err != nil {
		delivery.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	body := bytes.NewBufferString(reqDB.Body)

	urlStr := reqDB.Scheme + "://" + reqDB.Host + reqDB.Path
	req, err := http.NewRequest(reqDB.Method, urlStr, body)
	if err != nil {
		delivery.SendError(w, http.StatusOK, err.Error())
		return
	}

	for key, value := range reqDB.Header {
		req.Header.Add(key, value)
	}

	resp := h.Proxy.ProxyHTTP(req)

	delivery.Send(w, http.StatusOK, resp)
}

var (
	xxe                 = "<!DOCTYPE foo [\n  <!ELEMENT foo ANY >\n  <!ENTITY xxe SYSTEM \"file:///etc/passwd\" >]>\n<foo>&xxe;</foo>\n"
	xml                 = "<?xml"
	target              = "root:"
	requestIsVulnerable = "request is vulnerable!!!😈"
)

func (h *Handlers) VulnerabilityScan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		delivery.SendError(w, http.StatusNotFound, "")
		return
	}

	reqDB, err := h.Repo.GetRequest(id)
	if err != nil {
		delivery.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	if ind := strings.Index(reqDB.Body, xml); ind != -1 {
		reqDB.Body = reqDB.Body[:ind] + xxe
	}
	body := bytes.NewBufferString(reqDB.Body)

	urlStr := reqDB.Scheme + "://" + reqDB.Host + reqDB.Path
	req, err := http.NewRequest(reqDB.Method, urlStr, body)
	if err != nil {
		delivery.SendError(w, http.StatusOK, err.Error())
		return
	}

	for key, value := range reqDB.Header {
		req.Header.Add(key, value)
	}

	resp := h.Proxy.ProxyHTTP(req)

	if ind := strings.Index(resp.Body, target); ind != -1 {
		delivery.Send(w, http.StatusOK, requestIsVulnerable)
		return
	}

	delivery.Send(w, http.StatusOK, resp)
}
