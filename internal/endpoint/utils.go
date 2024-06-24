package endpoint

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

type response struct {
	StatusCode int    `json:"statusCode"`
	Data       []byte `json:"data"`
}

type errorResponse struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
}

func WriteResponse(ctx *fasthttp.RequestCtx, status int, data []byte) {
	resp := response{
		StatusCode: status,
		Data:       data,
	}

	WriteJson(ctx, status, resp)
}

func WriteErrorResponse(ctx *fasthttp.RequestCtx, status int, err string) {
	resp := errorResponse{
		StatusCode: status,
		Error:      err,
	}

	WriteJson(ctx, status, resp)
}

func WriteJson(ctx *fasthttp.RequestCtx, status int, object any) {
	prepared, err := json.Marshal(object)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(prepared)
}
