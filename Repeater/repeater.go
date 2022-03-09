package Repeater

import (
	"Proxy/Repeater/delivery"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
)

type Handlers struct {
	Repo Repository
}

func SetRepeaterRouting(r *router.Router, h *Handlers) {
	r.GET("/requests", h.GetRequests)
	r.GET("/requests/{id}", h.GetRequest)
	r.GET("/repeat/{id}", h.RepeatRequest)
	r.GET("/scan/{id}", h.VulnerabilityScan)
}

func (h *Handlers) GetRequests(ctx *fasthttp.RequestCtx) {
	req, err := h.Repo.GetRequests()
	if err != nil {
		delivery.SendError(ctx, http.StatusNotFound, err.Error())
		return
	}

	delivery.Send(ctx, http.StatusOK, req)
}

func (h *Handlers) GetRequest(ctx *fasthttp.RequestCtx) {
	id, err := strconv.Atoi(fmt.Sprintf("%s", ctx.UserValue("id")))
	if err != nil {
		delivery.SendError(ctx, http.StatusNotFound, "")
		return
	}

	req, err := h.Repo.GetRequest(id)
	if err != nil {
		delivery.SendError(ctx, http.StatusNotFound, err.Error())
		return
	}

	delivery.Send(ctx, http.StatusOK, req)
}

func (h *Handlers) RepeatRequest(ctx *fasthttp.RequestCtx) {

}

func (h *Handlers) VulnerabilityScan(ctx *fasthttp.RequestCtx) {

}
