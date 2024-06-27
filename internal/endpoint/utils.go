package endpoint

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

// todo correct answers
type response struct {
	StatusCode int    `json:"statusCode"`
	Data       []byte `json:"data"`
}

type errorResponse struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
}

func WriteResponse(ctx *fasthttp.RequestCtx, statusCode int, data []byte) {
	resp := response{
		StatusCode: statusCode,
		Data:       data,
	}

	WriteJson(ctx, resp)
}

func WriteErrorResponse(ctx *fasthttp.RequestCtx, statusCode int, err string) {
	resp := errorResponse{
		StatusCode: statusCode,
		Error:      err,
	}

	WriteJson(ctx, resp)
}

func WriteJson(ctx *fasthttp.RequestCtx, object any) {
	prepared, err := json.Marshal(object)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(prepared)
}
