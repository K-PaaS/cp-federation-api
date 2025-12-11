FROM golang:alpine AS builder
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY pkg ./pkg

RUN go mod download
RUN go build -o main ./cmd/api
WORKDIR /dist
RUN cp /build/main .

FROM alpine
COPY --from=builder /dist/main .
COPY cmd/api/app/intra/config.env /cmd/api/app/intra/config.env
COPY cmd/api/app/metrics/config.env /cmd/api/app/metrics/config.env
COPY cmd/api/app/localize /cmd/api/app/localize

RUN addgroup -S 1000 && adduser -S 1000 -G 1000
RUN mkdir -p /home/1000
RUN chown -R 1000:1000 /home/1000

WORKDIR /cmd/api
ENTRYPOINT ["/main"]