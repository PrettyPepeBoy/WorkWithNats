FROM golang:1.22-alpine as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build  -o ./main

FROM alpine:3
COPY --from=build ./app/configuration.yaml .
COPY --from=build ./app/main .
ENTRYPOINT ["./main"]
