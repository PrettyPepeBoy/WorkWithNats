package endpoint

import (
	"errors"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/cache"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/objects/product"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type HttpHandler struct {
	productCache []cache.Cache[int, []byte]
	productTable *product.Table
}

func NewHttpHandler(productCache []cache.Cache[int, []byte], productTable *product.Table) *HttpHandler {
	return &HttpHandler{
		productCache: productCache,
		productTable: productTable,
	}
}

func (h *HttpHandler) Handle(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {

	case "/api/v1/product":
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			h.getProduct(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}

	case "/api/v1/user":
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}

	case "/api/v1/cache":
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			h.getCache(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}

	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}

func (h *HttpHandler) getProduct(ctx *fasthttp.RequestCtx) {
	id, err := ctx.QueryArgs().GetUint("id")
	if err != nil {
		WriteErrorResponse(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	index := ProductHash(id)
	data, find := h.productCache[index].Get(id)
	if find {
		WriteResponse(ctx, fasthttp.StatusOK, data)
		return
	}

	databaseData, err := h.productTable.GetById(id)
	if err != nil {
		if errors.Is(err, product.ErrRowNotExist) {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
		} else {
			logrus.Error("failed to get data from database, error: ", err)
			WriteErrorResponse(ctx, fasthttp.StatusInternalServerError, err.Error())
		}
		return
	}

	h.productCache[index].PutKey(id, databaseData)

	ctx.SetBody(databaseData)
}

func (h *HttpHandler) getCache(ctx *fasthttp.RequestCtx) {
	data := make([]byte, 0, 2048)
	for i := 0; i < len(h.productCache); i++ {
		rawByte, err := h.productCache[i].GetAllRawData()
		if err != nil {
			logrus.Error("failed to marshal json, error:", err)
			WriteErrorResponse(ctx, fasthttp.StatusInternalServerError, err.Error())
		}

		data = append(data, rawByte...)
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(data)
}
