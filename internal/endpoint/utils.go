package endpoint

import (
	"encoding/json"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/objects/product"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"html/template"
)

var (
	productTmpl  *template.Template
	productsTmpl *template.Template
)

type response struct {
	StatusCode int    `json:"statusCode"`
	Data       string `json:"data"`
}

type errorResponse struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
}

func WriteResponse(ctx *fasthttp.RequestCtx, statusCode int, data any) {
	WriteJson(ctx, data)
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

func ProductHTMLResponse(ctx *fasthttp.RequestCtx, products product.Product) {
	err := productTmpl.Execute(ctx, products)
	if err != nil {
		logrus.Errorf("failed to execute template, error: %v", err)
		return
	}
}

func ProductsHTMLResponse(ctx *fasthttp.RequestCtx, products product.Products) {
	err := productsTmpl.Execute(ctx, products)
	if err != nil {
		logrus.Errorf("failed to execute template, error: %v", err)
		return
	}
}

func init() {
	var err error
	productTmpl, err = template.New("productInfo").Parse("Name: {{.Name}}\n " +
		"Category: {{.Category}} \n" +
		"Location: {{.Location}} \n" +
		"Color: {{.Color}} \n" +
		"Price: {{.Price}} \n" +
		"Amount: {{.Amount}}")

	if err != nil {
		logrus.Fatalf("failed to parse tmpl, error: %v", err)
	}

	productsTmpl, err = template.New("productsInfo").Parse("{{range .Product}} \n" +
		"Name: {{.Name}}\n " +
		"Category: {{.Category}} \n" +
		"Location: {{.Location}} \n" +
		"Color: {{.Color}} \n" +
		"Price: {{.Price}} \n" +
		"Amount: {{.Amount}}{{end}}")

	if err != nil {
		logrus.Fatalf("failed to parse tmpl, error: %v", err)
	}
}
