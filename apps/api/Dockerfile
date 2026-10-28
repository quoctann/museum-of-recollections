FROM golang:1.27.3-alpine3.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/storage-init ./cmd/storage-init

FROM alpine:3.23.4
RUN apk add --no-cache ca-certificates && addgroup -S -g 10001 api && adduser -S -D -H -u 10001 -G api api
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/storage-init /usr/local/bin/storage-init
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
