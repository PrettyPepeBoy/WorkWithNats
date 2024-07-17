FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ./main

FROM alpine:3
WORKDIR /newApp
COPY --from=builder ./app/main .
COPY --from=builder ./app/configuration.yaml .
COPY --from=builder ./app/product.html .
ENTRYPOINT ["./main"]