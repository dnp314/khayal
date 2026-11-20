# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/khayal ./cmd

FROM alpine:3.22
RUN apk add --no-cache tzdata ca-certificates \
 && adduser -D -u 10001 app
WORKDIR /app

COPY --from=build /out/khayal /app/khayal
COPY static /app/static

USER app
EXPOSE 8000
ENTRYPOINT ["/app/khayal"]
