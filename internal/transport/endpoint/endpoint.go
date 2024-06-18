package endpoint

import (
	"TestTaskNats/internal/cache"
	"TestTaskNats/internal/database/postgres"
	"TestTaskNats/internal/services/cacheservice"
	"TestTaskNats/internal/services/database/postgresservice"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

func CreateMux(strg *postgres.Storage, c *cache.Cache) func(ctx *fasthttp.RequestCtx) { //create package endpoint
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/product":
			if string(ctx.Method()) == fasthttp.MethodGet {
				getProduct(ctx, c, strg)
			} else {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
			}
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}
	return m
}

func getProduct(ctx *fasthttp.RequestCtx, c *cache.Cache, strg *postgres.Storage) {
	find, err := cacheservice.FindInCache(c, ctx.PostBody())
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	if find {
		_, err = ctx.Write([]byte("find in cache"))
		if err != nil {
			logrus.Errorf("error occured, error: %v", err)
		}
		ctx.SetStatusCode(fasthttp.StatusOK)
		return
	}
	data, err := postgresservice.GetDataFromTable(strg, ctx.PostBody())
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	if data == nil {
		_, err = ctx.Write([]byte("product was not found in database"))
		ctx.SetStatusCode(fasthttp.StatusOK)
		return
	}
	err = cacheservice.PutInCache(c, data)
	if err != nil {
		logrus.Errorf("failed to add request body in cache")
	}
	_, err = ctx.Write(data)
	if err != nil {
		logrus.Errorf("error occured, error: %v", err)
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}
