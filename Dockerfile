# WulfVault Dockerfile
# Two stages: build a static Go binary, then ship it on an empty (scratch)
# base. The runtime image contains no OS packages, no shell and no package
# manager, only the binary, the web assets and the CA bundle.

# ---- build ----
FROM golang:1.27.4-alpine3.24 AS builder

RUN apk upgrade --no-cache && apk add --no-cache ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# modernc.org/sqlite is pure Go, so no cgo is needed. timetzdata embeds the
# time zone database because the runtime image has no /usr/share/zoneinfo.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -tags timetzdata \
        -ldflags="-s -w" -o /rootfs/app/wulfvault ./cmd/server

# Lay out the runtime file system and fail the build if the web assets are
# missing (they are served from disk, not embedded).
RUN mkdir -p /rootfs/app/web /rootfs/data /rootfs/uploads /rootfs/tmp /rootfs/etc/ssl/certs \
 && cp -r web/static /rootfs/app/web/static \
 && cp /etc/ssl/certs/ca-certificates.crt /rootfs/etc/ssl/certs/ \
 && chmod 1777 /rootfs/tmp \
 && test -s /rootfs/app/web/static/css/app-shell.css \
 && test -s /rootfs/app/web/static/css/style.css \
 && test -s /rootfs/app/web/static/js/dashboard.js \
 && test "$(find /rootfs/app/web/static -type f | wc -l)" -ge 8

# ---- runtime ----
FROM scratch

COPY --from=builder /rootfs/ /

WORKDIR /app

EXPOSE 8080

ENV DATA_DIR=/data \
    UPLOADS_DIR=/uploads \
    PORT=8080

CMD ["/app/wulfvault"]
