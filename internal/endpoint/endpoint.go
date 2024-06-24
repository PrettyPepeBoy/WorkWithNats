package endpoint

import (
	"errors"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/cache"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/product"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/services/database/postgresservice"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type HttpHandler struct {
	productCache *cache.Cache
	productTable *product.Table
}

func NewHttpHandler(productCache *cache.Cache, productTable *product.Table) *HttpHandler {
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

		case fasthttp.MethodPost:
			h.postProduct(ctx)

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

	data, find := h.productCache.Get(id)
	if find {
		WriteJson(ctx, fasthttp.StatusOK, data)
		return
	}

	data, err = h.productTable.GetById(id)
	if err != nil {
		if errors.Is(err, product.ErrRowNotExist) {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
		} else {
			logrus.Error("failed to get resp from database, error: ", err)
			WriteErrorResponse(ctx, fasthttp.StatusInternalServerError, err.Error())
		}
		return
	}

	h.productCache.PutKey(id, data)

	WriteJson(ctx, fasthttp.StatusOK, data)
}

// FIXME: DELETE ME
func (h *HttpHandler) postProduct(ctx *fasthttp.RequestCtx) {
	resp, err := h.productTable.CheckInTable()
	if err != nil {
		WriteErrorResponse(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	return resp, err

	data, err := postgresservice.CheckInTable(h.productTable)

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	for _, elem := range data {
		val, ok := elem.(string)
		if !ok {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}
		_, _ = ctx.Write([]byte(val))
	}
}
