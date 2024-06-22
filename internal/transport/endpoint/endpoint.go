package endpoint

import (
	"TestTaskNats/internal/cache"
	"TestTaskNats/internal/database/postgres"
	"TestTaskNats/internal/models"
	"TestTaskNats/internal/services/database/postgresservice"
	"TestTaskNats/internal/transport/Response"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type HttpHandler struct {
	Cache   *cache.Cache
	Storage *postgres.Storage
}

func NewInternalServicesForHttpHandlers(cch *cache.Cache, strg *postgres.Storage) HttpHandler {
	handlerServices := HttpHandler{
		Cache:   cch,
		Storage: strg,
	}
	return handlerServices
}

func (s *HttpHandler) CreateMux() func(ctx *fasthttp.RequestCtx) {
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/product":
			if string(ctx.Method()) == fasthttp.MethodGet {
				s.getProductFromDatabase(ctx)
			} else if string(ctx.Method()) == fasthttp.MethodPost { //create new method correctly
				s.checkTable(ctx)
			} else {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
			}
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}
	return m
}

func (s *HttpHandler) getProductFromDatabase(ctx *fasthttp.RequestCtx) {
	var id models.ProductID
	ctx.SetContentType("application/json")
	err := json.Unmarshal(ctx.PostBody(), &id)
	if err != nil {
		logrus.Error("[endpoint/getProduct] failed to unmarshal json from ctx.postBody, error ", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		_, err = ctx.Write(Response.ErrorResponse(fasthttp.StatusInternalServerError, err))
		if err != nil {
			logrus.Error("[endpoint/getProduct] failed to write response to context, error ", err)
		}
		return
	}
	data, find := s.Cache.ShowKey(id.ID)
	if find {
		ctx.SetStatusCode(fasthttp.StatusOK)
		_, err = ctx.Write(Response.OkResponse(fasthttp.StatusOK, data))
		if err != nil {
			logrus.Error("[endpoint/getProduct] failed to write response to context, error ", err)
		}
		return
	}
	productBody, err := postgresservice.GetDataFromTable(s.Storage, id.ID)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	if productBody == nil {
		ctx.SetStatusCode(fasthttp.StatusOK)
		_, err = ctx.Write(Response.OkResponse(fasthttp.StatusOK, []byte("product was not found in database")))
		if err != nil {
			logrus.Error("[endpoint/getProduct failed to write response to context, error", err)
		}
		return
	}
	s.Cache.PutKey(id.ID, productBody)
	ctx.SetStatusCode(fasthttp.StatusOK)
	_, err = ctx.Write(Response.OkResponse(fasthttp.StatusOK, productBody))
	if err != nil {
		logrus.Errorf("error occured, error: %v", err)
	}
}

func (s *HttpHandler) checkTable(ctx *fasthttp.RequestCtx) {
	data, err := postgresservice.CheckInTable(s.Storage)

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
