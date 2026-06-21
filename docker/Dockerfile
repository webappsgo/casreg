FROM casjaysdev/go:latest AS builder

ARG VERSION=devel
ARG COMMIT_ID=N/A
ARG BUILD_DATE=unknown

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w \
        -X 'main.Version=${VERSION}' \
        -X 'main.CommitID=${COMMIT_ID}' \
        -X 'main.BuildDate=${BUILD_DATE}'" \
    -o /app/binaries/casreg \
    ./src

FROM alpine:latest

LABEL org.opencontainers.image.title="casreg" \
      org.opencontainers.image.description="Self-hosted Docker Registry Platform" \
      org.opencontainers.image.vendor="webappsgo" \
      org.opencontainers.image.url="https://github.com/webappsgo/casreg"

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/binaries/casreg /usr/local/bin/casreg

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["/usr/local/bin/casreg"]
