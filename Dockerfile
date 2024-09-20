ARG GO_VERSION=1.22

FROM golang:${GO_VERSION} AS builder
WORKDIR /build
RUN go env -w GOCACHE=/go-cache \
    && go env -w GOMODCACHE=/gomod-cache

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache,id=gomod-cache \
    go mod download && go mod verify

COPY . ./

RUN --mount=type=cache,target=/gomod-cache,id=gomod-cache --mount=type=cache,target=/go-cache,id=go-cache \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags '-static'" -o /build/front ./cmd/front && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags '-static'" -o /build/store ./cmd/store && \
    chmod +x front && chmod +x store


FROM alpine:3.20 AS main
WORKDIR /
RUN apk update && apk add --no-cache tzdata curl

COPY --from=builder /build/front /front
COPY --from=builder /build/store /store