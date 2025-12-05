FROM golang:1.25-alpine as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o apiserver main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/apiserver .
ENTRYPOINT ["/app/apiserver"]
