package endpoint

import (
	"encoding/json"
	"errors"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/cache"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/objects/product"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type HttpHandler struct {
	productCache *cache.Cache[int, []byte]
	productTable *product.Table
}

func NewHttpHandler(productCache *cache.Cache[int, []byte], productTable *product.Table) *HttpHandler {
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

	case "/api/v1/product/all":
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			h.getAllProducts(ctx)
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
	var p product.Products
	p.Product = make([]product.Product, 1)

	id, err := ctx.QueryArgs().GetUint("id")
	if err != nil {
		WriteErrorResponse(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	cacheData, find := h.productCache.Get(id)
	if find {
		rawByte, ok := cacheData.([]byte)
		if !ok {
			logrus.Info("failed to cast data")
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(rawByte, &p.Product[0])
		if err != nil {
			logrus.Error("failed to unmarshal json", err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		ProductsHTMLResponse(ctx, p)
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

	err = json.Unmarshal(databaseData, &p.Product[0])
	if err != nil {
		logrus.Error("failed to unmarshal json", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ProductsHTMLResponse(ctx, p)
	h.productCache.PutKey(id, databaseData)
}

func (h *HttpHandler) getAllProducts(ctx *fasthttp.RequestCtx) {
	var Products product.Products
	var Product product.Product
	var id uint32
	rows, err := h.productTable.GetAllFromTable()
	if err != nil {
		logrus.Error("failed to get all products from table")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	for rows.Next() {
		err = rows.Scan(&id, &Product)
		if err != nil {
			return
		}
		Product.Id = id
		Products.Product = append(Products.Product, Product)
	}

	ProductsHTMLResponse(ctx, Products)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h *HttpHandler) getCache(ctx *fasthttp.RequestCtx) {
	data := make([]byte, 0, 2048)
	for i := 0; i < len(h.productCache.Buckets); i++ {
		rawByte, err := h.productCache.Buckets[i].GetAllRawData()
		if err != nil {
			logrus.Error("failed to marshal json, error:", err)
			WriteErrorResponse(ctx, fasthttp.StatusInternalServerError, err.Error())
		}

		data = append(data, rawByte...)
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(data)
}
