# build stage
FROM golang:1.22-bookworm AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

# run unit tests
RUN go test -v -race ./...

# build the binary if tests pass
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /tiny-pushgateway ./...

# tiny runtime image
FROM gcr.io/distroless/static-debian12
COPY --from=builder /tiny-pushgateway /tiny-pushgateway
EXPOSE 9091
ENTRYPOINT ["/tiny-pushgateway"]
