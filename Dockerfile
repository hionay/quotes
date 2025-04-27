FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
    -ldflags="-w -s" \
    -o ./quotes .

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /app/quotes /
COPY --from=builder /app/templates /templates
COPY --from=builder /app/static /static
USER nonroot:nonroot
ENTRYPOINT ["/quotes"]
