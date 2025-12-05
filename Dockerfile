FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o apiserver main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/apiserver .
ENTRYPOINT ["/app/apiserver"]
