package endpoint

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/cache"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/objects/product"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"strconv"
	"time"
)

type metrics struct {
	devices  prometheus.Gauge
	counter  prometheus.Counter
	duration *prometheus.HistogramVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{

		devices: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "TestTaskNatsApp",
				Name:      "connected_devices",
				Help:      "number of connected devices",
			}),

		counter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "TestTaskNatsApp",
				Name:      "ping_request_count",
				Help:      "No of request handled by Ping handler",
			}),

		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "TestTaskNatsApp",
				Name:      "reply_time",
				Help:      "how long api reply takes time",
				Buckets:   []float64{0.1, 0.25, 0.5, 1},
			}, []string{"endpoint", "method", "status"}),
	}

	reg.MustRegister(m.devices, m.counter, m.duration)
	return m
}

type HttpHandler struct {
	productCache *cache.Cache[cache.Int, cache.ByteSlc]
	productTable *product.Table
	promHandler  fasthttp.RequestHandler

	metrics *metrics
}

func NewHttpHandler(productCache *cache.Cache[cache.Int, cache.ByteSlc], productTable *product.Table) *HttpHandler {
	reg := prometheus.NewRegistry()

	return &HttpHandler{
		productCache: productCache,
		productTable: productTable,
		promHandler:  fasthttpadaptor.NewFastHTTPHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})),

		metrics: newMetrics(reg),
	}
}

func (h *HttpHandler) Handle(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {

	case "/api/v1/product":
		h.metrics.counter.Inc()
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			h.getProduct(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}

	case "/api/v1/product/all":
		h.metrics.counter.Inc()
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			h.getAllProducts(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}

	case "/api/v1/user":
		h.metrics.counter.Inc()
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}

	case "/api/v1/cache":
		h.metrics.counter.Inc()
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			h.getCache(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}

	case "/api/v1/cache/dump":
		h.metrics.counter.Inc()
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			h.dumpCache(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}

	case "/api/v1/metrics":
		h.promHandler(ctx)

	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}

func (h *HttpHandler) getProduct(ctx *fasthttp.RequestCtx) {
	t := time.Now()

	var p product.Products
	p.Product = make([]product.Product, 1)

	id, err := ctx.QueryArgs().GetUint("id")
	if err != nil {
		WriteErrorResponse(ctx, fasthttp.StatusBadRequest, err.Error())
		h.metrics.durationReply("getProduct", fasthttp.MethodGet, strconv.Itoa(fasthttp.StatusBadRequest), t)
		return
	}

	cacheData, find := h.productCache.Get(cache.Int(id))
	if find {
		rawByte, err := cacheData.Marshal()
		if err != nil {
			logrus.Info("failed to cast data")
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			h.metrics.durationReply("getProduct", fasthttp.MethodGet, strconv.Itoa(fasthttp.StatusInternalServerError), t)
			return
		}

		err = json.Unmarshal(rawByte, &p.Product[0])
		if err != nil {
			logrus.Error("failed to unmarshal json", err)
			h.metrics.durationReply("getProduct", fasthttp.MethodGet, strconv.Itoa(fasthttp.StatusInternalServerError), t)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		ProductsHTMLResponse(ctx, p)
		h.metrics.durationReply("getProduct", fasthttp.MethodGet, strconv.Itoa(fasthttp.StatusOK), t)
		return
	}

	databaseData, err := h.productTable.GetById(id)
	if err != nil {
		if errors.Is(err, product.ErrRowNotExist) {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			h.metrics.durationReply("getProduct", fasthttp.MethodGet, strconv.Itoa(fasthttp.StatusNoContent), t)
		} else {
			logrus.Error("failed to get data from database, error: ", err)
			WriteErrorResponse(ctx, fasthttp.StatusInternalServerError, err.Error())
			h.metrics.durationReply("getProduct", fasthttp.MethodGet, strconv.Itoa(fasthttp.StatusInternalServerError), t)
		}
		return
	}

	err = json.Unmarshal(databaseData, &p.Product[0])
	if err != nil {
		logrus.Error("failed to unmarshal json", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		h.metrics.durationReply("getProduct", fasthttp.MethodGet, strconv.Itoa(fasthttp.StatusInternalServerError), t)
		return
	}

	ProductsHTMLResponse(ctx, p)
	h.metrics.durationReply("getProduct", fasthttp.MethodGet, strconv.Itoa(fasthttp.StatusOK), t)
	h.productCache.PutKey(cache.Int(id), databaseData)
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
	bufWriter := bufio.NewWriter(ctx)
	h.productCache.GetAllRawData(bufWriter)

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h *HttpHandler) dumpCache(ctx *fasthttp.RequestCtx) {
	ctx.SetBodyStreamWriter(h.productCache.GetAllRawData)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (m *metrics) durationReply(endpoint string, method string, status string, t time.Time) {
	m.duration.With(prometheus.Labels{"endpoint": endpoint, "method": method, "status": status}).Observe(time.Since(t).Seconds())
}
